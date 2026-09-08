package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
	"github.com/S-zhi/blender-shot-preview/internal/productiontools"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestAgentServiceInvokerProjectsInputsAndIdentity(t *testing.T) {
	service := &recordingAgentService{output: `{"ok":true}`}
	invoker := AgentServiceInvoker{Service: service}
	taskInput := snapshotForTest(t, map[string]any{
		"user_id": "user-1", "prompt": "make a forest", "conversation_id": "conversation-1",
	})

	_, err := invoker.InvokeAgent(context.Background(), AgentRequest{
		TaskID: "task-1", NodeID: NodeIntent, AgentID: "intent-agent", TaskInput: taskInput,
		Dependencies: DependencyOutputs{NodeInitialize: taskInput},
	})
	if err != nil {
		t.Fatalf("invoke intent agent: %v", err)
	}
	if service.last.UserID != "user-1" || service.last.TenantID != "user-1" || service.last.SessionID != "task-1" {
		t.Fatalf("unexpected agent identity: %+v", service.last)
	}
	assertJSONEqual(t, service.last.Input, `{"prompt":"make a forest"}`)

	_, err = invoker.InvokeAgent(context.Background(), AgentRequest{
		TaskID: "task-1", NodeID: NodeDesignShots, AgentID: "shot-designer-agent", TaskInput: taskInput,
		Dependencies: DependencyOutputs{
			NodeValidateSpec: snapshotForTest(t, map[string]any{"version": "scene-spec/v1", "scene_id": "scene-1"}),
			NodeCreateAssets: snapshotForTest(t, map[string]any{"asset_manifests": []any{map[string]any{"asset_id": "tree"}}}),
		},
	})
	if err != nil {
		t.Fatalf("invoke shot designer: %v", err)
	}
	var design map[string]any
	if err := json.Unmarshal([]byte(service.last.Input), &design); err != nil {
		t.Fatal(err)
	}
	if design["scene_spec"] == nil || design["asset_manifests"] == nil {
		t.Fatalf("shot designer input is incomplete: %s", service.last.Input)
	}

	_, err = invoker.InvokeAgent(context.Background(), AgentRequest{
		TaskID: "task-1", NodeID: NodeAssembleScene, AgentID: "scene-assembly-agent", TaskInput: taskInput,
		Dependencies: DependencyOutputs{
			NodeCreateAssets: snapshotForTest(t, map[string]any{"asset_manifests": []any{map[string]any{"asset_id": "tree"}}}),
			NodeDesignShots:  snapshotForTest(t, map[string]any{"shots": []any{map[string]any{"id": "shot-1"}}}),
		},
	})
	if err != nil {
		t.Fatalf("invoke scene assembly: %v", err)
	}
	var assembly map[string]any
	if err := json.Unmarshal([]byte(service.last.Input), &assembly); err != nil {
		t.Fatal(err)
	}
	if assembly["asset_manifests"] == nil || assembly["shot_plan"] == nil || assembly["output_path"] != "tasks/task-1/scene.blend" {
		t.Fatalf("scene assembly input is incomplete: %s", service.last.Input)
	}
}

func TestShotPreviewStepsRejectsInvalidSpec(t *testing.T) {
	_, err := (ShotPreviewSteps{}).InvokeStep(context.Background(), StepRequest{
		NodeID: NodeValidateSpec, Step: string(NodeValidateSpec),
		Dependencies: DependencyOutputs{NodeIntent: snapshotForTest(t, map[string]any{
			"version": "scene-spec/v2", "scene_id": "scene-1",
		})},
	})
	if !errors.Is(err, ErrInvalidSubmission) {
		t.Fatalf("expected invalid submission, got %v", err)
	}
}

func TestRegistryToolInvokerWaitsForRenderArtifact(t *testing.T) {
	registry := agent.NewMemorySkillRegistry()
	recorder := &scriptedTool{responses: map[string][]string{
		productiontools.ToolBlenderRenderSubmit:   {`{"ok":true,"data":{"job_id":"job-1"}}`},
		productiontools.ToolBlenderRenderStatus:   {`{"ok":true,"data":{"status":"running"}}`, `{"ok":true,"data":{"status":"succeeded"}}`},
		productiontools.ToolBlenderRenderArtifact: {`{"ok":true,"data":{"path":"preview/frame_0001.png"}}`},
	}}
	registerScriptedTools(t, registry, recorder,
		productiontools.ToolBlenderRenderSubmit,
		productiontools.ToolBlenderRenderStatus,
		productiontools.ToolBlenderRenderArtifact,
	)

	output, err := (RegistryToolInvoker{Skills: registry, PollInterval: time.Millisecond}).InvokeTool(context.Background(), ToolRequest{
		TaskID: "task-1", NodeID: NodePreviewRender, Tool: productiontools.ToolBlenderRenderSubmit,
		Dependencies: DependencyOutputs{NodeAssembleScene: snapshotForTest(t, map[string]any{"blend_file": "scene.blend"})},
	})
	if err != nil {
		t.Fatalf("invoke render: %v", err)
	}
	assertJSONEqual(t, string(output), `{"path":"preview/frame_0001.png"}`)
	if recorder.callCount(productiontools.ToolBlenderRenderStatus) != 2 {
		t.Fatalf("expected two status calls, got %d", recorder.callCount(productiontools.ToolBlenderRenderStatus))
	}
}

func TestRegistryToolInvokerCancelsRenderWithContext(t *testing.T) {
	registry := agent.NewMemorySkillRegistry()
	recorder := &scriptedTool{responses: map[string][]string{
		productiontools.ToolBlenderRenderSubmit: {`{"ok":true,"data":{"job_id":"job-1"}}`},
		productiontools.ToolBlenderRenderStatus: {`{"ok":true,"data":{"status":"running"}}`},
		productiontools.ToolBlenderRenderCancel: {`{"ok":true,"data":{"status":"cancelled"}}`},
	}}
	registerScriptedTools(t, registry, recorder,
		productiontools.ToolBlenderRenderSubmit,
		productiontools.ToolBlenderRenderStatus,
		productiontools.ToolBlenderRenderCancel,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (RegistryToolInvoker{Skills: registry, PollInterval: time.Hour}).InvokeTool(ctx, ToolRequest{
		TaskID: "task-1", NodeID: NodePreviewRender, Tool: productiontools.ToolBlenderRenderSubmit,
		Dependencies: DependencyOutputs{NodeAssembleScene: snapshotForTest(t, map[string]any{"blend_file": "scene.blend"})},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if recorder.callCount(productiontools.ToolBlenderRenderCancel) != 1 {
		t.Fatalf("expected render cancellation, calls: %+v", recorder.calls)
	}
}

func TestNewShotPreviewRunnerCreatesUsableBoundary(t *testing.T) {
	runner := NewShotPreviewRunner(NewMemoryRepository(), &recordingAgentService{output: `{}`}, agent.NewMemorySkillRegistry())
	if runner == nil || runner.repository == nil || runner.agents == nil || runner.steps == nil || runner.tools == nil {
		t.Fatal("shot preview runner dependencies were not assembled")
	}
}

type recordingAgentService struct {
	last   agent.AgentRequest
	output string
}

func (s *recordingAgentService) Run(_ context.Context, request agent.AgentRequest) (*agent.AgentResult, error) {
	s.last = request
	return &agent.AgentResult{Status: agent.RunStatusSucceeded, Output: s.output}, nil
}
func (*recordingAgentService) Stream(context.Context, agent.AgentRequest) (<-chan agent.AgentEvent, error) {
	return nil, errors.New("not implemented")
}
func (*recordingAgentService) Submit(context.Context, agent.AgentRequest) (string, error) {
	return "", errors.New("not implemented")
}
func (*recordingAgentService) GetRun(context.Context, string) (*agent.AgentRun, error) {
	return nil, errors.New("not implemented")
}
func (*recordingAgentService) Cancel(context.Context, string) error {
	return errors.New("not implemented")
}

type scriptedTool struct {
	mu        sync.Mutex
	responses map[string][]string
	calls     []string
}

func (t *scriptedTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "scripted", Desc: "scripted test tool"}, nil
}

func (t *scriptedTool) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	toolID, _ := ctx.Value(scriptedToolIDKey{}).(string)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, toolID)
	responses := t.responses[toolID]
	if len(responses) == 0 {
		return "", errors.New("unexpected tool call")
	}
	response := responses[0]
	if len(responses) > 1 {
		t.responses[toolID] = responses[1:]
	}
	return response, nil
}

func (t *scriptedTool) callCount(toolID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	count := 0
	for _, call := range t.calls {
		if call == toolID {
			count++
		}
	}
	return count
}

type scriptedToolIDKey struct{}

type identifiedTool struct {
	id   string
	base *scriptedTool
}

func (t identifiedTool) Info(ctx context.Context) (*schema.ToolInfo, error) { return t.base.Info(ctx) }
func (t identifiedTool) InvokableRun(ctx context.Context, arguments string, options ...tool.Option) (string, error) {
	return t.base.InvokableRun(context.WithValue(ctx, scriptedToolIDKey{}, t.id), arguments, options...)
}

func registerScriptedTools(t *testing.T, registry *agent.MemorySkillRegistry, scripted *scriptedTool, ids ...string) {
	t.Helper()
	for _, id := range ids {
		id := id
		if err := registry.Register(id, func(context.Context) (tool.BaseTool, error) {
			return identifiedTool{id: id, base: scripted}, nil
		}); err != nil {
			t.Fatalf("register tool %s: %v", id, err)
		}
	}
}

func snapshotForTest(t *testing.T, value any) Snapshot {
	t.Helper()
	snapshot, err := NewSnapshot(value)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func assertJSONEqual(t *testing.T, actual, expected string) {
	t.Helper()
	var actualValue, expectedValue any
	if err := json.Unmarshal([]byte(actual), &actualValue); err != nil {
		t.Fatalf("invalid actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &expectedValue); err != nil {
		t.Fatalf("invalid expected JSON: %v", err)
	}
	actualJSON, _ := json.Marshal(actualValue)
	expectedJSON, _ := json.Marshal(expectedValue)
	if string(actualJSON) != string(expectedJSON) {
		t.Fatalf("JSON mismatch: got %s want %s", actualJSON, expectedJSON)
	}
}
