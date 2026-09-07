package service

import (
	"context"
	"errors"
	"fmt"
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

// ShotPreviewServiceImpl is the MVP service placeholder. Agent Runtime and DAO
// integrations will be added behind this service boundary in later iterations.
type ShotPreviewServiceImpl struct{}

func (s *ShotPreviewServiceImpl) SubmitTask(_ context.Context, userID, prompt, _, _ string) (string, TaskStatus, error) {
	return fmt.Sprintf("task-%s", userID), TaskStatusAccepted, nil
}
