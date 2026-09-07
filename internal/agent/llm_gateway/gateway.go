// Package llmgateway manages provider credentials.
// It deliberately does not make model requests or depend on an agent package.
package llmgateway

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
)

type KeyCommandType string

const (
	KeyCommandSave   KeyCommandType = "save"
	KeyCommandUpdate KeyCommandType = "update"
	KeyCommandDelete KeyCommandType = "delete"
)

var (
	ErrInvalidCommand      = errors.New("invalid credential command")
	ErrCredentialNotFound  = errors.New("credential not found")
	ErrCredentialForbidden = errors.New("credential access forbidden")
)

// KeyCommand groups all credential mutations behind one command interface.
// Save requires Provider, UserID, Name, and APIKey. Update modifies only
// non-nil mutable fields. Delete requires KeyID and UserID.
type KeyCommand struct {
	Type     KeyCommandType
	KeyID    string
	UserID   string
	Provider Provider
	Name     *string
	APIKey   *string
	BaseURL  *string
}

type KeyCommandResult struct {
	KeyID string
}

// KeyManager is the command-side interface for credential lifecycle changes.
type KeyManager interface {
	Execute(ctx context.Context, command KeyCommand) (KeyCommandResult, error)
}

// KeyUser is the query-side interface used by later callers.
// The returned API key is plaintext only in process memory and must not be logged.
type KeyUser interface {
	Use(ctx context.Context, keyID, userID string) (UsableKey, error)
}

// Gateway is intentionally the only public implementation boundary needed by
// callers. It implements both KeyManager and KeyUser.
type Gateway struct {
	store    CredentialStore
	cipher   SecretCipher
	adapters map[Provider]providerAdapter
	now      func() time.Time
	newID    func() (string, error)
}

func NewGateway(store CredentialStore, secretCipher SecretCipher) (*Gateway, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: credential store is required", ErrInvalidCommand)
	}
	if secretCipher == nil {
		return nil, fmt.Errorf("%w: secret cipher is required", ErrInvalidCommand)
	}

	return &Gateway{
		store:  store,
		cipher: secretCipher,
		adapters: map[Provider]providerAdapter{
			ProviderOpenAI:    openAIAdapter{},
			ProviderAnthropic: anthropicAdapter{},
		},
		now:   time.Now,
		newID: newCredentialID,
	}, nil
}

func (g *Gateway) Execute(ctx context.Context, command KeyCommand) (KeyCommandResult, error) {
	if err := validateUserID(command.UserID); err != nil {
		return KeyCommandResult{}, err
	}
	command.UserID = strings.TrimSpace(command.UserID)
	command.KeyID = strings.TrimSpace(command.KeyID)

	switch command.Type {
	case KeyCommandSave:
		return g.save(ctx, command)
	case KeyCommandUpdate:
		return g.update(ctx, command)
	case KeyCommandDelete:
		return g.delete(ctx, command)
	default:
		return KeyCommandResult{}, fmt.Errorf("%w: unsupported command type %q", ErrInvalidCommand, command.Type)
	}
}

func (g *Gateway) Use(ctx context.Context, keyID, userID string) (UsableKey, error) {
	if err := validateUserID(userID); err != nil {
		return UsableKey{}, err
	}
	userID = strings.TrimSpace(userID)
	keyID = strings.TrimSpace(keyID)
	if strings.TrimSpace(keyID) == "" {
		return UsableKey{}, fmt.Errorf("%w: key ID is required", ErrInvalidCommand)
	}

	record, err := g.store.Find(ctx, keyID)
	if err != nil {
		return UsableKey{}, err
	}
	if record.UserID != userID {
		return UsableKey{}, ErrCredentialForbidden
	}
	if !record.Enabled {
		return UsableKey{}, ErrCredentialNotFound
	}

	adapter, err := g.adapter(record.Provider)
	if err != nil {
		return UsableKey{}, err
	}
	secret, err := g.cipher.Decrypt(record.Secret, credentialAAD(record))
	if err != nil {
		return UsableKey{}, fmt.Errorf("decrypt credential %q: %w", keyID, err)
	}
	defer clear(secret)

	return adapter.usableKey(record, string(secret))
}

func (g *Gateway) save(ctx context.Context, command KeyCommand) (KeyCommandResult, error) {
	adapter, err := g.adapter(command.Provider)
	if err != nil {
		return KeyCommandResult{}, err
	}
	name, apiKey, baseURL, err := validateSaveCommand(command, adapter)
	if err != nil {
		return KeyCommandResult{}, err
	}

	keyID, err := g.newID()
	if err != nil {
		return KeyCommandResult{}, fmt.Errorf("create credential ID: %w", err)
	}
	now := g.now().UTC()
	record := CredentialRecord{
		ID:         keyID,
		UserID:     command.UserID,
		Provider:   command.Provider,
		Name:       name,
		BaseURL:    baseURL,
		Enabled:    true,
		KeyVersion: 1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	record.Secret, err = g.cipher.Encrypt([]byte(apiKey), credentialAAD(record))
	if err != nil {
		return KeyCommandResult{}, fmt.Errorf("encrypt credential: %w", err)
	}
	if err := g.store.Create(ctx, record); err != nil {
		return KeyCommandResult{}, err
	}
	return KeyCommandResult{KeyID: keyID}, nil
}

func (g *Gateway) update(ctx context.Context, command KeyCommand) (KeyCommandResult, error) {
	if strings.TrimSpace(command.KeyID) == "" {
		return KeyCommandResult{}, fmt.Errorf("%w: key ID is required", ErrInvalidCommand)
	}
	if command.Name == nil && command.APIKey == nil && command.BaseURL == nil {
		return KeyCommandResult{}, fmt.Errorf("%w: update requires at least one field", ErrInvalidCommand)
	}

	record, err := g.store.Find(ctx, command.KeyID)
	if err != nil {
		return KeyCommandResult{}, err
	}
	if record.UserID != command.UserID {
		return KeyCommandResult{}, ErrCredentialForbidden
	}
	if command.Provider != "" && command.Provider != record.Provider {
		return KeyCommandResult{}, fmt.Errorf("%w: provider cannot be changed", ErrInvalidCommand)
	}
	adapter, err := g.adapter(record.Provider)
	if err != nil {
		return KeyCommandResult{}, err
	}

	if command.Name != nil {
		name := strings.TrimSpace(*command.Name)
		if name == "" {
			return KeyCommandResult{}, fmt.Errorf("%w: name is required", ErrInvalidCommand)
		}
		record.Name = name
	}
	if command.BaseURL != nil {
		baseURL, err := adapter.normalizeBaseURL(*command.BaseURL)
		if err != nil {
			return KeyCommandResult{}, err
		}
		record.BaseURL = baseURL
	}
	if command.APIKey != nil {
		apiKey := strings.TrimSpace(*command.APIKey)
		if err := adapter.validateAPIKey(apiKey); err != nil {
			return KeyCommandResult{}, err
		}
		record.KeyVersion++
		record.Secret, err = g.cipher.Encrypt([]byte(apiKey), credentialAAD(record))
		if err != nil {
			return KeyCommandResult{}, fmt.Errorf("encrypt credential: %w", err)
		}
	}
	record.UpdatedAt = g.now().UTC()
	if err := g.store.Update(ctx, record); err != nil {
		return KeyCommandResult{}, err
	}
	return KeyCommandResult{KeyID: record.ID}, nil
}

func (g *Gateway) delete(ctx context.Context, command KeyCommand) (KeyCommandResult, error) {
	if strings.TrimSpace(command.KeyID) == "" {
		return KeyCommandResult{}, fmt.Errorf("%w: key ID is required", ErrInvalidCommand)
	}
	record, err := g.store.Find(ctx, command.KeyID)
	if err != nil {
		return KeyCommandResult{}, err
	}
	if record.UserID != command.UserID {
		return KeyCommandResult{}, ErrCredentialForbidden
	}
	if err := g.store.Delete(ctx, record.ID); err != nil {
		return KeyCommandResult{}, err
	}
	return KeyCommandResult{KeyID: record.ID}, nil
}

func (g *Gateway) adapter(provider Provider) (providerAdapter, error) {
	adapter, ok := g.adapters[provider]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported provider %q", ErrInvalidCommand, provider)
	}
	return adapter, nil
}

// CredentialStore is the persistence boundary. The production implementation
// can use PostgreSQL without changing gateway or provider adapter code.
type CredentialStore interface {
	Create(ctx context.Context, credential CredentialRecord) error
	Find(ctx context.Context, keyID string) (CredentialRecord, error)
	Update(ctx context.Context, credential CredentialRecord) error
	Delete(ctx context.Context, keyID string) error
}

type CredentialRecord struct {
	ID         string
	UserID     string
	Provider   Provider
	Name       string
	BaseURL    string
	Secret     EncryptedSecret
	KeyVersion uint32
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type EncryptedSecret struct {
	Ciphertext []byte
	Nonce      []byte
}

// UsableKey contains a decrypted credential for a single in-process use.
// Callers must treat APIKey and Headers as sensitive values.
type UsableKey struct {
	KeyID    string
	Provider Provider
	BaseURL  string
	APIKey   string
	Headers  map[string]string
}

type providerAdapter interface {
	provider() Provider
	normalizeBaseURL(baseURL string) (string, error)
	validateAPIKey(apiKey string) error
	usableKey(record CredentialRecord, apiKey string) (UsableKey, error)
}

// SecretCipher allows a secret manager or KMS-backed implementation to replace
// AES-GCM later without changing the command or use interfaces.
type SecretCipher interface {
	Encrypt(plaintext, aad []byte) (EncryptedSecret, error)
	Decrypt(secret EncryptedSecret, aad []byte) ([]byte, error)
}

type aesGCMCipher struct {
	aead cipher.AEAD
}

// NewAESGCMCipher creates the first-version credential encryption mechanism.
// The supplied master key must be exactly 32 bytes and must come from runtime
// configuration or a secret manager, never from the database.
func NewAESGCMCipher(masterKey []byte) (SecretCipher, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("AES-256-GCM master key must be 32 bytes, got %d", len(masterKey))
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesGCMCipher{aead: aead}, nil
}

func (c aesGCMCipher) Encrypt(plaintext, aad []byte) (EncryptedSecret, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return EncryptedSecret{}, err
	}
	return EncryptedSecret{
		Ciphertext: c.aead.Seal(nil, nonce, plaintext, aad),
		Nonce:      nonce,
	}, nil
}

func (c aesGCMCipher) Decrypt(secret EncryptedSecret, aad []byte) ([]byte, error) {
	if len(secret.Nonce) != c.aead.NonceSize() {
		return nil, errors.New("invalid credential nonce")
	}
	return c.aead.Open(nil, secret.Nonce, secret.Ciphertext, aad)
}

func validateSaveCommand(command KeyCommand, adapter providerAdapter) (name, apiKey, baseURL string, err error) {
	if command.Name == nil || command.APIKey == nil {
		return "", "", "", fmt.Errorf("%w: save requires name and API key", ErrInvalidCommand)
	}
	name = strings.TrimSpace(*command.Name)
	if name == "" {
		return "", "", "", fmt.Errorf("%w: name is required", ErrInvalidCommand)
	}
	apiKey = strings.TrimSpace(*command.APIKey)
	if err := adapter.validateAPIKey(apiKey); err != nil {
		return "", "", "", err
	}
	if command.BaseURL != nil {
		baseURL, err = adapter.normalizeBaseURL(*command.BaseURL)
	} else {
		baseURL, err = adapter.normalizeBaseURL("")
	}
	return name, apiKey, baseURL, err
}

func validateUserID(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("%w: user ID is required", ErrInvalidCommand)
	}
	return nil
}

func normalizeHTTPURL(value, defaultURL string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultURL
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%w: base URL must be absolute", ErrInvalidCommand)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%w: base URL scheme must be HTTP or HTTPS", ErrInvalidCommand)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: base URL cannot include query or fragment", ErrInvalidCommand)
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func credentialAAD(record CredentialRecord) []byte {
	return []byte(strings.Join([]string{
		record.ID,
		record.UserID,
		string(record.Provider),
		fmt.Sprintf("%d", record.KeyVersion),
	}, ":"))
}

func newCredentialID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
