package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	llm_gateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	"github.com/S-zhi/blender-shot-preview/internal/gateway"
	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/assetservicev0_1"
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
		return nil, errors.New("LLM_GATEWAY_MASTER_KEY must be 64 hex characters")
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
		// Automatically generate a 32-byte master key for local development
		masterKey = make([]byte, 32)
		if _, randErr := rand.Read(masterKey); randErr != nil {
			log.Fatal(randErr)
		}
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

	// 1. Initialize shared services & handlers
	shotPreviewSvc := service.NewShotPreviewService(nil)
	llmKeySvc := service.NewLLMKeyService(keyManager)
	assetSvc := service.NewAssetService(nil)

	shotHandler := handlerv0_1.NewShotPreviewHandler(shotPreviewSvc)
	keyHandler := handlerv0_1.NewLLMKeyHandler(llmKeySvc)
	assetHandler := handlerv0_1.NewAssetHandler(assetSvc)

	// 2. Start Kitex Thrift RPC Server on 127.0.0.1:8889 in background
	kitexAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8889")
	kitexSvr := shotpreviewservicev0_1.NewServer(shotHandler, server.WithServiceAddr(kitexAddr))
	if err := llmkeyservicev0_1.RegisterService(kitexSvr, keyHandler); err != nil {
		log.Fatalf("failed to register LLMKeyService: %v", err)
	}
	if err := assetservicev0_1.RegisterService(kitexSvr, assetHandler); err != nil {
		log.Fatalf("failed to register AssetService: %v", err)
	}

	go func() {
		log.Println("[Kitex RPC] Server listening on 127.0.0.1:8889")
		if err := kitexSvr.Run(); err != nil {
			log.Printf("[Kitex RPC] Server error: %v", err)
		}
	}()

	// 3. Start HTTP Gateway on 127.0.0.1:8888 for Web Frontend
	httpGateway := gateway.NewHTTPGateway(shotHandler, keyHandler, assetHandler)
	httpServer := &http.Server{
		Addr:    "127.0.0.1:8888",
		Handler: httpGateway,
	}

	log.Println("[HTTP Gateway] API Server listening on http://127.0.0.1:8888")
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[HTTP Gateway] ListenAndServe error: %v", err)
	}
}

func newServer(manager llm_gateway.KeyManager) (server.Server, error) {
	shotHandler := handlerv0_1.NewShotPreviewHandler(service.NewShotPreviewService(nil))
	keyHandler := handlerv0_1.NewLLMKeyHandler(service.NewLLMKeyService(manager))
	assetHandler := handlerv0_1.NewAssetHandler(service.NewAssetService(nil))
	svr := shotpreviewservicev0_1.NewServer(shotHandler)
	if err := llmkeyservicev0_1.RegisterService(svr, keyHandler); err != nil {
		return nil, err
	}
	if err := assetservicev0_1.RegisterService(svr, assetHandler); err != nil {
		return nil, err
	}
	return svr, nil
}
