package service

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidTaskRequest  = errors.New("invalid shot preview task request")
	ErrTaskNotFound        = errors.New("shot preview task not found")
	ErrTaskNotCancellable  = errors.New("shot preview task is not cancellable")
	ErrTaskNotRetryable    = errors.New("shot preview task is not retryable")
	ErrPipelineUnavailable = errors.New("shot preview pipeline is unavailable")
)

const (
	ShotPreviewWorkflowID      = "shot-preview"
	ShotPreviewWorkflowVersion = "v1"
)

// ShotPreviewService is the stable application boundary used by RPC handlers.
// Its types deliberately omit pipeline invocations, raw snapshots, and local
// workspace paths so those implementation details cannot become RPC contracts.
type ShotPreviewService interface {
	CreateTask(ctx context.Context, request CreateTaskRequest) (CreateTaskResult, error)
	GetTask(ctx context.Context, request GetTaskRequest) (TaskView, error)
	CancelTask(ctx context.Context, request CancelTaskRequest) (TaskView, error)
	RetryTask(ctx context.Context, request RetryTaskRequest) (TaskView, error)
}

type CreateTaskRequest struct {
	UserID          string
	Prompt          string
	ConversationID  string
	RequestID       string
	WorkflowID      string
	WorkflowVersion string
}

type CreateTaskResult struct {
	TaskID    string
	RequestID string
	Status    CreateStatus
	Replayed  bool
}

type CreateStatus string

const (
	CreateStatusAccepted CreateStatus = "accepted"
	CreateStatusRejected CreateStatus = "rejected"
)

type GetTaskRequest struct {
	UserID string
	TaskID string
}

type CancelTaskRequest struct {
	UserID string
	TaskID string
}

type RetryTaskRequest struct {
	UserID string
	TaskID string
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusSucceeded NodeStatus = "succeeded"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusCancelled NodeStatus = "cancelled"
)

type TaskView struct {
	TaskID          string
	Status          TaskStatus
	WorkflowID      string
	WorkflowVersion string
	Nodes           []NodeView
	Artifacts       []ArtifactView
	Failure         *FailureView
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      *time.Time
}

type NodeView struct {
	ID         string
	Status     NodeStatus
	Attempts   int
	StartedAt  *time.Time
	FinishedAt *time.Time
	Failure    *FailureView
}

type ArtifactView struct {
	Type string
	URI  string
	Name string
}

type FailureView struct {
	Code      string
	Message   string
	Retryable bool
}
