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
	firstID, firstStatus, err := service.SubmitTask(context.Background(), "user-1", "A moonlit city shot", "conversation-1", "request-1")
	if err != nil || firstStatus != TaskStatusAccepted || firstID == "" {
		t.Fatalf("first SubmitTask() = id=%q status=%v err=%v", firstID, firstStatus, err)
	}
	secondID, secondStatus, err := service.SubmitTask(context.Background(), "user-1", "A moonlit city shot", "conversation-1", "request-1")
	if err != nil || secondStatus != TaskStatusAccepted || secondID != firstID {
		t.Fatalf("second SubmitTask() = id=%q status=%v err=%v", secondID, secondStatus, err)
	}
	task := waitForPipelineTask(t, runner, firstID)
	if task.Status != pipeline.TaskStatusSucceeded || len(task.Nodes) != 13 {
		t.Fatalf("pipeline task = %#v, want successful 13-stage task", task)
	}
}

func TestShotPreviewServiceRejectsIncompleteTask(t *testing.T) {
	service := NewShotPreviewService(nil)
	if _, status, err := service.SubmitTask(context.Background(), "", "shot", "", "request-1"); err != ErrRejected || status != TaskStatusRejected {
		t.Fatalf("empty user result = status=%v err=%v", status, err)
	}
	if _, status, err := service.SubmitTask(context.Background(), "user-1", "shot", "", ""); err != ErrRejected || status != TaskStatusRejected {
		t.Fatalf("empty request ID result = status=%v err=%v", status, err)
	}
}

func TestShotPreviewServiceRejectsUnavailablePipeline(t *testing.T) {
	service := NewShotPreviewService(nil)
	if _, status, err := service.SubmitTask(context.Background(), "user-1", "shot", "", "request-1"); err != ErrPipelineUnavailable || status != TaskStatusRejected {
		t.Fatalf("unavailable pipeline result = status=%v err=%v", status, err)
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
