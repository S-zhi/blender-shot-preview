package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"sync"

	llm_gateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/llmkeyservicev0_1"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/shotpreviewservicev0_1"
	"github.com/cloudwego/kitex/server"
)

type inMemoryCredentialStore struct {
	mu          sync.RWMutex
	credentials map[string]llm_gateway.CredentialRecord
}

func (s *inMemoryCredentialStore) Create(_ context.Context, record llm_gateway.CredentialRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.credentials[record.ID]; exists {
		return errors.New("credential already exists")
	}
	s.credentials[record.ID] = record
	return nil
}

func (s *inMemoryCredentialStore) Find(_ context.Context, keyID string) (llm_gateway.CredentialRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.credentials[keyID]
	if !ok {
		return llm_gateway.CredentialRecord{}, llm_gateway.ErrCredentialNotFound
	}
	return record, nil
}

func (s *inMemoryCredentialStore) Update(_ context.Context, record llm_gateway.CredentialRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.credentials[record.ID]; !exists {
		return llm_gateway.ErrCredentialNotFound
	}
	s.credentials[record.ID] = record
	return nil
}

func (s *inMemoryCredentialStore) Delete(_ context.Context, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.credentials[keyID]; !exists {
		return llm_gateway.ErrCredentialNotFound
	}
	delete(s.credentials, keyID)
	return nil
}

func newMasterKey() ([]byte, error) {
	encoded := os.Getenv("LLM_GATEWAY_MASTER_KEY")
	if encoded == "" {
		return nil, errors.New("LLM_GATEWAY_MASTER_KEY must be set to 64 hex characters")
	}
	key, err := hex.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, errors.New("LLM_GATEWAY_MASTER_KEY must be 64 hex characters")
	}
	return key, nil
}

func main() {
	masterKey, err := newMasterKey()
	if err != nil {
		log.Fatal(err)
	}
	cipher, err := llm_gateway.NewAESGCMCipher(masterKey)
	clear(masterKey)
	if err != nil {
		log.Fatal(err)
	}
	keyManager, err := llm_gateway.NewGateway(&inMemoryCredentialStore{credentials: make(map[string]llm_gateway.CredentialRecord)}, cipher)
	if err != nil {
		log.Fatal(err)
	}

	svr, err := newServer(keyManager)
	if err != nil {
		log.Fatal(err)
	}

	if err := svr.Run(); err != nil {
		log.Fatal(err)
	}
}

func newServer(manager llm_gateway.KeyManager) (server.Server, error) {
	handler := handlerv0_1.NewShotPreviewHandler(service.NewShotPreviewService(nil))
	keyHandler := handlerv0_1.NewLLMKeyHandler(service.NewLLMKeyService(manager))
	svr := shotpreviewservicev0_1.NewServer(handler)
	if err := llmkeyservicev0_1.RegisterService(svr, keyHandler); err != nil {
		return nil, err
	}
	return svr, nil
}
