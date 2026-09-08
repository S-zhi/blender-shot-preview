package v0_1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type stubShotPreviewService struct {
	createResult service.CreateTaskResult
	createErr    error
	task         service.TaskView
	err          error
}

func (s *stubShotPreviewService) CreateTask(context.Context, service.CreateTaskRequest) (service.CreateTaskResult, error) {
	return s.createResult, s.createErr
}
func (s *stubShotPreviewService) GetTask(context.Context, service.GetTaskRequest) (service.TaskView, error) {
	return s.task, s.err
}
func (s *stubShotPreviewService) CancelTask(context.Context, service.CancelTaskRequest) (service.TaskView, error) {
	return s.task, s.err
}
func (s *stubShotPreviewService) RetryTask(context.Context, service.RetryTaskRequest) (service.TaskView, error) {
	return s.task, s.err
}
func (s *stubShotPreviewService) ConfirmStep(context.Context, service.ConfirmStepRequest) error {
	return s.err
}
func (s *stubShotPreviewService) AdjustStep(context.Context, service.AdjustStepRequest) error {
	return s.err
}
func (s *stubShotPreviewService) SubscribeEvents(context.Context, string) (<-chan pipeline.PipelineEvent, func(), error) {
	ch := make(chan pipeline.PipelineEvent)
	return ch, func() { close(ch) }, s.err
}

func newTestHandler(svc service.ShotPreviewService) *ShotPreviewHandler {
	return NewShotPreviewHandlerWithRequestIDGenerator(svc, func() (string, error) { return "request-123", nil })
}

func TestCreateShotPreviewTaskAcceptedAndReplayed(t *testing.T) {
	serviceStub := &stubShotPreviewService{createResult: service.CreateTaskResult{
		TaskID: "task-123", RequestID: "request-123", Status: service.CreateStatusAccepted, Replayed: true,
	}}
	handler := newTestHandler(serviceStub)
	response, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{UserId: "user-1", Prompt: "shot"})
	if err != nil || response.TaskId != "task-123" || response.Status != api.TaskStatus_ACCEPTED || !response.GetReplayed() {
		t.Fatalf("response=%+v err=%v", response, err)
	}
}

func TestCreateShotPreviewTaskPassesRegisteredWorkflow(t *testing.T) {
	serviceStub := &capturingShotPreviewService{stubShotPreviewService: stubShotPreviewService{
		createResult: service.CreateTaskResult{TaskID: "task-123", RequestID: "request-123", Status: service.CreateStatusAccepted},
	}}
	_, err := newTestHandler(serviceStub).CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{
		UserId: "user-1", Prompt: "shot", WorkflowId: stringPointer(service.ShotPreviewWorkflowID), WorkflowVersion: stringPointer(service.ShotPreviewWorkflowVersion),
	})
	if err != nil || serviceStub.createRequest.WorkflowID != service.ShotPreviewWorkflowID || serviceStub.createRequest.WorkflowVersion != service.ShotPreviewWorkflowVersion {
		t.Fatalf("request=%+v err=%v", serviceStub.createRequest, err)
	}
}

func TestCreateShotPreviewTaskRejectsEmptyUserID(t *testing.T) {
	_, err := newTestHandler(&stubShotPreviewService{}).CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{Prompt: "shot"})
	var invalid *InvalidArgumentError
	if !errors.As(err, &invalid) || invalid.Field != "user_id" {
		t.Fatalf("error=%v, want user_id InvalidArgumentError", err)
	}
}

func TestGetShotPreviewTaskMapsTaskView(t *testing.T) {
	now := time.Date(2026, 9, 8, 1, 2, 3, 0, time.UTC)
	handler := newTestHandler(&stubShotPreviewService{task: service.TaskView{
		TaskID: "task-123", Status: service.TaskStatusRunning, WorkflowID: service.ShotPreviewWorkflowID,
		WorkflowVersion: service.ShotPreviewWorkflowVersion, CreatedAt: now, UpdatedAt: now,
		Nodes: []service.NodeView{{ID: "Intent", Status: service.NodeStatusSucceeded, Attempts: 1}},
	}})
	response, err := handler.GetShotPreviewTask(context.Background(), &api.GetShotPreviewTaskRequest{UserId: "user-1", TaskId: "task-123"})
	if err != nil || response.Task.GetTaskId() != "task-123" || response.Task.GetStatus() != api.ProductionTaskStatus_RUNNING || response.Task.Nodes[0].Status != api.ProductionTaskStatus_SUCCEEDED {
		t.Fatalf("response=%+v err=%v", response, err)
	}
}

func TestTaskOperationsMapNotFoundAndConflict(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{err: service.ErrTaskNotFound})
	_, err := handler.GetShotPreviewTask(context.Background(), &api.GetShotPreviewTaskRequest{UserId: "user-1", TaskId: "missing"})
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("get error=%v, want NotFoundError", err)
	}

	handler = newTestHandler(&stubShotPreviewService{err: service.ErrTaskNotRetryable})
	_, err = handler.RetryShotPreviewTask(context.Background(), &api.RetryShotPreviewTaskRequest{UserId: "user-1", TaskId: "task-123"})
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("retry error=%v, want ConflictError", err)
	}
}

func TestTaskOperationsValidateIDs(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{})
	_, err := handler.CancelShotPreviewTask(context.Background(), &api.CancelShotPreviewTaskRequest{UserId: "user-1"})
	var invalid *InvalidArgumentError
	if !errors.As(err, &invalid) || invalid.Field != "task_id" {
		t.Fatalf("error=%v, want task_id InvalidArgumentError", err)
	}
}

type capturingShotPreviewService struct {
	stubShotPreviewService
	createRequest service.CreateTaskRequest
}

func (s *capturingShotPreviewService) CreateTask(ctx context.Context, request service.CreateTaskRequest) (service.CreateTaskResult, error) {
	s.createRequest = request
	return s.stubShotPreviewService.CreateTask(ctx, request)
}

func stringPointer(value string) *string { return &value }
