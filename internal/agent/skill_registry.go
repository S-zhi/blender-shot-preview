package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

// SkillRegistry resolves registered implementation-owned Skills. Agent callers
// only name a skill ID; they cannot pass executable Go objects.
type SkillRegistry interface {
	Register(skillID string, build func(context.Context) (tool.BaseTool, error)) error
	RegisterWithKind(skillID string, kind ToolKind, build func(context.Context) (tool.BaseTool, error)) error
	Resolve(ctx context.Context, skillID string) (tool.BaseTool, error)
	Kind(skillID string) (ToolKind, error)
	Exists(skillID string) bool
}

type MemorySkillRegistry struct {
	mu       sync.RWMutex
	builders map[string]registeredSkill
}

type registeredSkill struct {
	kind  ToolKind
	build func(context.Context) (tool.BaseTool, error)
}

func NewMemorySkillRegistry() *MemorySkillRegistry {
	return &MemorySkillRegistry{builders: make(map[string]registeredSkill)}
}

func (r *MemorySkillRegistry) Register(skillID string, build func(context.Context) (tool.BaseTool, error)) error {
	return r.RegisterWithKind(skillID, ToolKindRead, build)
}

func (r *MemorySkillRegistry) RegisterWithKind(skillID string, kind ToolKind, build func(context.Context) (tool.BaseTool, error)) error {
	if skillID == "" || build == nil {
		return fmt.Errorf("%w: skill ID and builder are required", ErrInvalidDefinition)
	}
	if kind != ToolKindRead && kind != ToolKindWrite && kind != ToolKindIrreversible {
		return fmt.Errorf("%w: unknown tool kind %q", ErrInvalidDefinition, kind)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.builders[skillID]; exists {
		return fmt.Errorf("%w: %s", ErrInvalidDefinition, skillID)
	}
	r.builders[skillID] = registeredSkill{kind: kind, build: build}
	return nil
}

func (r *MemorySkillRegistry) Resolve(ctx context.Context, skillID string) (tool.BaseTool, error) {
	r.mu.RLock()
	skill, exists := r.builders[skillID]
	r.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}
	return skill.build(ctx)
}

func (r *MemorySkillRegistry) Kind(skillID string) (ToolKind, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, exists := r.builders[skillID]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}
	return skill.kind, nil
}

func (r *MemorySkillRegistry) Exists(skillID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.builders[skillID]
	return exists
}
