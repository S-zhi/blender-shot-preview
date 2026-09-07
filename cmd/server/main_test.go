package main

import (
	"context"
	"strings"
	"testing"

	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	handler "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

func TestKeyLifecycleAndRegistration(t *testing.T) {
	for _, provider := range []api.LLMProvider{api.LLMProvider_OPENAI, api.LLMProvider_ANTHROPIC} {
		t.Run(provider.String(), func(t *testing.T) {
			store := &inMemoryCredentialStore{credentials: make(map[string]llmgateway.CredentialRecord)}
			cipher, err := llmgateway.NewAESGCMCipher(make([]byte, 32))
			if err != nil {
				t.Fatal(err)
			}
			gateway, err := llmgateway.NewGateway(store, cipher)
			if err != nil {
				t.Fatal(err)
			}
			svr, err := newServer(gateway)
			if err != nil {
				t.Fatal(err)
			}
			infos := svr.GetServiceInfos()
			if infos["LLMKeyServiceV0_1"] == nil || infos["ShotPreviewServiceV0_1"] == nil {
				t.Fatal("both services must be registered")
			}
			h := handler.NewLLMKeyHandler(service.NewLLMKeyService(gateway))
			ctx := context.Background()
			name, secret := "test key", "fake-test-secret"
			created, err := h.ManageLLMKey(ctx, &api.ManageLLMKeyRequest{UserId: "owner", Operation: api.LLMKeyOperation_SAVE, Provider: &provider, Name: &name, ApiKey: &secret})
			if err != nil {
				t.Fatal(err)
			}
			if created.KeyId == "" || created.GetProvider() != provider {
				t.Fatal("invalid response")
			}
			record, err := store.Find(ctx, created.KeyId)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(record.Secret.Ciphertext), secret) {
				t.Fatal("plaintext stored")
			}
			replacement := "fake-replacement"
			_, err = h.ManageLLMKey(ctx, &api.ManageLLMKeyRequest{UserId: "other", Operation: api.LLMKeyOperation_UPDATE, KeyId: &created.KeyId, ApiKey: &replacement})
			assertStatus(t, err, 403)
			_, err = h.ManageLLMKey(ctx, &api.ManageLLMKeyRequest{UserId: "owner", Operation: api.LLMKeyOperation_UPDATE, KeyId: &created.KeyId, ApiKey: &replacement})
			if err != nil {
				t.Fatal(err)
			}
			used, err := gateway.Use(ctx, created.KeyId, "owner")
			if err != nil {
				t.Fatal(err)
			}
			if used.APIKey != replacement {
				t.Fatal("key update did not reach gateway")
			}
			_, err = h.ManageLLMKey(ctx, &api.ManageLLMKeyRequest{UserId: "owner", Operation: api.LLMKeyOperation_DELETE, KeyId: &created.KeyId})
			if err != nil {
				t.Fatal(err)
			}
			_, err = h.ManageLLMKey(ctx, &api.ManageLLMKeyRequest{UserId: "owner", Operation: api.LLMKeyOperation_DELETE, KeyId: &created.KeyId})
			assertStatus(t, err, 404)
		})
	}
}

func assertStatus(t *testing.T, err error, code int32) {
	t.Helper()
	status, ok := kerrors.FromBizStatusError(err)
	if !ok || status.BizStatusCode() != code {
		t.Fatalf("expected business status %d, got %v", code, err)
	}
}

func TestMasterKeyConfiguration(t *testing.T) {
	for _, value := range []string{"", "invalid", "ff"} {
		t.Setenv("LLM_GATEWAY_MASTER_KEY", value)
		if _, err := newMasterKey(); err == nil {
			t.Fatal("invalid key accepted")
		}
	}
	t.Setenv("LLM_GATEWAY_MASTER_KEY", strings.Repeat("ab", 32))
	if key, err := newMasterKey(); err != nil || len(key) != 32 {
		t.Fatal("valid key rejected")
	}
}
