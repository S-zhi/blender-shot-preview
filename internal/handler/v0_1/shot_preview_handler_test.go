package v0_1

import (
	"context"
	"errors"
	"testing"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type stubShotPreviewService struct {
	taskID string
	status service.TaskStatus
	err    error
}

func (s *stubShotPreviewService) SubmitTask(context.Context, string, string, string, string) (string, service.TaskStatus, error) {
	return s.taskID, s.status, s.err
}

func newTestHandler(svc service.ShotPreviewService) *ShotPreviewHandler {
	return NewShotPreviewHandlerWithRequestIDGenerator(svc, func() (string, error) {
		return "request-123", nil
	})
}

func TestCreateShotPreviewTaskAccepted(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{taskID: "task-123", status: service.TaskStatusAccepted})

	response, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{
		UserId: "user-1",
		Prompt: "Create a wide establishing shot",
	})
	if err != nil {
		t.Fatalf("CreateShotPreviewTask() error = %v", err)
	}
	if response.TaskId != "task-123" || response.Status != api.TaskStatus_ACCEPTED {
		t.Fatalf("CreateShotPreviewTask() response = %+v", response)
	}
	if response.GetRequestId() != "request-123" {
		t.Fatalf("request_id = %q, want request-123", response.GetRequestId())
	}
}

func TestCreateShotPreviewTaskRejectsEmptyUserID(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{})

	_, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{Prompt: "shot"})
	var invalid *InvalidArgumentError
	if !errors.As(err, &invalid) || invalid.Field != "user_id" {
		t.Fatalf("error = %v, want user_id InvalidArgumentError", err)
	}
}

func TestCreateShotPreviewTaskRejectsEmptyPrompt(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{})

	_, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{UserId: "user-1"})
	var invalid *InvalidArgumentError
	if !errors.As(err, &invalid) || invalid.Field != "prompt" {
		t.Fatalf("error = %v, want prompt InvalidArgumentError", err)
	}
}

func TestCreateShotPreviewTaskMapsServiceError(t *testing.T) {
	cause := errors.New("service unavailable")
	handler := newTestHandler(&stubShotPreviewService{err: cause})

	_, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{
		UserId: "user-1",
		Prompt: "shot",
	})
	var internal *InternalError
	if !errors.As(err, &internal) {
		t.Fatalf("error = %v, want InternalError", err)
	}
	if internal.RequestID != "request-123" || !errors.Is(err, cause) {
		t.Fatalf("internal error = %+v", internal)
	}
}

func TestCreateShotPreviewTaskReturnsBusinessRejection(t *testing.T) {
	handler := newTestHandler(&stubShotPreviewService{err: service.ErrRejected})

	response, err := handler.CreateShotPreviewTask(context.Background(), &api.CreateShotPreviewTaskRequest{
		UserId: "user-1",
		Prompt: "shot",
	})
	if err != nil {
		t.Fatalf("CreateShotPreviewTask() error = %v", err)
	}
	if response.Status != api.TaskStatus_REJECTED {
		t.Fatalf("status = %s, want REJECTED", response.Status)
	}
}
