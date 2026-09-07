package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

var ErrRejected = errors.New("shot preview task rejected")

type TaskStatus int

const (
	TaskStatusAccepted TaskStatus = iota + 1
	TaskStatusRejected
)

type ShotPreviewService interface {
	SubmitTask(ctx context.Context, userID, prompt, conversationID, requestID string) (taskID string, status TaskStatus, err error)
}

// ShotPreviewServiceImpl preserves the RPC service boundary while delegating
// asynchronous production to pipeline.Runner. Its zero value remains usable so
// existing server wiring can be upgraded without changing handler contracts.
type ShotPreviewServiceImpl struct {
	once   sync.Once
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

	input, err := pipeline.NewSnapshot(shotPreviewInput{
		UserID:         userID,
		Prompt:         prompt,
		ConversationID: strings.TrimSpace(conversationID),
		RequestID:      requestID,
	})
	if err != nil {
		return "", TaskStatusRejected, err
	}
	task, _, err := s.pipeline().Submit(ctx, pipeline.Submission{
		IdempotencyKey: userID + ":" + requestID,
		Input:          input,
		Workflow:       pipeline.ShotPreviewWorkflow(input.JSON),
	})
	if err != nil {
		return "", TaskStatusRejected, err
	}
	return task.ID, TaskStatusAccepted, nil
}

func (s *ShotPreviewServiceImpl) pipeline() *pipeline.Runner {
	s.once.Do(func() {
		if s.runner != nil {
			return
		}
		s.runner = pipeline.NewRunner(
			pipeline.NewMemoryRepository(),
			nil,
			pipeline.StepInvokerFunc(defaultStage),
			nil,
		)
	})
	return s.runner
}

// Pipeline exposes the runner for bootstrap recovery and future task-status
// endpoints without widening the existing RPC interface.
func (s *ShotPreviewServiceImpl) Pipeline() *pipeline.Runner {
	return s.pipeline()
}

type shotPreviewInput struct {
	UserID         string `json:"user_id"`
	Prompt         string `json:"prompt"`
	ConversationID string `json:"conversation_id,omitempty"`
	RequestID      string `json:"request_id"`
}

func defaultStage(_ context.Context, request pipeline.StepRequest) (json.RawMessage, error) {
	return json.Marshal(struct {
		Stage string `json:"stage"`
	}{Stage: request.Step})
}
