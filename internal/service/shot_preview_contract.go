package service

import (
	"context"
	"errors"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
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
	ConfirmStep(ctx context.Context, request ConfirmStepRequest) error
	AdjustStep(ctx context.Context, request AdjustStepRequest) error
	SubscribeEvents(ctx context.Context, taskID string) (<-chan pipeline.PipelineEvent, func(), error)
}

type CreateTaskRequest struct {
	UserID              string
	Prompt              string
	ConversationID      string
	RequestID           string
	WorkflowID          string
	WorkflowVersion     string
	RequireConfirmation bool
}

type ConfirmStepRequest struct {
	UserID         string
	TaskID         string
	NodeID         string
	AdjustedOutput *string
}

type AdjustStepRequest struct {
	UserID     string
	TaskID     string
	NodeID     string
	OutputJSON string
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
	NodeStatusPending             NodeStatus = "pending"
	NodeStatusRunning             NodeStatus = "running"
	NodeStatusWaitingConfirmation NodeStatus = "waiting_confirmation"
	NodeStatusSucceeded           NodeStatus = "succeeded"
	NodeStatusFailed              NodeStatus = "failed"
	NodeStatusCancelled           NodeStatus = "cancelled"
)

type TaskView struct {
	TaskID          string         `json:"task_id"`
	Status          TaskStatus     `json:"status"`
	WorkflowID      string         `json:"workflow_id"`
	WorkflowVersion string         `json:"workflow_version"`
	Nodes           []NodeView     `json:"nodes"`
	Artifacts       []ArtifactView `json:"artifacts"`
	Failure         *FailureView   `json:"failure,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	FinishedAt      *time.Time     `json:"finished_at,omitempty"`
}

type NodeView struct {
	ID         string       `json:"node_id"`
	Status     NodeStatus   `json:"status"`
	Attempts   int          `json:"attempts"`
	Input      *string      `json:"input,omitempty"`
	Output     *string      `json:"output,omitempty"`
	ErrorCode  string       `json:"error_code,omitempty"`
	StartedAt  *time.Time   `json:"started_at,omitempty"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
	Failure    *FailureView `json:"failure,omitempty"`
}

type ArtifactView struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type FailureView struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
