// Package skill manages configured Agent capabilities and their Tool bindings.
package skill

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/S-zhi/blender-shot-preview/internal/agent/tool"
)

var (
	ErrNotFound          = errors.New("skill not found")
	ErrDisabled          = errors.New("skill is disabled")
	ErrAlreadyRegistered = errors.New("skill already registered")
	ErrInvalidDefinition = errors.New("invalid skill definition")
)

// Status controls whether a Skill can be selected by an Agent.
type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

// Definition is persisted metadata for one reusable Agent capability.
type Definition struct {
	ID          string
	Name        string
	Description string
	ToolIDs     []string
	Status      Status
}

// Repository persists Skill definitions and their Tool bindings.
type Repository interface {
	Create(ctx context.Context, definition Definition) error
	Get(ctx context.Context, id string) (Definition, error)
	List(ctx context.Context) ([]Definition, error)
	Update(ctx context.Context, definition Definition) error
}

// Resolver is the Skill boundary consumed by Agent definition validation and
// the Eino Agent factory.
type Resolver interface {
	Exists(ctx context.Context, id string) bool
	Resolve(ctx context.Context, id string) ([]tool.Resolved, error)
}

// Registry loads Skill definitions and resolves each bound Tool through the
// Tool Registry. Skill is a grouping layer; executable code belongs to Tool.
type Registry struct {
	repository Repository
	tools      tool.Resolver
}

func NewRegistry(repository Repository, tools tool.Resolver) (*Registry, error) {
	if repository == nil || tools == nil {
		return nil, fmt.Errorf("%w: repository and tool resolver are required", ErrInvalidDefinition)
	}
	return &Registry{repository: repository, tools: tools}, nil
}

func (r *Registry) Exists(ctx context.Context, skillID string) bool {
	definition, err := r.repository.Get(ctx, strings.TrimSpace(skillID))
	if err != nil || definition.Status != StatusEnabled {
		return false
	}
	for _, toolID := range definition.ToolIDs {
		if !r.tools.Exists(ctx, toolID) {
			return false
		}
	}
	return true
}

func (r *Registry) Resolve(ctx context.Context, skillID string) ([]tool.Resolved, error) {
	definition, err := r.repository.Get(ctx, strings.TrimSpace(skillID))
	if err != nil {
		return nil, err
	}
	if definition.Status != StatusEnabled {
		return nil, fmt.Errorf("%w: %s", ErrDisabled, definition.ID)
	}
	resolved := make([]tool.Resolved, 0, len(definition.ToolIDs))
	for _, toolID := range definition.ToolIDs {
		entry, err := r.tools.Resolve(ctx, toolID)
		if err != nil {
			return nil, fmt.Errorf("resolve skill %q tool %q: %w", definition.ID, toolID, err)
		}
		resolved = append(resolved, entry)
	}
	return resolved, nil
}

// MemoryRepository is a test and bootstrap Repository implementation.
type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]Definition
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{items: make(map[string]Definition)}
}

func (r *MemoryRepository) Create(_ context.Context, definition Definition) error {
	definition, err := normalizeDefinition(definition)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[definition.ID]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, definition.ID)
	}
	r.items[definition.ID] = cloneDefinition(definition)
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, skillID string) (Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definition, exists := r.items[strings.TrimSpace(skillID)]
	if !exists {
		return Definition{}, fmt.Errorf("%w: %s", ErrNotFound, strings.TrimSpace(skillID))
	}
	return cloneDefinition(definition), nil
}

func (r *MemoryRepository) List(_ context.Context) ([]Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definitions := make([]Definition, 0, len(r.items))
	for _, definition := range r.items {
		definitions = append(definitions, cloneDefinition(definition))
	}
	return definitions, nil
}

func (r *MemoryRepository) Update(_ context.Context, definition Definition) error {
	definition, err := normalizeDefinition(definition)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[definition.ID]; !exists {
		return fmt.Errorf("%w: %s", ErrNotFound, definition.ID)
	}
	r.items[definition.ID] = cloneDefinition(definition)
	return nil
}

func normalizeDefinition(definition Definition) (Definition, error) {
	definition.ID = strings.TrimSpace(definition.ID)
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Description = strings.TrimSpace(definition.Description)
	definition.ToolIDs = uniqueTrimmed(definition.ToolIDs)
	if definition.ID == "" || definition.Name == "" || len(definition.ToolIDs) == 0 {
		return Definition{}, fmt.Errorf("%w: ID, name, and at least one tool are required", ErrInvalidDefinition)
	}
	if definition.Status == "" {
		definition.Status = StatusEnabled
	}
	if definition.Status != StatusEnabled && definition.Status != StatusDisabled {
		return Definition{}, fmt.Errorf("%w: unknown status %q", ErrInvalidDefinition, definition.Status)
	}
	return definition, nil
}

func uniqueTrimmed(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func cloneDefinition(definition Definition) Definition {
	definition.ToolIDs = append([]string(nil), definition.ToolIDs...)
	return definition
}
