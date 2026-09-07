package v0_1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type InvalidArgumentError struct {
	Field string
}

func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("invalid argument: %s is required", e.Field)
}

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

func (h *ShotPreviewHandler) CreateShotPreviewTask(
	ctx context.Context,
	request *api.CreateShotPreviewTaskRequest,
) (*api.CreateShotPreviewTaskResponse, error) {
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

	taskID, status, err := h.service.SubmitTask(
		ctx,
		userID,
		prompt,
		strings.TrimSpace(request.GetConversationId()),
		requestID,
	)
	if errors.Is(err, service.ErrRejected) {
		return &api.CreateShotPreviewTaskResponse{
			TaskId:    taskID,
			Status:    api.TaskStatus_REJECTED,
			RequestId: &requestID,
		}, nil
	}
	if err != nil {
		return nil, &InternalError{RequestID: requestID, Cause: err}
	}

	return &api.CreateShotPreviewTaskResponse{
		TaskId:    taskID,
		Status:    mapTaskStatus(status),
		RequestId: &requestID,
	}, nil
}

func mapTaskStatus(status service.TaskStatus) api.TaskStatus {
	if status == service.TaskStatusRejected {
		return api.TaskStatus_REJECTED
	}
	return api.TaskStatus_ACCEPTED
}

func generateRequestID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
