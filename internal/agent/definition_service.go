package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/agent/skill"
)

const (
	defaultMaxSteps        = 12
	maximumMaxSteps        = 40
	defaultMaxOutputTokens = 4096
	maximumMaxOutputTokens = 8192
	defaultTimeout         = 2 * time.Minute
	maximumTimeout         = 20 * time.Minute
)

type AgentDefinitionService interface {
	Create(ctx context.Context, request CreateAgentRequest) (AgentDefinition, error)
	Get(ctx context.Context, agentID string) (AgentDefinition, error)
	List(ctx context.Context) ([]AgentDefinition, error)
	Update(ctx context.Context, request UpdateAgentRequest) (AgentDefinition, error)
	Enable(ctx context.Context, agentID string) error
	Disable(ctx context.Context, agentID string) error
}

type DefinitionService struct {
	repository DefinitionRepository
	skills     skill.Resolver
	models     ModelResolver
	now        func() time.Time
}

func NewDefinitionService(repository DefinitionRepository, skills skill.Resolver, models ModelResolver) (*DefinitionService, error) {
	if repository == nil || skills == nil || models == nil {
		return nil, fmt.Errorf("%w: repository, skills, and models are required", ErrInvalidDefinition)
	}
	return &DefinitionService{repository: repository, skills: skills, models: models, now: time.Now}, nil
}

func (s *DefinitionService) Create(ctx context.Context, request CreateAgentRequest) (AgentDefinition, error) {
	definition := definitionFromCreate(request)
	if err := s.validate(ctx, &definition); err != nil {
		return AgentDefinition{}, err
	}
	now := s.now().UTC()
	definition.CreatedAt = now
	definition.UpdatedAt = now
	definition.Status = AgentStatusEnabled
	if err := s.repository.Create(ctx, definition); err != nil {
		return AgentDefinition{}, err
	}
	return definition, nil
}

func (s *DefinitionService) Get(ctx context.Context, agentID string) (AgentDefinition, error) {
	return s.repository.Get(ctx, strings.TrimSpace(agentID))
}

func (s *DefinitionService) List(ctx context.Context) ([]AgentDefinition, error) {
	return s.repository.List(ctx)
}

func (s *DefinitionService) Update(ctx context.Context, request UpdateAgentRequest) (AgentDefinition, error) {
	definition, err := s.repository.Get(ctx, strings.TrimSpace(request.ID))
	if err != nil {
		return AgentDefinition{}, err
	}
	definition.Name = request.Name
	definition.Description = request.Description
	definition.SystemPrompt = request.SystemPrompt
	definition.ModelProfile = request.ModelProfile
	definition.SkillIDs = request.SkillIDs
	definition.InputSchema = request.InputSchema
	definition.OutputSchema = request.OutputSchema
	definition.MaxSteps = request.MaxSteps
	definition.MaxOutputTokens = request.MaxOutputTokens
	definition.Timeout = request.Timeout
	if err := s.validate(ctx, &definition); err != nil {
		return AgentDefinition{}, err
	}
	definition.UpdatedAt = s.now().UTC()
	if err := s.repository.Update(ctx, definition); err != nil {
		return AgentDefinition{}, err
	}
	return definition, nil
}

func (s *DefinitionService) Enable(ctx context.Context, agentID string) error {
	return s.setStatus(ctx, agentID, AgentStatusEnabled)
}

func (s *DefinitionService) Disable(ctx context.Context, agentID string) error {
	return s.setStatus(ctx, agentID, AgentStatusDisabled)
}

func (s *DefinitionService) setStatus(ctx context.Context, agentID string, status AgentStatus) error {
	definition, err := s.repository.Get(ctx, strings.TrimSpace(agentID))
	if err != nil {
		return err
	}
	definition.Status = status
	definition.UpdatedAt = s.now().UTC()
	return s.repository.Update(ctx, definition)
}

func (s *DefinitionService) validate(ctx context.Context, definition *AgentDefinition) error {
	definition.ID = strings.TrimSpace(definition.ID)
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Description = strings.TrimSpace(definition.Description)
	definition.SystemPrompt = strings.TrimSpace(definition.SystemPrompt)
	definition.ModelProfile = strings.TrimSpace(definition.ModelProfile)
	if definition.ID == "" || definition.Name == "" || definition.SystemPrompt == "" || definition.ModelProfile == "" {
		return fmt.Errorf("%w: ID, name, system prompt, and model profile are required", ErrInvalidDefinition)
	}
	if !s.models.Exists(ctx, definition.ModelProfile) {
		return fmt.Errorf("%w: model profile %q does not exist", ErrInvalidDefinition, definition.ModelProfile)
	}
	definition.SkillIDs = uniqueTrimmed(definition.SkillIDs)
	for _, skillID := range definition.SkillIDs {
		if !s.skills.Exists(ctx, skillID) {
			return fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
		}
	}
	if definition.MaxSteps == 0 {
		definition.MaxSteps = defaultMaxSteps
	}
	if definition.MaxSteps < 1 || definition.MaxSteps > maximumMaxSteps {
		return fmt.Errorf("%w: max steps must be between 1 and %d", ErrInvalidDefinition, maximumMaxSteps)
	}
	if definition.MaxOutputTokens == 0 {
		definition.MaxOutputTokens = defaultMaxOutputTokens
	}
	if definition.MaxOutputTokens < 1 || definition.MaxOutputTokens > maximumMaxOutputTokens {
		return fmt.Errorf("%w: max output tokens must be between 1 and %d", ErrInvalidDefinition, maximumMaxOutputTokens)
	}
	if definition.Timeout == 0 {
		definition.Timeout = defaultTimeout
	}
	if definition.Timeout < time.Second || definition.Timeout > maximumTimeout {
		return fmt.Errorf("%w: timeout must be between 1s and %s", ErrInvalidDefinition, maximumTimeout)
	}
	if err := validateSchema(definition.InputSchema, "input schema"); err != nil {
		return err
	}
	if err := validateSchema(definition.OutputSchema, "output schema"); err != nil {
		return err
	}
	return nil
}

func definitionFromCreate(request CreateAgentRequest) AgentDefinition {
	return AgentDefinition{
		ID: request.ID, Name: request.Name, Description: request.Description,
		SystemPrompt: request.SystemPrompt, ModelProfile: request.ModelProfile,
		SkillIDs: request.SkillIDs, InputSchema: request.InputSchema,
		OutputSchema: request.OutputSchema, MaxSteps: request.MaxSteps, Timeout: request.Timeout,
		MaxOutputTokens: request.MaxOutputTokens,
	}
}

func validateSchema(value json.RawMessage, name string) error {
	if len(value) == 0 {
		return nil
	}
	var parsed any
	if err := json.Unmarshal(value, &parsed); err != nil {
		return fmt.Errorf("%w: %s is not valid JSON", ErrInvalidDefinition, name)
	}
	if _, ok := parsed.(map[string]any); !ok {
		return fmt.Errorf("%w: %s must be a JSON object", ErrInvalidDefinition, name)
	}
	return nil
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
