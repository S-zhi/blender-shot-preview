package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// OpenAICompatibleChatModel implements model.BaseChatModel by calling OpenAI-compatible
// chat completions HTTP endpoints (OpenAI, DeepSeek, Qwen, local vLLM/Ollama, etc.).
type OpenAICompatibleChatModel struct {
	baseURL    string
	apiKey     string
	headers    map[string]string
	modelName  string
	httpClient *http.Client
}

// OpenAIModelConfig provides configuration for creating OpenAICompatibleChatModel.
type OpenAIModelConfig struct {
	BaseURL    string
	APIKey     string
	Headers    map[string]string
	ModelName  string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewOpenAICompatibleChatModel constructs a new OpenAICompatibleChatModel.
func NewOpenAICompatibleChatModel(cfg OpenAIModelConfig) *OpenAICompatibleChatModel {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	modelName := strings.TrimSpace(cfg.ModelName)
	if modelName == "" {
		modelName = "gpt-4o"
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout == 0 {
			timeout = 60 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	headers := make(map[string]string, len(cfg.Headers))
	for k, v := range cfg.Headers {
		headers[k] = v
	}
	return &OpenAICompatibleChatModel{
		baseURL:    baseURL,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		headers:    headers,
		modelName:  modelName,
		httpClient: httpClient,
	}
}

// NewChatModelFromKey builds an OpenAICompatibleChatModel directly from a decrypted UsableKey.
func NewChatModelFromKey(key llmgateway.UsableKey, modelName string) model.BaseChatModel {
	return NewOpenAICompatibleChatModel(OpenAIModelConfig{
		BaseURL:   key.BaseURL,
		APIKey:    key.APIKey,
		Headers:   key.Headers,
		ModelName: modelName,
	})
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model     string              `json:"model"`
	Messages  []openAIChatMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (m *OpenAICompatibleChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	options := model.GetCommonOptions(&model.Options{}, opts...)

	messages := make([]openAIChatMessage, 0, len(input))
	for _, msg := range input {
		role := string(msg.Role)
		if role == "" {
			role = "user"
		}
		messages = append(messages, openAIChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	reqBody := openAIChatRequest{
		Model:    m.modelName,
		Messages: messages,
	}
	if options != nil && options.MaxTokens != nil && *options.MaxTokens > 0 {
		reqBody.MaxTokens = *options.MaxTokens
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal openai chat request: %w", err)
	}

	url := m.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create openai http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if m.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)
	}
	for k, v := range m.headers {
		httpReq.Header.Set(k, v)
	}

	httpResp, err := m.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, &LLMError{Code: LLMTimeout, Cause: errors.New("model request timed out")}
		}
		return nil, &LLMError{Code: LLMConnectionFailed, Cause: errors.New("model endpoint is unreachable")}
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &LLMError{Code: LLMConnectionFailed, Cause: errors.New("failed to read model response")}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		code := LLMProviderError
		switch httpResp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			code = LLMAuthFailed
		case http.StatusTooManyRequests:
			code = LLMRateLimited
		}
		return nil, &LLMError{Code: code, Cause: fmt.Errorf("model provider returned HTTP %d", httpResp.StatusCode)}
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, &LLMError{Code: LLMInvalidResponse, Cause: errors.New("model response is not valid JSON")}
	}

	if chatResp.Error != nil {
		return nil, &LLMError{Code: LLMProviderError, Cause: errors.New("model provider returned an error")}
	}

	if len(chatResp.Choices) == 0 {
		return nil, &LLMError{Code: LLMInvalidResponse, Cause: errors.New("model response contained no choices")}
	}

	choice := chatResp.Choices[0]
	return &schema.Message{
		Role:    schema.RoleType(choice.Message.Role),
		Content: choice.Message.Content,
	}, nil
}

func (m *OpenAICompatibleChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func (m *OpenAICompatibleChatModel) BindTools(_ []*schema.ToolInfo) error {
	return nil
}
