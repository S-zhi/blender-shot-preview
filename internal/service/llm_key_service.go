package service

import (
	"context"
	"fmt"

	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
)

// LLMKeyService translates the public key-management use case into the
// gateway's provider-independent command model.
type LLMKeyService interface {
	ManageKey(ctx context.Context, command llmgateway.KeyCommand) (llmgateway.KeyCommandResult, error)
}

type LLMKeyServiceImpl struct {
	manager llmgateway.KeyManager
}

func NewLLMKeyService(manager llmgateway.KeyManager) *LLMKeyServiceImpl {
	return &LLMKeyServiceImpl{manager: manager}
}

func (s *LLMKeyServiceImpl) ManageKey(ctx context.Context, command llmgateway.KeyCommand) (llmgateway.KeyCommandResult, error) {
	if s == nil || s.manager == nil {
		return llmgateway.KeyCommandResult{}, fmt.Errorf("llm key manager is unavailable")
	}
	return s.manager.Execute(ctx, command)
}
