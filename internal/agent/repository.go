package agent

import (
	"context"
	"sync"
)

type DefinitionRepository interface {
	Create(ctx context.Context, definition AgentDefinition) error
	Get(ctx context.Context, id string) (AgentDefinition, error)
	List(ctx context.Context) ([]AgentDefinition, error)
	Update(ctx context.Context, definition AgentDefinition) error
}

type RunRepository interface {
	Create(ctx context.Context, run AgentRun) error
	Get(ctx context.Context, runID string) (AgentRun, error)
	Update(ctx context.Context, run AgentRun) error
	AppendEvent(ctx context.Context, event AgentEvent) error
	ListEvents(ctx context.Context, runID string) ([]AgentEvent, error)
}

// MemoryDefinitionRepository is a test and bootstrap implementation. Production
// code can replace it with a DAO-backed implementation without changing services.
type MemoryDefinitionRepository struct {
	mu    sync.RWMutex
	items map[string]AgentDefinition
}

func NewMemoryDefinitionRepository() *MemoryDefinitionRepository {
	return &MemoryDefinitionRepository{items: make(map[string]AgentDefinition)}
}

func (r *MemoryDefinitionRepository) Create(_ context.Context, definition AgentDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[definition.ID]; exists {
		return ErrInvalidDefinition
	}
	r.items[definition.ID] = cloneDefinition(definition)
	return nil
}

func (r *MemoryDefinitionRepository) Get(_ context.Context, id string) (AgentDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definition, exists := r.items[id]
	if !exists {
		return AgentDefinition{}, ErrAgentNotFound
	}
	return cloneDefinition(definition), nil
}

func (r *MemoryDefinitionRepository) List(_ context.Context) ([]AgentDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definitions := make([]AgentDefinition, 0, len(r.items))
	for _, definition := range r.items {
		definitions = append(definitions, cloneDefinition(definition))
	}
	return definitions, nil
}

func (r *MemoryDefinitionRepository) Update(_ context.Context, definition AgentDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[definition.ID]; !exists {
		return ErrAgentNotFound
	}
	r.items[definition.ID] = cloneDefinition(definition)
	return nil
}

// MemoryRunRepository is a test and bootstrap implementation of run auditing.
type MemoryRunRepository struct {
	mu     sync.RWMutex
	runs   map[string]AgentRun
	events map[string][]AgentEvent
}

func NewMemoryRunRepository() *MemoryRunRepository {
	return &MemoryRunRepository{runs: make(map[string]AgentRun), events: make(map[string][]AgentEvent)}
}

func (r *MemoryRunRepository) Create(_ context.Context, run AgentRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.ID] = cloneRun(run)
	return nil
}

func (r *MemoryRunRepository) Get(_ context.Context, runID string) (AgentRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, exists := r.runs[runID]
	if !exists {
		return AgentRun{}, ErrRunNotFound
	}
	return cloneRun(run), nil
}

func (r *MemoryRunRepository) Update(_ context.Context, run AgentRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.runs[run.ID]; !exists {
		return ErrRunNotFound
	}
	r.runs[run.ID] = cloneRun(run)
	return nil
}

func (r *MemoryRunRepository) AppendEvent(_ context.Context, event AgentEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.runs[event.RunID]; !exists {
		return ErrRunNotFound
	}
	r.events[event.RunID] = append(r.events[event.RunID], event)
	return nil
}

func (r *MemoryRunRepository) ListEvents(_ context.Context, runID string) ([]AgentEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, exists := r.runs[runID]; !exists {
		return nil, ErrRunNotFound
	}
	return append([]AgentEvent(nil), r.events[runID]...), nil
}

func cloneDefinition(definition AgentDefinition) AgentDefinition {
	definition.SkillIDs = append([]string(nil), definition.SkillIDs...)
	definition.InputSchema = append([]byte(nil), definition.InputSchema...)
	definition.OutputSchema = append([]byte(nil), definition.OutputSchema...)
	return definition
}

func cloneRun(run AgentRun) AgentRun {
	run.ConfigSnapshot = append([]byte(nil), run.ConfigSnapshot...)
	return run
}
