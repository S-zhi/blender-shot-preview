package v0_1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type InvalidArgumentError struct{ Field string }

func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("invalid argument: %s is required", e.Field)
}

type NotFoundError struct{}

func (*NotFoundError) Error() string { return "shot preview task not found" }

type ConflictError struct{ Message string }

func (e *ConflictError) Error() string { return e.Message }

type InternalError struct {
	RequestID string
	Cause     error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("internal error (request_id=%s)", e.RequestID)
}
func (e *InternalError) Unwrap() error { return e.Cause }

type RequestIDGenerator func() (string, error)

type ShotPreviewHandler struct {
	service      service.ShotPreviewService
	newRequestID RequestIDGenerator
}

func NewShotPreviewHandler(svc service.ShotPreviewService) *ShotPreviewHandler {
	return &ShotPreviewHandler{service: svc, newRequestID: generateRequestID}
}

func NewShotPreviewHandlerWithRequestIDGenerator(svc service.ShotPreviewService, generator RequestIDGenerator) *ShotPreviewHandler {
	return &ShotPreviewHandler{service: svc, newRequestID: generator}
}

func (h *ShotPreviewHandler) CreateShotPreviewTask(ctx context.Context, request *api.CreateShotPreviewTaskRequest) (*api.CreateShotPreviewTaskResponse, error) {
	if request == nil {
		return nil, &InvalidArgumentError{Field: "request"}
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, &InvalidArgumentError{Field: "user_id"}
	}
	prompt := strings.TrimSpace(request.GetPrompt())
	if prompt == "" {
		return nil, &InvalidArgumentError{Field: "prompt"}
	}
	requestID := strings.TrimSpace(request.GetRequestId())
	if requestID == "" {
		var err error
		requestID, err = h.newRequestID()
		if err != nil {
			return nil, &InternalError{RequestID: "unavailable", Cause: err}
		}
	}
	result, err := h.service.CreateTask(ctx, service.CreateTaskRequest{
		UserID: userID, Prompt: prompt, ConversationID: strings.TrimSpace(request.GetConversationId()), RequestID: requestID,
		WorkflowID: strings.TrimSpace(request.GetWorkflowId()), WorkflowVersion: strings.TrimSpace(request.GetWorkflowVersion()),
	})
	if errors.Is(err, service.ErrInvalidTaskRequest) {
		return &api.CreateShotPreviewTaskResponse{TaskId: result.TaskID, Status: api.TaskStatus_REJECTED, RequestId: &requestID}, nil
	}
	if err != nil {
		return nil, h.mapError(requestID, err)
	}
	replayed := result.Replayed
	return &api.CreateShotPreviewTaskResponse{
		TaskId: result.TaskID, Status: mapCreateStatus(result.Status), RequestId: &result.RequestID, Replayed: &replayed,
	}, nil
}

func (h *ShotPreviewHandler) GetShotPreviewTask(ctx context.Context, request *api.GetShotPreviewTaskRequest) (*api.GetShotPreviewTaskResponse, error) {
	userID, taskID, err := getInput(request)
	if err != nil {
		return nil, err
	}
	task, err := h.service.GetTask(ctx, service.GetTaskRequest{UserID: userID, TaskID: taskID})
	if err != nil {
		return nil, h.mapError("unavailable", err)
	}
	return &api.GetShotPreviewTaskResponse{Task: toAPITaskView(task)}, nil
}

func (h *ShotPreviewHandler) CancelShotPreviewTask(ctx context.Context, request *api.CancelShotPreviewTaskRequest) (*api.CancelShotPreviewTaskResponse, error) {
	userID, taskID, err := cancelInput(request)
	if err != nil {
		return nil, err
	}
	task, err := h.service.CancelTask(ctx, service.CancelTaskRequest{UserID: userID, TaskID: taskID})
	if err != nil {
		return nil, h.mapError("unavailable", err)
	}
	return &api.CancelShotPreviewTaskResponse{Task: toAPITaskView(task)}, nil
}

func (h *ShotPreviewHandler) RetryShotPreviewTask(ctx context.Context, request *api.RetryShotPreviewTaskRequest) (*api.RetryShotPreviewTaskResponse, error) {
	userID, taskID, err := retryInput(request)
	if err != nil {
		return nil, err
	}
	task, err := h.service.RetryTask(ctx, service.RetryTaskRequest{UserID: userID, TaskID: taskID})
	if err != nil {
		return nil, h.mapError("unavailable", err)
	}
	return &api.RetryShotPreviewTaskResponse{Task: toAPITaskView(task)}, nil
}

func getInput(request *api.GetShotPreviewTaskRequest) (string, string, error) {
	if request == nil {
		return "", "", &InvalidArgumentError{Field: "request"}
	}
	return taskInput(request.GetUserId(), request.GetTaskId())
}

func cancelInput(request *api.CancelShotPreviewTaskRequest) (string, string, error) {
	if request == nil {
		return "", "", &InvalidArgumentError{Field: "request"}
	}
	return taskInput(request.GetUserId(), request.GetTaskId())
}

func retryInput(request *api.RetryShotPreviewTaskRequest) (string, string, error) {
	if request == nil {
		return "", "", &InvalidArgumentError{Field: "request"}
	}
	return taskInput(request.GetUserId(), request.GetTaskId())
}

func taskInput(rawUserID, rawTaskID string) (string, string, error) {
	userID := strings.TrimSpace(rawUserID)
	if userID == "" {
		return "", "", &InvalidArgumentError{Field: "user_id"}
	}
	taskID := strings.TrimSpace(rawTaskID)
	if taskID == "" {
		return "", "", &InvalidArgumentError{Field: "task_id"}
	}
	return userID, taskID, nil
}

func (h *ShotPreviewHandler) mapError(requestID string, err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidTaskRequest):
		return &InvalidArgumentError{Field: "request"}
	case errors.Is(err, service.ErrTaskNotFound):
		return &NotFoundError{}
	case errors.Is(err, service.ErrTaskNotCancellable), errors.Is(err, service.ErrTaskNotRetryable):
		return &ConflictError{Message: err.Error()}
	default:
		return &InternalError{RequestID: requestID, Cause: err}
	}
}

func mapCreateStatus(status service.CreateStatus) api.TaskStatus {
	if status == service.CreateStatusRejected {
		return api.TaskStatus_REJECTED
	}
	return api.TaskStatus_ACCEPTED
}

func toAPITaskView(task service.TaskView) *api.ShotPreviewTaskView {
	nodes := make([]*api.TaskNodeView, 0, len(task.Nodes))
	for _, node := range task.Nodes {
		nodes = append(nodes, &api.TaskNodeView{
			NodeId: node.ID, Status: mapNodeStatus(node.Status), Attempts: int32(node.Attempts),
			StartedAt: formatTime(node.StartedAt), FinishedAt: formatTime(node.FinishedAt), Failure: toAPIFailure(node.Failure),
		})
	}
	artifacts := make([]*api.TaskArtifactView, 0, len(task.Artifacts))
	for _, artifact := range task.Artifacts {
		artifacts = append(artifacts, &api.TaskArtifactView{Type: artifact.Type, Uri: artifact.URI, Name: artifact.Name})
	}
	return &api.ShotPreviewTaskView{
		TaskId: task.TaskID, Status: mapTaskStatus(task.Status), WorkflowId: task.WorkflowID, WorkflowVersion: task.WorkflowVersion,
		Nodes: nodes, Artifacts: artifacts, Failure: toAPIFailure(task.Failure),
		CreatedAt: task.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: task.UpdatedAt.UTC().Format(time.RFC3339Nano),
		FinishedAt: formatTime(task.FinishedAt),
	}
}

func mapTaskStatus(status service.TaskStatus) api.ProductionTaskStatus {
	switch status {
	case service.TaskStatusRunning:
		return api.ProductionTaskStatus_RUNNING
	case service.TaskStatusSucceeded:
		return api.ProductionTaskStatus_SUCCEEDED
	case service.TaskStatusFailed:
		return api.ProductionTaskStatus_FAILED
	case service.TaskStatusCancelled:
		return api.ProductionTaskStatus_CANCELLED
	default:
		return api.ProductionTaskStatus_PENDING
	}
}

func mapNodeStatus(status service.NodeStatus) api.ProductionTaskStatus {
	return mapTaskStatus(service.TaskStatus(status))
}

func toAPIFailure(failure *service.FailureView) *api.TaskFailure {
	if failure == nil {
		return nil
	}
	return &api.TaskFailure{Code: failure.Code, Message: failure.Message, Retryable: failure.Retryable}
}

func formatTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339Nano)
	return &formatted
}

func generateRequestID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
