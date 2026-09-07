// Package tool manages configured tool metadata and Eino tool providers.
package tool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	einotool "github.com/cloudwego/eino/components/tool"
)

var (
	ErrNotFound          = errors.New("tool not found")
	ErrDisabled          = errors.New("tool is disabled")
	ErrAlreadyRegistered = errors.New("tool provider already registered")
	ErrProviderNotFound  = errors.New("tool provider not found")
	ErrInvalidDefinition = errors.New("invalid tool definition")
)

// Status controls whether a tool can be resolved for Agent execution.
type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

// Kind determines how the Agent runtime protects a tool call.
type Kind string

const (
	KindRead         Kind = "read"
	KindWrite        Kind = "write"
	KindIrreversible Kind = "irreversible"
)

// Definition is persisted metadata. The executable provider remains in code.
type Definition struct {
	ID          string
	Name        string
	Description string
	Kind        Kind
	Status      Status
}

// Repository persists Tool metadata. A DAO-backed implementation can replace
// MemoryRepository without changing the Registry or Agent runtime.
type Repository interface {
	Create(ctx context.Context, definition Definition) error
	Get(ctx context.Context, id string) (Definition, error)
	List(ctx context.Context) ([]Definition, error)
	Update(ctx context.Context, definition Definition) error
}

// Provider builds the executable Eino tool for a registered Tool ID.
type Provider interface {
	Build(ctx context.Context) (einotool.BaseTool, error)
}

// ProviderFunc adapts a function into a Provider.
type ProviderFunc func(ctx context.Context) (einotool.BaseTool, error)

func (f ProviderFunc) Build(ctx context.Context) (einotool.BaseTool, error) {
	return f(ctx)
}

// Resolved binds persisted metadata to the Eino tool created for one run.
type Resolved struct {
	Definition Definition
	Tool       einotool.BaseTool
}

// Resolver is the Tool boundary consumed by Skill Registry.
type Resolver interface {
	Exists(ctx context.Context, id string) bool
	Resolve(ctx context.Context, id string) (Resolved, error)
}

// Registry holds implementation-owned Tool providers. It never stores Eino
// tool instances, because each Agent run receives a freshly built tool.
type Registry struct {
	repository Repository

	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry(repository Repository) (*Registry, error) {
	if repository == nil {
		return nil, fmt.Errorf("%w: repository is required", ErrInvalidDefinition)
	}
	return &Registry{repository: repository, providers: make(map[string]Provider)}, nil
}

// Register associates an executable provider with a persisted Tool ID.
func (r *Registry) Register(toolID string, provider Provider) error {
	toolID = strings.TrimSpace(toolID)
	if toolID == "" || provider == nil {
		return fmt.Errorf("%w: tool ID and provider are required", ErrInvalidDefinition)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[toolID]; exists {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, toolID)
	}
	r.providers[toolID] = provider
	return nil
}

func (r *Registry) Exists(ctx context.Context, toolID string) bool {
	definition, err := r.repository.Get(ctx, strings.TrimSpace(toolID))
	if err != nil || definition.Status != StatusEnabled {
		return false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[definition.ID]
	return exists
}

func (r *Registry) Resolve(ctx context.Context, toolID string) (Resolved, error) {
	definition, err := r.repository.Get(ctx, strings.TrimSpace(toolID))
	if err != nil {
		return Resolved{}, err
	}
	if definition.Status != StatusEnabled {
		return Resolved{}, fmt.Errorf("%w: %s", ErrDisabled, definition.ID)
	}
	r.mu.RLock()
	provider, exists := r.providers[definition.ID]
	r.mu.RUnlock()
	if !exists {
		return Resolved{}, fmt.Errorf("%w: %s", ErrProviderNotFound, definition.ID)
	}
	base, err := provider.Build(ctx)
	if err != nil {
		return Resolved{}, fmt.Errorf("build tool %q: %w", definition.ID, err)
	}
	if base == nil {
		return Resolved{}, fmt.Errorf("build tool %q: provider returned nil", definition.ID)
	}
	return Resolved{Definition: definition, Tool: base}, nil
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
	r.items[definition.ID] = definition
	return nil
}

func (r *MemoryRepository) Get(_ context.Context, toolID string) (Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definition, exists := r.items[strings.TrimSpace(toolID)]
	if !exists {
		return Definition{}, fmt.Errorf("%w: %s", ErrNotFound, strings.TrimSpace(toolID))
	}
	return definition, nil
}

func (r *MemoryRepository) List(_ context.Context) ([]Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	definitions := make([]Definition, 0, len(r.items))
	for _, definition := range r.items {
		definitions = append(definitions, definition)
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
	r.items[definition.ID] = definition
	return nil
}

func normalizeDefinition(definition Definition) (Definition, error) {
	definition.ID = strings.TrimSpace(definition.ID)
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Description = strings.TrimSpace(definition.Description)
	if definition.ID == "" || definition.Name == "" {
		return Definition{}, fmt.Errorf("%w: ID and name are required", ErrInvalidDefinition)
	}
	if definition.Kind == "" {
		definition.Kind = KindRead
	}
	if definition.Kind != KindRead && definition.Kind != KindWrite && definition.Kind != KindIrreversible {
		return Definition{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidDefinition, definition.Kind)
	}
	if definition.Status == "" {
		definition.Status = StatusEnabled
	}
	if definition.Status != StatusEnabled && definition.Status != StatusDisabled {
		return Definition{}, fmt.Errorf("%w: unknown status %q", ErrInvalidDefinition, definition.Status)
	}
	return definition, nil
}
