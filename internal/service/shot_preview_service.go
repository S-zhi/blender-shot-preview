package service

import (
	"context"
	"errors"
	"strings"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

var ErrRejected = errors.New("shot preview task rejected")
var ErrPipelineUnavailable = errors.New("shot preview pipeline is unavailable")

type TaskStatus int

const (
	TaskStatusAccepted TaskStatus = iota + 1
	TaskStatusRejected
)

type ShotPreviewService interface {
	SubmitTask(ctx context.Context, userID, prompt, conversationID, requestID string) (taskID string, status TaskStatus, err error)
}

// ShotPreviewServiceImpl preserves the RPC service boundary while delegating
// asynchronous production to a fully configured pipeline.Runner.
type ShotPreviewServiceImpl struct {
	runner *pipeline.Runner
}

func NewShotPreviewService(runner *pipeline.Runner) *ShotPreviewServiceImpl {
	return &ShotPreviewServiceImpl{runner: runner}
}

func (s *ShotPreviewServiceImpl) SubmitTask(ctx context.Context, userID, prompt, conversationID, requestID string) (string, TaskStatus, error) {
	userID = strings.TrimSpace(userID)
	prompt = strings.TrimSpace(prompt)
	requestID = strings.TrimSpace(requestID)
	if userID == "" || prompt == "" || requestID == "" {
		return "", TaskStatusRejected, ErrRejected
	}
	if s == nil || s.runner == nil {
		return "", TaskStatusRejected, ErrPipelineUnavailable
	}

	input, err := pipeline.NewSnapshot(shotPreviewInput{
		UserID:         userID,
		Prompt:         prompt,
		ConversationID: strings.TrimSpace(conversationID),
		RequestID:      requestID,
	})
	if err != nil {
		return "", TaskStatusRejected, err
	}
	task, _, err := s.runner.Submit(ctx, pipeline.Submission{
		IdempotencyKey: userID + ":" + requestID,
		Input:          input,
		Workflow:       pipeline.ShotPreviewWorkflow(input.JSON),
	})
	if err != nil {
		return "", TaskStatusRejected, err
	}
	return task.ID, TaskStatusAccepted, nil
}

// Pipeline exposes the runner for bootstrap recovery and future task-status
// endpoints without widening the existing RPC interface.
func (s *ShotPreviewServiceImpl) Pipeline() *pipeline.Runner {
	if s == nil {
		return nil
	}
	return s.runner
}

type shotPreviewInput struct {
	UserID         string `json:"user_id"`
	Prompt         string `json:"prompt"`
	ConversationID string `json:"conversation_id,omitempty"`
	RequestID      string `json:"request_id"`
}
