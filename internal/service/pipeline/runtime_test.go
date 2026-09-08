package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
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

type mockKeyUser struct {
	key llmgateway.UsableKey
	err error
}

func (m mockKeyUser) Use(_ context.Context, keyID, userID string) (llmgateway.UsableKey, error) {
	if m.err != nil {
		return llmgateway.UsableKey{}, m.err
	}
	return m.key, nil
}

func TestAgentServiceInvokerResolvesAndBindsKeys(t *testing.T) {
	service := &recordingAgentService{output: `{"ok":true}`}
	keys := mockKeyUser{
		key: llmgateway.UsableKey{
			KeyID:   "k-1",
			APIKey:  "sk-test-secret-123",
			BaseURL: "https://api.openai.com/v1",
		},
	}
	invoker := AgentServiceInvoker{Service: service, Keys: keys}
	taskInput := snapshotForTest(t, map[string]any{
		"user_id": "user-999", "prompt": "make a test scene",
	})

	_, err := invoker.InvokeAgent(context.Background(), AgentRequest{
		TaskID: "task-test", NodeID: NodeIntent, AgentID: "intent-agent", TaskInput: taskInput,
		Dependencies: DependencyOutputs{NodeInitialize: taskInput},
	})
	if err != nil {
		t.Fatalf("invoke agent: %v", err)
	}

	if service.last.Model == nil {
		t.Fatalf("expected request.Model to be populated from Keys, got nil")
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

func TestShotPreviewWorkflowEndToEnd(t *testing.T) {
	registry := agent.NewMemorySkillRegistry()
	scripted := &scriptedTool{responses: map[string][]string{
		productiontools.ToolBlenderRenderSubmit: {
			`{"ok":true,"data":{"job_id":"preview-job"}}`,
			`{"ok":true,"data":{"job_id":"final-job"}}`,
		},
		productiontools.ToolBlenderRenderStatus: {
			`{"ok":true,"data":{"status":"succeeded"}}`,
			`{"ok":true,"data":{"status":"succeeded"}}`,
		},
		productiontools.ToolBlenderRenderArtifact: {
			`{"ok":true,"data":{"artifact_path":"tasks/task-1/preview/frame_0001.png"}}`,
			`{"ok":true,"data":{"artifact_path":"tasks/task-1/final/frame_0001.png"}}`,
		},
		productiontools.ToolFFmpegEncode: {
			`{"ok":true,"data":{"output_path":"tasks/task-1/shot-preview.mp4","codec":"libx264","fps":24}}`,
		},
		productiontools.ToolFFprobeInspect: {
			`{"ok":true,"data":{"format":{"duration":"10.0"}}}`,
		},
	}}
	registerScriptedTools(t, registry, scripted,
		productiontools.ToolBlenderRenderSubmit,
		productiontools.ToolBlenderRenderStatus,
		productiontools.ToolBlenderRenderArtifact,
		productiontools.ToolFFmpegEncode,
		productiontools.ToolFFprobeInspect,
	)

	agentResponses := map[string]string{
		"intent-agent": `{
			"version": "scene-spec/v1",
			"scene_id": "scene-1",
			"summary": "forest establishing shot",
			"style": {"visual_style": "realistic", "mood": "peaceful"},
			"environment": {"location": "forest", "lighting": "daylight"},
			"assets": [],
			"actions": [],
			"shots": []
		}`,
		"scene-planner-agent": `{
			"version": "scene-plan/v1",
			"scene_id": "scene-1",
			"asset_tasks": [
				{
					"task_id": "task-tree",
					"asset_id": "tree",
					"asset_kind": "model",
					"description": "pine tree",
					"depends_on": [],
					"output_path": "assets/tree.blend"
				}
			],
			"environment_plan": {},
			"action_plan": []
		}`,
		"asset-creator-agent": `{
			"version": "asset-manifest/v1",
			"asset_id": "tree",
			"asset_kind": "model",
			"source_files": [],
			"blend_file": "assets/tree.blend",
			"inspection": {"passed": true}
		}`,
		"shot-designer-agent": `{
			"version": "shot-plan/v1",
			"scene_id": "scene-1",
			"shots": [{"id": "shot-1"}],
			"constraints": {}
		}`,
		"scene-assembly-agent": `{
			"version": "scene-assembly/v1",
			"scene_id": "scene-1",
			"blend_file": "tasks/task-1/scene.blend",
			"inspection": {"passed": true}
		}`,
	}

	mockAgents := &e2eAgentService{responses: agentResponses}
	runner := NewShotPreviewRunner(NewMemoryRepository(), mockAgents, registry)

	taskInput := snapshotForTest(t, map[string]any{
		"user_id":         "user-1",
		"prompt":          "Create a peaceful forest shot",
		"conversation_id": "conv-1",
		"request_id":      "req-1",
	})

	task, created, err := runner.Submit(context.Background(), Submission{
		IdempotencyKey: "e2e-workflow-test",
		Input:          taskInput,
		Workflow:       ShotPreviewWorkflow(taskInput.JSON),
	})
	if err != nil || !created {
		t.Fatalf("Submit() err=%v created=%t", err, created)
	}

	deadline := time.Now().Add(5 * time.Second)
	var finalTask Task
	for time.Now().Before(deadline) {
		current, err := runner.Get(context.Background(), task.ID)
		if err != nil {
			t.Fatalf("Get() err=%v", err)
		}
		if current.Status.Terminal() {
			finalTask = current
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if finalTask.Status != TaskStatusSucceeded {
		t.Fatalf("pipeline task status = %s, want %s (nodes: %+v)", finalTask.Status, TaskStatusSucceeded, finalTask.Nodes)
	}

	if len(finalTask.Nodes) != 13 {
		t.Fatalf("expected 13 nodes, got %d", len(finalTask.Nodes))
	}

	for _, node := range finalTask.Nodes {
		if node.Status != NodeStatusSucceeded {
			t.Fatalf("node %s status = %s, want %s", node.ID, node.Status, NodeStatusSucceeded)
		}
	}
}

type e2eAgentService struct {
	responses map[string]string
}

func (s *e2eAgentService) Run(_ context.Context, request agent.AgentRequest) (*agent.AgentResult, error) {
	resp, ok := s.responses[request.AgentID]
	if !ok {
		return nil, fmt.Errorf("unexpected agent: %s", request.AgentID)
	}
	return &agent.AgentResult{
		Status: agent.RunStatusSucceeded,
		Output: resp,
	}, nil
}
func (*e2eAgentService) Stream(context.Context, agent.AgentRequest) (<-chan agent.AgentEvent, error) {
	return nil, errors.New("not implemented")
}
func (*e2eAgentService) Submit(context.Context, agent.AgentRequest) (string, error) {
	return "", errors.New("not implemented")
}
func (*e2eAgentService) GetRun(context.Context, string) (*agent.AgentRun, error) {
	return nil, errors.New("not implemented")
}
func (*e2eAgentService) Cancel(context.Context, string) error {
	return errors.New("not implemented")
}

