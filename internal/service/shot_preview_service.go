package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

// ShotPreviewServiceImpl preserves the RPC service boundary while delegating
// asynchronous production to a fully configured pipeline.Runner.
type ShotPreviewServiceImpl struct {
	runner *pipeline.Runner
}

func NewShotPreviewService(runner *pipeline.Runner) *ShotPreviewServiceImpl {
	return &ShotPreviewServiceImpl{runner: runner}
}

func (s *ShotPreviewServiceImpl) CreateTask(ctx context.Context, request CreateTaskRequest) (CreateTaskResult, error) {
	userID := strings.TrimSpace(request.UserID)
	prompt := strings.TrimSpace(request.Prompt)
	requestID := strings.TrimSpace(request.RequestID)
	if userID == "" || prompt == "" || requestID == "" {
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, ErrInvalidTaskRequest
	}
	if workflowID := strings.TrimSpace(request.WorkflowID); workflowID != "" && workflowID != ShotPreviewWorkflowID {
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, ErrInvalidTaskRequest
	}
	if version := strings.TrimSpace(request.WorkflowVersion); version != "" && version != ShotPreviewWorkflowVersion {
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, ErrInvalidTaskRequest
	}
	if s == nil || s.runner == nil {
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, ErrPipelineUnavailable
	}

	input, err := pipeline.NewSnapshot(shotPreviewInput{
		UserID:         userID,
		Prompt:         prompt,
		ConversationID: strings.TrimSpace(request.ConversationID),
		RequestID:      requestID,
	})
	if err != nil {
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, err
	}
	task, created, err := s.runner.Submit(ctx, pipeline.Submission{
		IdempotencyKey: userID + ":" + requestID,
		Input:          input,
		Workflow:       pipeline.ShotPreviewWorkflow(input.JSON),
	})
	if err != nil {
		if errors.Is(err, pipeline.ErrIdempotencyConflict) {
			return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, ErrInvalidTaskRequest
		}
		return CreateTaskResult{RequestID: requestID, Status: CreateStatusRejected}, err
	}
	return CreateTaskResult{
		TaskID: task.ID, RequestID: requestID, Status: CreateStatusAccepted, Replayed: !created,
	}, nil
}

func (s *ShotPreviewServiceImpl) GetTask(ctx context.Context, request GetTaskRequest) (TaskView, error) {
	task, err := s.taskForUser(ctx, request.UserID, request.TaskID)
	if err != nil {
		return TaskView{}, err
	}
	return taskView(task), nil
}

func (s *ShotPreviewServiceImpl) CancelTask(ctx context.Context, request CancelTaskRequest) (TaskView, error) {
	if _, err := s.taskForUser(ctx, request.UserID, request.TaskID); err != nil {
		return TaskView{}, err
	}
	if err := s.runner.Cancel(ctx, strings.TrimSpace(request.TaskID)); err != nil {
		if errors.Is(err, pipeline.ErrTaskAlreadyTerminal) {
			return TaskView{}, ErrTaskNotCancellable
		}
		if errors.Is(err, pipeline.ErrTaskNotFound) {
			return TaskView{}, ErrTaskNotFound
		}
		return TaskView{}, err
	}
	return s.GetTask(ctx, GetTaskRequest{UserID: request.UserID, TaskID: request.TaskID})
}

func (s *ShotPreviewServiceImpl) RetryTask(ctx context.Context, request RetryTaskRequest) (TaskView, error) {
	if _, err := s.taskForUser(ctx, request.UserID, request.TaskID); err != nil {
		return TaskView{}, err
	}
	task, err := s.runner.Retry(ctx, strings.TrimSpace(request.TaskID))
	if err != nil {
		if errors.Is(err, pipeline.ErrTaskNotFailed) {
			return TaskView{}, ErrTaskNotRetryable
		}
		if errors.Is(err, pipeline.ErrTaskNotFound) {
			return TaskView{}, ErrTaskNotFound
		}
		return TaskView{}, err
	}
	return taskView(task), nil
}

func (s *ShotPreviewServiceImpl) taskForUser(ctx context.Context, userID, taskID string) (pipeline.Task, error) {
	if s == nil || s.runner == nil {
		return pipeline.Task{}, ErrPipelineUnavailable
	}
	userID = strings.TrimSpace(userID)
	taskID = strings.TrimSpace(taskID)
	if userID == "" || taskID == "" {
		return pipeline.Task{}, ErrInvalidTaskRequest
	}
	task, err := s.runner.Get(ctx, taskID)
	if errors.Is(err, pipeline.ErrTaskNotFound) {
		return pipeline.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return pipeline.Task{}, err
	}
	input, err := pipeline.DecodeSnapshot[shotPreviewInput](task.Input)
	if err != nil || input.UserID != userID {
		// Returning not found for a task owned by another caller avoids turning
		// this endpoint into a task-ID enumeration oracle.
		return pipeline.Task{}, ErrTaskNotFound
	}
	return task, nil
}

func taskView(task pipeline.Task) TaskView {
	view := TaskView{
		TaskID:          task.ID,
		Status:          TaskStatus(task.Status),
		WorkflowID:      ShotPreviewWorkflowID,
		WorkflowVersion: ShotPreviewWorkflowVersion,
		Nodes:           make([]NodeView, 0, len(task.Nodes)),
		Artifacts:       []ArtifactView{},
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
		FinishedAt:      cloneTime(task.FinishedAt),
	}
	for _, node := range task.Nodes {
		nodeView := NodeView{
			ID:         string(node.ID),
			Status:     NodeStatus(node.Status),
			Attempts:   node.Attempts,
			StartedAt:  cloneTime(node.StartedAt),
			FinishedAt: cloneTime(node.FinishedAt),
		}
		if node.Status == pipeline.NodeStatusFailed {
			nodeView.Failure = &FailureView{
				Code: "NODE_FAILED", Message: "Production step failed", Retryable: true,
			}
			if view.Failure == nil {
				view.Failure = nodeView.Failure
			}
		}
		view.Nodes = append(view.Nodes, nodeView)
	}
	return view
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
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
