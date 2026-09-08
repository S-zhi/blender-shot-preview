package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

func TestShotPreviewServiceSubmitsIdempotentAsyncPipeline(t *testing.T) {
	runner := pipeline.NewRunner(
		pipeline.NewMemoryRepository(),
		pipeline.AgentInvokerFunc(func(_ context.Context, request pipeline.AgentRequest) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"agent": request.AgentID})
		}),
		pipeline.StepInvokerFunc(func(_ context.Context, request pipeline.StepRequest) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"stage": request.Step})
		}),
		pipeline.ToolInvokerFunc(func(_ context.Context, request pipeline.ToolRequest) (json.RawMessage, error) {
			return json.Marshal(map[string]string{"tool": request.Tool})
		}),
	)
	service := NewShotPreviewService(runner)
	request := CreateTaskRequest{UserID: "user-1", Prompt: "A moonlit city shot", ConversationID: "conversation-1", RequestID: "request-1"}
	first, err := service.CreateTask(context.Background(), request)
	if err != nil || first.Status != CreateStatusAccepted || first.TaskID == "" || first.Replayed {
		t.Fatalf("first CreateTask() = %#v err=%v", first, err)
	}
	second, err := service.CreateTask(context.Background(), request)
	if err != nil || second.Status != CreateStatusAccepted || second.TaskID != first.TaskID || !second.Replayed {
		t.Fatalf("second CreateTask() = %#v err=%v", second, err)
	}
	task := waitForPipelineTask(t, runner, first.TaskID)
	if task.Status != pipeline.TaskStatusSucceeded || len(task.Nodes) != 13 {
		t.Fatalf("pipeline task = %#v, want successful 13-stage task", task)
	}
}

func TestShotPreviewServiceRejectsIncompleteTask(t *testing.T) {
	service := NewShotPreviewService(nil)
	if result, err := service.CreateTask(context.Background(), CreateTaskRequest{Prompt: "shot", RequestID: "request-1"}); err != ErrInvalidTaskRequest || result.Status != CreateStatusRejected {
		t.Fatalf("empty user result = result=%#v err=%v", result, err)
	}
	if result, err := service.CreateTask(context.Background(), CreateTaskRequest{UserID: "user-1", Prompt: "shot"}); err != ErrInvalidTaskRequest || result.Status != CreateStatusRejected {
		t.Fatalf("empty request ID result = result=%#v err=%v", result, err)
	}
	if result, err := service.CreateTask(context.Background(), CreateTaskRequest{UserID: "user-1", Prompt: "shot", RequestID: "request-1", WorkflowID: "arbitrary-dag"}); err != ErrInvalidTaskRequest || result.Status != CreateStatusRejected {
		t.Fatalf("unsupported workflow result = result=%#v err=%v", result, err)
	}
}

func TestShotPreviewServiceRejectsUnavailablePipeline(t *testing.T) {
	service := NewShotPreviewService(nil)
	if result, err := service.CreateTask(context.Background(), CreateTaskRequest{UserID: "user-1", Prompt: "shot", RequestID: "request-1"}); err != ErrPipelineUnavailable || result.Status != CreateStatusRejected {
		t.Fatalf("unavailable pipeline result = result=%#v err=%v", result, err)
	}
}

func waitForPipelineTask(t *testing.T, runner *pipeline.Runner, taskID string) pipeline.Task {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		task, err := runner.Get(context.Background(), taskID)
		if err != nil {
			t.Fatalf("pipeline.Get(%q): %v", taskID, err)
		}
		if task.Status.Terminal() {
			return task
		}
		time.Sleep(5 * time.Millisecond)
	}
	task, err := runner.Get(context.Background(), taskID)
	t.Fatalf("pipeline task %q did not complete, last=%#v err=%v", taskID, task, err)
	return pipeline.Task{}
}
