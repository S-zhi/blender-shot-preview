package llmgateway

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGatewayCredentialLifecycle(t *testing.T) {
	t.Parallel()

	store := newMemoryCredentialStore()
	gateway := newTestGateway(t, store)
	name := "primary OpenAI"
	apiKey := "secret-before-update"
	created, err := gateway.Execute(context.Background(), KeyCommand{
		Type:     KeyCommandSave,
		UserID:   "user-a",
		Provider: ProviderOpenAI,
		Name:     &name,
		APIKey:   &apiKey,
	})
	if err != nil {
		t.Fatalf("save credential: %v", err)
	}
	if created.KeyID == "" {
		t.Fatal("save credential returned empty key ID")
	}

	record, err := store.Find(context.Background(), created.KeyID)
	if err != nil {
		t.Fatalf("find stored credential: %v", err)
	}
	if strings.Contains(string(record.Secret.Ciphertext), apiKey) {
		t.Fatal("stored ciphertext contains plaintext API key")
	}
	if record.BaseURL != defaultOpenAIBaseURL {
		t.Fatalf("OpenAI base URL = %q, want %q", record.BaseURL, defaultOpenAIBaseURL)
	}

	used, err := gateway.Use(context.Background(), created.KeyID, "user-a")
	if err != nil {
		t.Fatalf("use saved credential: %v", err)
	}
	if used.APIKey != apiKey {
		t.Fatalf("used API key = %q, want %q", used.APIKey, apiKey)
	}
	if used.Headers["Authorization"] != "Bearer "+apiKey {
		t.Fatalf("OpenAI authorization header = %q", used.Headers["Authorization"])
	}

	updatedKey := "secret-after-update"
	updatedName := "rotated OpenAI"
	if _, err := gateway.Execute(context.Background(), KeyCommand{
		Type:   KeyCommandUpdate,
		KeyID:  created.KeyID,
		UserID: "user-a",
		Name:   &updatedName,
		APIKey: &updatedKey,
	}); err != nil {
		t.Fatalf("update credential: %v", err)
	}

	used, err = gateway.Use(context.Background(), created.KeyID, "user-a")
	if err != nil {
		t.Fatalf("use updated credential: %v", err)
	}
	if used.APIKey != updatedKey {
		t.Fatalf("updated API key = %q, want %q", used.APIKey, updatedKey)
	}

	if _, err := gateway.Execute(context.Background(), KeyCommand{
		Type:   KeyCommandDelete,
		KeyID:  created.KeyID,
		UserID: "user-a",
	}); err != nil {
		t.Fatalf("delete credential: %v", err)
	}
	if _, err := gateway.Use(context.Background(), created.KeyID, "user-a"); !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("use deleted credential error = %v, want ErrCredentialNotFound", err)
	}
}

func TestGatewayRejectsAnotherUser(t *testing.T) {
	t.Parallel()

	gateway := newTestGateway(t, newMemoryCredentialStore())
	name, apiKey := "Anthropic key", "anthropic-secret"
	created, err := gateway.Execute(context.Background(), KeyCommand{
		Type:     KeyCommandSave,
		UserID:   "owner",
		Provider: ProviderAnthropic,
		Name:     &name,
		APIKey:   &apiKey,
	})
	if err != nil {
		t.Fatalf("save credential: %v", err)
	}

	if _, err := gateway.Use(context.Background(), created.KeyID, "other-user"); !errors.Is(err, ErrCredentialForbidden) {
		t.Fatalf("use another user's credential error = %v, want ErrCredentialForbidden", err)
	}
	if _, err := gateway.Execute(context.Background(), KeyCommand{
		Type:   KeyCommandDelete,
		KeyID:  created.KeyID,
		UserID: "other-user",
	}); !errors.Is(err, ErrCredentialForbidden) {
		t.Fatalf("delete another user's credential error = %v, want ErrCredentialForbidden", err)
	}
}

func TestAnthropicUsableKey(t *testing.T) {
	t.Parallel()

	gateway := newTestGateway(t, newMemoryCredentialStore())
	name, apiKey := "Anthropic key", "anthropic-secret"
	created, err := gateway.Execute(context.Background(), KeyCommand{
		Type:     KeyCommandSave,
		UserID:   "user-a",
		Provider: ProviderAnthropic,
		Name:     &name,
		APIKey:   &apiKey,
	})
	if err != nil {
		t.Fatalf("save credential: %v", err)
	}

	used, err := gateway.Use(context.Background(), created.KeyID, "user-a")
	if err != nil {
		t.Fatalf("use credential: %v", err)
	}
	if used.BaseURL != defaultAnthropicBaseURL {
		t.Fatalf("Anthropic base URL = %q, want %q", used.BaseURL, defaultAnthropicBaseURL)
	}
	if used.Headers["x-api-key"] != apiKey {
		t.Fatalf("Anthropic x-api-key header = %q", used.Headers["x-api-key"])
	}
	if used.Headers["anthropic-version"] != anthropicAPIVersion {
		t.Fatalf("Anthropic version header = %q", used.Headers["anthropic-version"])
	}
}

func TestAESGCMCipherBindsCredentialMetadata(t *testing.T) {
	t.Parallel()

	cipher, err := NewAESGCMCipher([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	secret, err := cipher.Encrypt([]byte("api-key"), []byte("credential-a"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := cipher.Decrypt(secret, []byte("credential-b")); err == nil {
		t.Fatal("decrypt with different AAD succeeded")
	}
}

func newTestGateway(t *testing.T, store CredentialStore) *Gateway {
	t.Helper()
	cipher, err := NewAESGCMCipher([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	gateway, err := NewGateway(store, cipher)
	if err != nil {
		t.Fatalf("new gateway: %v", err)
	}
	return gateway
}

type memoryCredentialStore struct {
	credentials map[string]CredentialRecord
}

func newMemoryCredentialStore() *memoryCredentialStore {
	return &memoryCredentialStore{credentials: make(map[string]CredentialRecord)}
}

func (s *memoryCredentialStore) Create(_ context.Context, credential CredentialRecord) error {
	if _, exists := s.credentials[credential.ID]; exists {
		return errors.New("credential already exists")
	}
	s.credentials[credential.ID] = cloneCredential(credential)
	return nil
}

func (s *memoryCredentialStore) Find(_ context.Context, keyID string) (CredentialRecord, error) {
	credential, ok := s.credentials[keyID]
	if !ok {
		return CredentialRecord{}, ErrCredentialNotFound
	}
	return cloneCredential(credential), nil
}

func (s *memoryCredentialStore) Update(_ context.Context, credential CredentialRecord) error {
	if _, exists := s.credentials[credential.ID]; !exists {
		return ErrCredentialNotFound
	}
	s.credentials[credential.ID] = cloneCredential(credential)
	return nil
}

func (s *memoryCredentialStore) Delete(_ context.Context, keyID string) error {
	if _, exists := s.credentials[keyID]; !exists {
		return ErrCredentialNotFound
	}
	delete(s.credentials, keyID)
	return nil
}

func cloneCredential(credential CredentialRecord) CredentialRecord {
	credential.Secret.Ciphertext = append([]byte(nil), credential.Secret.Ciphertext...)
	credential.Secret.Nonce = append([]byte(nil), credential.Secret.Nonce...)
	return credential
}
