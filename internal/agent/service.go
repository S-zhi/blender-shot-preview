package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

type AgentService interface {
	Run(ctx context.Context, request AgentRequest) (*AgentResult, error)
	Stream(ctx context.Context, request AgentRequest) (<-chan AgentEvent, error)
	Submit(ctx context.Context, request AgentRequest) (string, error)
	GetRun(ctx context.Context, runID string) (*AgentRun, error)
	Cancel(ctx context.Context, runID string) error
}

type Service struct {
	definitions DefinitionRepository
	runs        RunRepository
	factory     AgentFactory
	now         func() time.Time
	newRunID    func() (string, error)

	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewService(definitions DefinitionRepository, runs RunRepository, factory AgentFactory) (*Service, error) {
	if definitions == nil || runs == nil || factory == nil {
		return nil, fmt.Errorf("%w: definitions, runs, and factory are required", ErrInvalidDefinition)
	}
	return &Service{
		definitions: definitions,
		runs:        runs,
		factory:     factory,
		now:         time.Now,
		newRunID:    newRunID,
		cancels:     make(map[string]context.CancelFunc),
	}, nil
}

func (s *Service) Run(ctx context.Context, request AgentRequest) (*AgentResult, error) {
	run, definition, err := s.prepare(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.execute(ctx, run, definition, false, nil)
}

func (s *Service) Stream(ctx context.Context, request AgentRequest) (<-chan AgentEvent, error) {
	run, definition, err := s.prepare(ctx, request)
	if err != nil {
		return nil, err
	}
	events := make(chan AgentEvent, 32)
	go func() {
		defer close(events)
		_, _ = s.execute(ctx, run, definition, true, events)
	}()
	return events, nil
}

func (s *Service) Submit(_ context.Context, _ AgentRequest) (string, error) {
	return "", ErrNotImplemented
}

func (s *Service) GetRun(ctx context.Context, runID string) (*AgentRun, error) {
	run, err := s.runs.Get(ctx, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *Service) Cancel(_ context.Context, runID string) error {
	runID = strings.TrimSpace(runID)
	s.mu.Lock()
	cancel, exists := s.cancels[runID]
	s.mu.Unlock()
	if !exists {
		if _, err := s.runs.Get(context.Background(), runID); err != nil {
			return err
		}
		return ErrRunCancelled
	}
	cancel()
	return nil
}

func (s *Service) prepare(ctx context.Context, request AgentRequest) (AgentRun, AgentDefinition, error) {
	if strings.TrimSpace(request.AgentID) == "" || strings.TrimSpace(request.Input) == "" {
		return AgentRun{}, AgentDefinition{}, fmt.Errorf("%w: agent ID and input are required", ErrInvalidDefinition)
	}
	if len(request.SkillIDs) > 0 {
		return AgentRun{}, AgentDefinition{}, ErrDynamicSkillBlocked
	}
	definition, err := s.definitions.Get(ctx, strings.TrimSpace(request.AgentID))
	if err != nil {
		return AgentRun{}, AgentDefinition{}, err
	}
	if definition.Status != AgentStatusEnabled {
		return AgentRun{}, AgentDefinition{}, ErrAgentDisabled
	}
	if err := validateInput(definition.InputSchema, request.Input); err != nil {
		return AgentRun{}, AgentDefinition{}, err
	}
	runID, err := s.newRunID()
	if err != nil {
		return AgentRun{}, AgentDefinition{}, fmt.Errorf("create agent run ID: %w", err)
	}
	snapshot, err := json.Marshal(definition)
	if err != nil {
		return AgentRun{}, AgentDefinition{}, fmt.Errorf("snapshot agent definition: %w", err)
	}
	now := s.now().UTC()
	run := AgentRun{
		ID: runID, AgentID: definition.ID, SessionID: request.SessionID,
		UserID: request.UserID, TenantID: request.TenantID, Input: request.Input,
		Status: RunStatusPending, ConfigSnapshot: snapshot, StartedAt: now,
	}
	if err := s.runs.Create(ctx, run); err != nil {
		return AgentRun{}, AgentDefinition{}, err
	}
	return run, definition, nil
}

func (s *Service) execute(parent context.Context, run AgentRun, definition AgentDefinition, streaming bool, out chan<- AgentEvent) (*AgentResult, error) {
	ctx, cancel := context.WithTimeout(withRunID(parent, run.ID), definition.Timeout)
	s.registerCancel(run.ID, cancel)
	defer func() {
		cancel()
		s.unregisterCancel(run.ID)
	}()

	run.Status = RunStatusRunning
	if err := s.runs.Update(context.Background(), run); err != nil {
		return nil, err
	}
	sequence := 0
	publish := func(event AgentEvent) {
		sequence++
		event.RunID = run.ID
		event.Sequence = sequence
		event.CreatedAt = s.now().UTC()
		_ = s.runs.AppendEvent(context.Background(), event)
		if out != nil {
			out <- event
		}
	}
	publish(AgentEvent{Type: EventAgentStarted})

	runtime, err := s.factory.Build(ctx, definition, streaming)
	if err == nil {
		err = s.consume(ctx, runtime, run.Input, publish, &run)
	}
	run.FinishedAt = pointerTo(s.now().UTC())
	if err != nil {
		run.Error = safeError(err)
		run.Status = terminalStatus(ctx, err)
		publish(AgentEvent{Type: terminalEvent(run.Status), Content: run.Error, ErrorCode: string(run.Status)})
	} else {
		if validationErr := validateOutput(definition.OutputSchema, run.Output); validationErr != nil {
			run.Error = safeError(validationErr)
			run.Status = RunStatusFailed
			publish(AgentEvent{Type: EventAgentFailed, Content: run.Error, ErrorCode: "output_validation"})
			err = validationErr
		} else {
			run.Status = RunStatusSucceeded
			publish(AgentEvent{Type: EventAgentComplete, Content: run.Output})
		}
	}
	if updateErr := s.runs.Update(context.Background(), run); updateErr != nil {
		return nil, updateErr
	}
	return &AgentResult{RunID: run.ID, Status: run.Status, Output: run.Output, Error: run.Error, Started: run.StartedAt, Ended: *run.FinishedAt}, err
}

func (s *Service) consume(ctx context.Context, runtime *RuntimeAgent, input string, publish func(AgentEvent), run *AgentRun) error {
	iterator := runtime.Runner.Query(ctx, input)
	for {
		event, ok := iterator.Next()
		if !ok {
			return nil
		}
		if event.Err != nil {
			return event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		message, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			return err
		}
		if message == nil {
			continue
		}
		switch event.Output.MessageOutput.Role {
		case schema.Tool:
			publish(AgentEvent{Type: EventToolFinished, ToolName: event.Output.MessageOutput.ToolName, Content: message.Content})
		case schema.Assistant:
			for _, call := range message.ToolCalls {
				publish(AgentEvent{Type: EventToolStarted, ToolName: call.Function.Name})
			}
			if message.Content != "" {
				run.Output = message.Content
				publish(AgentEvent{Type: EventMessageDelta, Content: message.Content})
			}
		}
	}
}

func (s *Service) registerCancel(runID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancels[runID] = cancel
}

func (s *Service) unregisterCancel(runID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, runID)
}

func newRunID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func validateInput(schemaJSON json.RawMessage, input string) error {
	if len(schemaJSON) == 0 {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return fmt.Errorf("%w: input must match configured JSON schema", ErrInvalidDefinition)
	}
	return validateRequiredFields(schemaJSON, value, "input")
}

func validateOutput(schemaJSON json.RawMessage, output string) error {
	if len(schemaJSON) == 0 {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(output), &value); err != nil {
		return fmt.Errorf("%w: output must match configured JSON schema", ErrInvalidDefinition)
	}
	return validateRequiredFields(schemaJSON, value, "output")
}

func validateRequiredFields(schemaJSON json.RawMessage, value map[string]any, name string) error {
	var schemaDefinition struct {
		Type       string   `json:"type"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(schemaJSON, &schemaDefinition); err != nil {
		return fmt.Errorf("%w: invalid %s schema", ErrInvalidDefinition, name)
	}
	for _, field := range schemaDefinition.Required {
		if _, exists := value[field]; !exists {
			return fmt.Errorf("%w: %s missing required field %q", ErrInvalidDefinition, name, field)
		}
	}
	if schemaDefinition.Type != "" && schemaDefinition.Type != "object" {
		return fmt.Errorf("%w: %s schema root type must be object", ErrInvalidDefinition, name)
	}
	for field, property := range schemaDefinition.Properties {
		fieldValue, exists := value[field]
		if !exists || property.Type == "" {
			continue
		}
		if !matchesJSONType(fieldValue, property.Type) {
			return fmt.Errorf("%w: %s field %q must be %s", ErrInvalidDefinition, name, field, property.Type)
		}
	}
	return nil
}

func matchesJSONType(value any, expected string) bool {
	switch expected {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && number == float64(int64(number))
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	default:
		return false
	}
}

func terminalStatus(ctx context.Context, err error) RunStatus {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return RunStatusTimeout
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return RunStatusCancelled
	}
	return RunStatusFailed
}

func terminalEvent(status RunStatus) EventType {
	if status == RunStatusCancelled {
		return EventAgentCanceled
	}
	return EventAgentFailed
}

func safeError(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func pointerTo(value time.Time) *time.Time { return &value }
