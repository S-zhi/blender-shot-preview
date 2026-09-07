package v0_1

import (
	"context"
	"errors"
	"strings"
	"testing"

	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

type failingKeyService struct{ called bool }

func (s *failingKeyService) ManageKey(context.Context, llmgateway.KeyCommand) (llmgateway.KeyCommandResult, error) {
	s.called = true
	return llmgateway.KeyCommandResult{}, errors.New("sensitive-provider-secret")
}

func TestKeyHandlerRejectsInvalidEnumsAndRedactsErrors(t *testing.T) {
	service := &failingKeyService{}
	h := NewLLMKeyHandler(service)
	unknown := api.LLMProvider(99)
	for _, request := range []*api.ManageLLMKeyRequest{
		nil,
		{UserId: "", Operation: api.LLMKeyOperation_SAVE},
		{UserId: "owner", Operation: api.LLMKeyOperation(99)},
		{UserId: "owner", Operation: api.LLMKeyOperation_SAVE, Provider: &unknown},
	} {
		if _, err := h.ManageLLMKey(context.Background(), request); err == nil {
			t.Fatal("invalid request accepted")
		}
		if service.called {
			t.Fatal("invalid request reached service")
		}
	}
	_, err := h.ManageLLMKey(context.Background(), &api.ManageLLMKeyRequest{UserId: "owner", Operation: api.LLMKeyOperation_SAVE})
	status, ok := kerrors.FromBizStatusError(err)
	if !ok || status.BizStatusCode() != 500 || strings.Contains(err.Error(), "sensitive-provider-secret") {
		t.Fatalf("unsafe error: %v", err)
	}
}
