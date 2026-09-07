package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ModelResolver is the only model dependency of AgentFactory. Its production
// implementation belongs to the separate LLM gateway integration.
type ModelResolver interface {
	Exists(ctx context.Context, profile string) bool
	Resolve(ctx context.Context, profile string) (model.BaseChatModel, error)
}

type StaticModelResolver struct {
	models map[string]model.BaseChatModel
}

func NewStaticModelResolver(models map[string]model.BaseChatModel) *StaticModelResolver {
	copyOfModels := make(map[string]model.BaseChatModel, len(models))
	for profile, chatModel := range models {
		copyOfModels[profile] = chatModel
	}
	return &StaticModelResolver{models: copyOfModels}
}

func (r *StaticModelResolver) Exists(_ context.Context, profile string) bool {
	_, exists := r.models[profile]
	return exists
}

func (r *StaticModelResolver) Resolve(_ context.Context, profile string) (model.BaseChatModel, error) {
	chatModel, exists := r.models[profile]
	if !exists {
		return nil, fmt.Errorf("%w: model profile %q", ErrInvalidDefinition, profile)
	}
	return chatModel, nil
}

// outputLimitedModel applies the AgentDefinition output limit to every Eino
// model request while preserving standard tool-binding behavior.
type outputLimitedModel struct {
	base      model.BaseChatModel
	maxTokens int
}

func newOutputLimitedModel(base model.BaseChatModel, maxTokens int) model.BaseChatModel {
	return &outputLimitedModel{base: base, maxTokens: maxTokens}
}

func (m *outputLimitedModel) Generate(ctx context.Context, input []*schema.Message, options ...model.Option) (*schema.Message, error) {
	return m.base.Generate(ctx, input, append(options, model.WithMaxTokens(m.maxTokens))...)
}

func (m *outputLimitedModel) Stream(ctx context.Context, input []*schema.Message, options ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.base.Stream(ctx, input, append(options, model.WithMaxTokens(m.maxTokens))...)
}

func (m *outputLimitedModel) BindTools(tools []*schema.ToolInfo) error {
	if chatModel, ok := m.base.(model.ChatModel); ok {
		return chatModel.BindTools(tools)
	}
	return fmt.Errorf("model does not support tool binding")
}

func (m *outputLimitedModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	if toolCalling, ok := m.base.(model.ToolCallingChatModel); ok {
		bound, err := toolCalling.WithTools(tools)
		if err != nil {
			return nil, err
		}
		return &outputLimitedModel{base: bound, maxTokens: m.maxTokens}, nil
	}
	if err := m.BindTools(tools); err != nil {
		return nil, err
	}
	return m, nil
}
