package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// RuntimeAgent deliberately keeps Eino types inside the agent package. Callers
// use AgentService instead of interacting with the agent or runner directly.
type RuntimeAgent struct {
	Runner *adk.Runner
}

type AgentFactory interface {
	Build(ctx context.Context, definition AgentDefinition, streaming bool) (*RuntimeAgent, error)
}

type EinoAgentFactory struct {
	models ModelResolver
	skills SkillRegistry
}

func NewEinoAgentFactory(models ModelResolver, skills SkillRegistry) (*EinoAgentFactory, error) {
	if models == nil || skills == nil {
		return nil, fmt.Errorf("%w: models and skills are required", ErrInvalidDefinition)
	}
	return &EinoAgentFactory{models: models, skills: skills}, nil
}

func (f *EinoAgentFactory) Build(ctx context.Context, definition AgentDefinition, streaming bool) (*RuntimeAgent, error) {
	chatModel, err := f.models.Resolve(ctx, definition.ModelProfile)
	if err != nil {
		return nil, err
	}
	maxOutputTokens := definition.MaxOutputTokens
	if maxOutputTokens == 0 {
		maxOutputTokens = defaultMaxOutputTokens
	}
	chatModel = newOutputLimitedModel(chatModel, maxOutputTokens)
	tools := make([]tool.BaseTool, 0, len(definition.SkillIDs))
	guard := newToolExecutionGuard(definition.MaxSteps)
	for _, skillID := range definition.SkillIDs {
		resolved, err := f.skills.Resolve(ctx, skillID)
		if err != nil {
			return nil, err
		}
		kind, err := f.skills.Kind(skillID)
		if err != nil {
			return nil, err
		}
		guarded, err := guard.wrap(resolved, kind)
		if err != nil {
			return nil, fmt.Errorf("guard skill %q: %w", skillID, err)
		}
		tools = append(tools, guarded)
	}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          definition.ID,
		Description:   definition.Description,
		Instruction:   definition.SystemPrompt,
		Model:         chatModel,
		MaxIterations: definition.MaxSteps,
		ToolsConfig: adk.ToolsConfig{ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools:               tools,
			ExecuteSequentially: true,
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("build Eino agent: %w", err)
	}
	return &RuntimeAgent{Runner: adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: streaming})}, nil
}
