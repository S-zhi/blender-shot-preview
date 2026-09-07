package agent

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/agent/skill"
	agenttool "github.com/S-zhi/blender-shot-preview/internal/agent/tool"
	"github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestDefinitionServiceCreatesValidatedAgent(t *testing.T) {
	t.Parallel()

	skills := newTestSkillRegistry(t)
	definitions := NewMemoryDefinitionRepository()
	service, err := NewDefinitionService(definitions, skills, NewStaticModelResolver(map[string]model.BaseChatModel{
		"general-chat": &fixedModel{response: "ok"},
	}))
	if err != nil {
		t.Fatalf("new definition service: %v", err)
	}

	definition, err := service.Create(context.Background(), CreateAgentRequest{
		ID: "order-support", Name: "Order Support", SystemPrompt: "Handle order questions.",
		ModelProfile: "general-chat", SkillIDs: []string{"echo", "echo"},
	})
	if err != nil {
		t.Fatalf("create definition: %v", err)
	}
	if definition.Status != AgentStatusEnabled {
		t.Fatalf("status = %q, want enabled", definition.Status)
	}
	if definition.MaxSteps != defaultMaxSteps || definition.MaxOutputTokens != defaultMaxOutputTokens || definition.Timeout != defaultTimeout {
		t.Fatalf("unexpected defaults: steps=%d tokens=%d timeout=%s", definition.MaxSteps, definition.MaxOutputTokens, definition.Timeout)
	}
	if len(definition.SkillIDs) != 1 || definition.SkillIDs[0] != "echo" {
		t.Fatalf("skill IDs = %#v", definition.SkillIDs)
	}

	if err := service.Disable(context.Background(), definition.ID); err != nil {
		t.Fatalf("disable definition: %v", err)
	}
	disabled, err := service.Get(context.Background(), definition.ID)
	if err != nil {
		t.Fatalf("get disabled definition: %v", err)
	}
	if disabled.Status != AgentStatusDisabled {
		t.Fatalf("status after disable = %q", disabled.Status)
	}
}

func TestDefinitionServiceRejectsMissingModelAndSkill(t *testing.T) {
	t.Parallel()

	service, err := NewDefinitionService(NewMemoryDefinitionRepository(), newTestSkillRegistry(t), NewStaticModelResolver(nil))
	if err != nil {
		t.Fatalf("new definition service: %v", err)
	}
	_, err = service.Create(context.Background(), CreateAgentRequest{
		ID: "invalid", Name: "Invalid", SystemPrompt: "Prompt", ModelProfile: "missing", SkillIDs: []string{"missing"},
	})
	if !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("create invalid definition error = %v, want ErrInvalidDefinition", err)
	}
}

func TestEinoAgentFactoryBuildsConfiguredSkillTools(t *testing.T) {
	t.Parallel()

	factory, err := NewEinoAgentFactory(
		NewStaticModelResolver(map[string]model.BaseChatModel{"general-chat": &fixedModel{response: "ok"}}),
		newTestSkillRegistry(t),
	)
	if err != nil {
		t.Fatalf("new agent factory: %v", err)
	}

	runtime, err := factory.Build(context.Background(), AgentDefinition{
		ID: "order-support", Name: "Order Support", Description: "Order support agent.",
		SystemPrompt: "Handle order questions.", ModelProfile: "general-chat", SkillIDs: []string{"echo"},
		MaxSteps: 4, MaxOutputTokens: 100, Timeout: time.Minute, Status: AgentStatusEnabled,
	}, false)
	if err != nil {
		t.Fatalf("build agent: %v", err)
	}
	if runtime == nil || runtime.Runner == nil {
		t.Fatalf("runtime agent = %#v", runtime)
	}
}

func TestAgentServiceRunRecordsSnapshotAndEvents(t *testing.T) {
	t.Parallel()

	service, definitions, runs := newTestAgentService(t, &fixedModel{response: "The order has shipped."})
	createEnabledDefinition(t, definitions)
	result, err := service.Run(context.Background(), AgentRequest{AgentID: "order-support", SessionID: "session-1", Input: "Where is order A123?", UserID: "user-1"})
	if err != nil {
		t.Fatalf("run agent: %v", err)
	}
	if result.Status != RunStatusSucceeded || result.Output != "The order has shipped." {
		t.Fatalf("unexpected result: %#v", result)
	}
	run, err := service.GetRun(context.Background(), result.RunID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status != RunStatusSucceeded || len(run.ConfigSnapshot) == 0 {
		t.Fatalf("unexpected persisted run: %#v", run)
	}
	events, err := runs.ListEvents(context.Background(), result.RunID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) < 3 || events[0].Type != EventAgentStarted || events[len(events)-1].Type != EventAgentComplete {
		t.Fatalf("unexpected events: %#v", events)
	}
}

func TestAgentServicePassesConfiguredOutputTokenLimit(t *testing.T) {
	t.Parallel()

	chatModel := &optionCaptureModel{}
	service, definitions, _ := newTestAgentService(t, chatModel)
	if err := definitions.Create(context.Background(), AgentDefinition{
		ID: "limited", Name: "Limited", SystemPrompt: "Answer briefly.", ModelProfile: "general-chat",
		MaxSteps: 12, MaxOutputTokens: 1234, Timeout: time.Minute, Status: AgentStatusEnabled,
	}); err != nil {
		t.Fatalf("create limited definition: %v", err)
	}
	if _, err := service.Run(context.Background(), AgentRequest{AgentID: "limited", Input: "hello"}); err != nil {
		t.Fatalf("run limited agent: %v", err)
	}
	if chatModel.maxTokens != 1234 {
		t.Fatalf("max tokens = %d, want 1234", chatModel.maxTokens)
	}
}

func TestAgentServiceBlocksRequestScopedSkills(t *testing.T) {
	t.Parallel()

	service, definitions, _ := newTestAgentService(t, &fixedModel{response: "ok"})
	createEnabledDefinition(t, definitions)
	_, err := service.Run(context.Background(), AgentRequest{AgentID: "order-support", Input: "hello", SkillIDs: []string{"echo"}})
	if !errors.Is(err, ErrDynamicSkillBlocked) {
		t.Fatalf("run with dynamic skill error = %v, want ErrDynamicSkillBlocked", err)
	}
}

func TestAgentServiceValidatesConfiguredInputAndOutputSchemas(t *testing.T) {
	t.Parallel()

	service, definitions, _ := newTestAgentService(t, &fixedModel{response: `{"answer":42}`})
	if err := definitions.Create(context.Background(), AgentDefinition{
		ID: "structured", Name: "Structured", SystemPrompt: "Return JSON.", ModelProfile: "general-chat",
		MaxSteps: 12, MaxOutputTokens: 4096, Timeout: time.Minute, Status: AgentStatusEnabled,
		InputSchema:  []byte(`{"type":"object","required":["question"],"properties":{"question":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","required":["answer"],"properties":{"answer":{"type":"string"}}}`),
	}); err != nil {
		t.Fatalf("create structured definition: %v", err)
	}
	if _, err := service.Run(context.Background(), AgentRequest{AgentID: "structured", Input: `{"question":12}`}); !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("invalid input error = %v, want ErrInvalidDefinition", err)
	}
	result, err := service.Run(context.Background(), AgentRequest{AgentID: "structured", Input: `{"question":"hello"}`})
	if !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("invalid output error = %v, want ErrInvalidDefinition", err)
	}
	if result == nil || result.Status != RunStatusFailed {
		t.Fatalf("invalid output result = %#v", result)
	}
}

func TestAgentServiceStreamAndCancel(t *testing.T) {
	t.Parallel()

	service, definitions, _ := newTestAgentService(t, &blockingModel{})
	createEnabledDefinition(t, definitions)
	events, err := service.Stream(context.Background(), AgentRequest{AgentID: "order-support", Input: "wait"})
	if err != nil {
		t.Fatalf("stream agent: %v", err)
	}
	started, ok := <-events
	if !ok || started.Type != EventAgentStarted {
		t.Fatalf("first stream event = %#v", started)
	}
	if err := service.Cancel(context.Background(), started.RunID); err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	var cancelled bool
	for event := range events {
		if event.Type == EventAgentCanceled {
			cancelled = true
		}
	}
	if !cancelled {
		t.Fatal("stream did not publish cancellation")
	}
	run, err := service.GetRun(context.Background(), started.RunID)
	if err != nil {
		t.Fatalf("get cancelled run: %v", err)
	}
	if run.Status != RunStatusCancelled {
		t.Fatalf("run status = %q, want cancelled", run.Status)
	}
}

func TestToolExecutionGuardRetriesReadWithStableIdempotencyKey(t *testing.T) {
	t.Parallel()

	base := &recordingTool{errors: []error{
		&ToolError{Code: "UNAVAILABLE", Retryable: true, Cause: errors.New("temporary")},
		nil,
	}}
	guarded, err := newToolExecutionGuard(12).wrap(base, ToolKindRead)
	if err != nil {
		t.Fatalf("wrap tool: %v", err)
	}
	invokable := guarded.(einotool.InvokableTool)
	result, err := invokable.InvokableRun(withRunID(context.Background(), "run-1"), `{"id":"A123"}`)
	if err != nil {
		t.Fatalf("invoke read tool: %v", err)
	}
	if result != "ok" || len(base.executions) != 2 {
		t.Fatalf("unexpected tool calls: result=%q executions=%#v", result, base.executions)
	}
	if base.executions[0].IdempotencyKey != base.executions[1].IdempotencyKey {
		t.Fatalf("retry changed idempotency key: %#v", base.executions)
	}
	if base.executions[0].Attempt != 1 || base.executions[1].Attempt != 2 {
		t.Fatalf("retry attempts = %#v", base.executions)
	}
}

func TestToolExecutionGuardDoesNotRetryWriteAndLimitsRepeats(t *testing.T) {
	t.Parallel()

	base := &recordingTool{errors: []error{&ToolError{Code: "UNAVAILABLE", Retryable: true, Cause: errors.New("temporary")}}}
	guarded, err := newToolExecutionGuard(12).wrap(base, ToolKindWrite)
	if err != nil {
		t.Fatalf("wrap tool: %v", err)
	}
	invokable := guarded.(einotool.InvokableTool)
	if _, err := invokable.InvokableRun(withRunID(context.Background(), "run-1"), `{}`); err == nil {
		t.Fatal("write tool unexpectedly succeeded")
	}
	if len(base.executions) != 1 {
		t.Fatalf("write tool attempts = %d, want 1", len(base.executions))
	}

	successful := &recordingTool{}
	guarded, err = newToolExecutionGuard(12).wrap(successful, ToolKindRead)
	if err != nil {
		t.Fatalf("wrap repeat tool: %v", err)
	}
	invokable = guarded.(einotool.InvokableTool)
	for range maximumCallsPerTool {
		if _, err := invokable.InvokableRun(withRunID(context.Background(), "run-2"), `{}`); err != nil {
			t.Fatalf("invoke repeated tool: %v", err)
		}
	}
	if _, err := invokable.InvokableRun(withRunID(context.Background(), "run-2"), `{}`); !errors.Is(err, ErrRepeatedToolLimit) {
		t.Fatalf("repeat limit error = %v, want ErrRepeatedToolLimit", err)
	}
}

func newTestAgentService(t *testing.T, chatModel model.BaseChatModel) (*Service, *MemoryDefinitionRepository, *MemoryRunRepository) {
	t.Helper()
	definitions := NewMemoryDefinitionRepository()
	runs := NewMemoryRunRepository()
	skills := newTestSkillRegistry(t)
	resolver := NewStaticModelResolver(map[string]model.BaseChatModel{"general-chat": chatModel})
	factory, err := NewEinoAgentFactory(resolver, skills)
	if err != nil {
		t.Fatalf("new factory: %v", err)
	}
	service, err := NewService(definitions, runs, factory)
	if err != nil {
		t.Fatalf("new agent service: %v", err)
	}
	return service, definitions, runs
}

func newTestSkillRegistry(t *testing.T) *skill.Registry {
	t.Helper()

	toolsRepository := agenttool.NewMemoryRepository()
	if err := toolsRepository.Create(context.Background(), agenttool.Definition{
		ID: "echo", Name: "Echo", Description: "Echo a value.", Kind: agenttool.KindRead,
	}); err != nil {
		t.Fatalf("create echo tool: %v", err)
	}
	tools, err := agenttool.NewRegistry(toolsRepository)
	if err != nil {
		t.Fatalf("new tool registry: %v", err)
	}
	if err := tools.Register("echo", agenttool.ProviderFunc(func(context.Context) (einotool.BaseTool, error) {
		return testTool{}, nil
	})); err != nil {
		t.Fatalf("register echo tool: %v", err)
	}

	skillsRepository := skill.NewMemoryRepository()
	if err := skillsRepository.Create(context.Background(), skill.Definition{
		ID: "echo", Name: "Echo", Description: "Echo a value.", ToolIDs: []string{"echo"},
	}); err != nil {
		t.Fatalf("create echo skill: %v", err)
	}
	skills, err := skill.NewRegistry(skillsRepository, tools)
	if err != nil {
		t.Fatalf("new skill registry: %v", err)
	}
	return skills
}

func createEnabledDefinition(t *testing.T, definitions DefinitionRepository) {
	t.Helper()
	if err := definitions.Create(context.Background(), AgentDefinition{
		ID: "order-support", Name: "Order Support", SystemPrompt: "Answer clearly.",
		ModelProfile: "general-chat", MaxSteps: 12, MaxOutputTokens: 4096, Timeout: time.Minute, Status: AgentStatusEnabled,
	}); err != nil {
		t.Fatalf("create test definition: %v", err)
	}
}

type fixedModel struct {
	response string
}

func (m *fixedModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	return schema.AssistantMessage(m.response, nil), nil
}

func (m *fixedModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(m.response, nil)}), nil
}

func (m *fixedModel) BindTools(_ []*schema.ToolInfo) error { return nil }

type optionCaptureModel struct {
	maxTokens int
}

func (m *optionCaptureModel) Generate(_ context.Context, _ []*schema.Message, options ...model.Option) (*schema.Message, error) {
	configured := model.GetCommonOptions(&model.Options{}, options...)
	if configured.MaxTokens != nil {
		m.maxTokens = *configured.MaxTokens
	}
	return schema.AssistantMessage("ok", nil), nil
}

func (m *optionCaptureModel) Stream(ctx context.Context, input []*schema.Message, options ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	message, err := m.Generate(ctx, input, options...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{message}), nil
}

func (m *optionCaptureModel) BindTools(_ []*schema.ToolInfo) error { return nil }

type blockingModel struct {
	once sync.Once
}

func (m *blockingModel) Generate(ctx context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (m *blockingModel) Stream(ctx context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (m *blockingModel) BindTools(_ []*schema.ToolInfo) error { return nil }

type testTool struct{}

func (testTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "echo", Desc: "Echo a value."}, nil
}

func (testTool) InvokableRun(_ context.Context, arguments string, _ ...einotool.Option) (string, error) {
	return arguments, nil
}

type recordingTool struct {
	mu         sync.Mutex
	errors     []error
	executions []ToolExecution
}

func (t *recordingTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "record", Desc: "Record calls."}, nil
}

func (t *recordingTool) InvokableRun(ctx context.Context, _ string, _ ...einotool.Option) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	execution, ok := ToolExecutionFromContext(ctx)
	if !ok {
		return "", errors.New("missing tool execution metadata")
	}
	t.executions = append(t.executions, execution)
	index := len(t.executions) - 1
	if index < len(t.errors) && t.errors[index] != nil {
		return "", t.errors[index]
	}
	return "ok", nil
}
