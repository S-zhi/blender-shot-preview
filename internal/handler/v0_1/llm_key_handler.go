package v0_1

import (
	"context"
	"errors"
	"strings"

	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

type LLMKeyHandler struct {
	service service.LLMKeyService
}

func NewLLMKeyHandler(svc service.LLMKeyService) *LLMKeyHandler {
	return &LLMKeyHandler{service: svc}
}

func (h *LLMKeyHandler) ManageLLMKey(ctx context.Context, request *api.ManageLLMKeyRequest) (*api.ManageLLMKeyResponse, error) {
	if request == nil {
		return nil, kerrors.NewBizStatusError(400, "request is required")
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, kerrors.NewBizStatusError(400, "user_id is required")
	}

	command, err := mapLLMKeyCommand(request, userID)
	if err != nil {
		return nil, err
	}
	if h == nil || h.service == nil {
		return nil, kerrors.NewBizStatusError(500, "LLM key service unavailable")
	}
	result, err := h.service.ManageKey(ctx, command)
	if errors.Is(err, llmgateway.ErrInvalidCommand) {
		return nil, kerrors.NewBizStatusError(400, "invalid credential command")
	}
	if errors.Is(err, llmgateway.ErrCredentialForbidden) {
		return nil, kerrors.NewBizStatusError(403, "credential access forbidden")
	}
	if errors.Is(err, llmgateway.ErrCredentialNotFound) {
		return nil, kerrors.NewBizStatusError(404, "credential not found")
	}
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, "credential management failed")
	}

	response := &api.ManageLLMKeyResponse{KeyId: result.KeyID}
	if request.IsSetProvider() {
		provider := request.GetProvider()
		response.Provider = &provider
	}
	return response, nil
}

func mapLLMKeyCommand(request *api.ManageLLMKeyRequest, userID string) (llmgateway.KeyCommand, error) {
	command := llmgateway.KeyCommand{UserID: userID}
	switch request.GetOperation() {
	case api.LLMKeyOperation_SAVE:
		command.Type = llmgateway.KeyCommandSave
	case api.LLMKeyOperation_UPDATE:
		command.Type = llmgateway.KeyCommandUpdate
	case api.LLMKeyOperation_DELETE:
		command.Type = llmgateway.KeyCommandDelete
	default:
		return command, kerrors.NewBizStatusError(400, "unsupported operation")
	}

	if request.IsSetKeyId() {
		command.KeyID = strings.TrimSpace(request.GetKeyId())
	}
	if request.IsSetProvider() {
		provider, err := mapProvider(request.GetProvider())
		if err != nil {
			return command, err
		}
		command.Provider = provider
	}
	if request.IsSetName() {
		value := request.GetName()
		command.Name = &value
	}
	if request.IsSetApiKey() {
		value := request.GetApiKey()
		command.APIKey = &value
	}
	if request.IsSetBaseUrl() {
		value := request.GetBaseUrl()
		command.BaseURL = &value
	}
	return command, nil
}

func mapProvider(provider api.LLMProvider) (llmgateway.Provider, error) {
	switch provider {
	case api.LLMProvider_OPENAI:
		return llmgateway.ProviderOpenAI, nil
	case api.LLMProvider_ANTHROPIC:
		return llmgateway.ProviderAnthropic, nil
	default:
		return "", kerrors.NewBizStatusError(400, "unsupported provider")
	}
}
