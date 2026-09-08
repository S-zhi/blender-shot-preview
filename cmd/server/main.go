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
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
	llm_gateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	"github.com/S-zhi/blender-shot-preview/internal/gateway"
	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/productiontools"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/assetservicev0_1"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/llmkeyservicev0_1"
	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1/shotpreviewservicev0_1"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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

func (s *inMemoryCredentialStore) FindActiveByUser(_ context.Context, userID string) (llm_gateway.CredentialRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var selected llm_gateway.CredentialRecord
	for _, record := range s.credentials {
		if record.UserID != userID || !record.Enabled {
			continue
		}
		if selected.ID == "" || record.CreatedAt.After(selected.CreatedAt) {
			selected = record
		}
	}
	if selected.ID == "" {
		return llm_gateway.CredentialRecord{}, llm_gateway.ErrCredentialNotFound
	}
	return selected, nil
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
	if os.Getenv("LLM_GATEWAY_MASTER_KEY") == "" {
		// Automatically generate a 64-hex character master key for local development
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			log.Fatal(err)
		}
		_ = os.Setenv("LLM_GATEWAY_MASTER_KEY", hex.EncodeToString(key))
	}

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

	// 1. Initialize shared services & handlers
	runner := buildPipelineRunner(keyManager)
	shotPreviewSvc := service.NewShotPreviewService(runner)
	llmKeySvc := service.NewLLMKeyService(keyManager)
	assetSvc := service.NewAssetService(nil)

	shotHandler := handlerv0_1.NewShotPreviewHandler(shotPreviewSvc)
	keyHandler := handlerv0_1.NewLLMKeyHandler(llmKeySvc)
	assetHandler := handlerv0_1.NewAssetHandler(assetSvc)

	// 2. Start Kitex Thrift RPC Server on 127.0.0.1:8889 in background
	kitexAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8889")
	kitexSvr, err := newServerWithHandlers(shotHandler, keyHandler, assetHandler, server.WithServiceAddr(kitexAddr))
	if err != nil {
		log.Fatalf("failed to initialize Kitex server: %v", err)
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

func newServer(manager llm_gateway.KeyManager, opts ...server.Option) (server.Server, error) {
	var keyUser llm_gateway.KeyUser
	if ku, ok := manager.(llm_gateway.KeyUser); ok {
		keyUser = ku
	}
	runner := buildPipelineRunner(keyUser)
	shotPreviewSvc := service.NewShotPreviewService(runner)
	shotHandler := handlerv0_1.NewShotPreviewHandler(shotPreviewSvc)
	keyHandler := handlerv0_1.NewLLMKeyHandler(service.NewLLMKeyService(manager))
	assetHandler := handlerv0_1.NewAssetHandler(service.NewAssetService(nil))
	return newServerWithHandlers(shotHandler, keyHandler, assetHandler, opts...)
}

func newServerWithHandlers(
	shotHandler *handlerv0_1.ShotPreviewHandler,
	keyHandler *handlerv0_1.LLMKeyHandler,
	assetHandler *handlerv0_1.AssetHandler,
	opts ...server.Option,
) (server.Server, error) {
	svr := shotpreviewservicev0_1.NewServer(shotHandler, opts...)
	if err := llmkeyservicev0_1.RegisterService(svr, keyHandler); err != nil {
		return nil, err
	}
	if err := assetservicev0_1.RegisterService(svr, assetHandler); err != nil {
		return nil, err
	}
	return svr, nil
}

func buildPipelineRunner(keyUser llm_gateway.KeyUser) *pipeline.Runner {
	workspaceDir := os.Getenv("BLENDER_WORKSPACE")
	if workspaceDir == "" {
		workspaceDir = filepath.Join(os.TempDir(), "blender-shot-preview")
	}
	_ = os.MkdirAll(workspaceDir, 0755)
	_ = os.MkdirAll(filepath.Join(workspaceDir, "tasks"), 0755)
	_ = os.MkdirAll(filepath.Join(workspaceDir, "assets"), 0755)
	_ = os.WriteFile(filepath.Join(workspaceDir, "tasks", "scene.blend"), []byte("BLENDER DEV PLACEHOLDER"), 0644)
	_ = os.WriteFile(filepath.Join(workspaceDir, "assets", "asset-1.blend"), []byte("BLENDER DEV PLACEHOLDER"), 0644)

	blenderBin := os.Getenv("BLENDER_BINARY")
	if blenderBin == "" {
		if _, err := exec.LookPath("blender"); err != nil {
			if _, err := os.Stat("/Applications/Blender.app/Contents/MacOS/Blender"); err == nil {
				blenderBin = "/Applications/Blender.app/Contents/MacOS/Blender"
			}
		}
	}

	skillRegistry := agent.NewMemorySkillRegistry()
	_, _ = productiontools.RegisterProductionTools(skillRegistry, productiontools.Config{
		Workspace:     workspaceDir,
		BlenderBinary: blenderBin,
		Executor:      devCommandExecutor{wsRoot: workspaceDir},
	})

	repo := pipeline.NewMemoryRepository()
	defRepo := agent.NewMemoryDefinitionRepository()
	runRepo := agent.NewMemoryRunRepository()
	devMode := strings.EqualFold(strings.TrimSpace(os.Getenv("SHOT_PREVIEW_DEV_MODE")), "true")
	modelResolver := agent.NewStaticModelResolver(map[string]model.BaseChatModel{
		agent.ProductionModelProfile: &devChatModel{},
	})
	factory, err := agent.NewEinoAgentFactory(modelResolver, skillRegistry)
	if err != nil {
		return nil
	}
	defService, err := agent.NewDefinitionService(defRepo, skillRegistry, modelResolver)
	if err != nil {
		return nil
	}
	ctx := context.Background()
	for _, def := range agent.ProductionAgentDefinitions() {
		_, _ = defService.Create(ctx, agent.CreateAgentRequest{
			ID:              def.ID,
			Name:            def.Name,
			Description:     def.Description,
			SystemPrompt:    def.SystemPrompt,
			ModelProfile:    def.ModelProfile,
			SkillIDs:        def.SkillIDs,
			InputSchema:     def.InputSchema,
			OutputSchema:    def.OutputSchema,
			MaxSteps:        def.MaxSteps,
			MaxOutputTokens: def.MaxOutputTokens,
			Timeout:         def.Timeout,
		})
	}
	agentSvc, err := agent.NewService(defRepo, runRepo, factory)
	if err != nil {
		return nil
	}

	if devMode {
		// Development mode is explicit and intentionally bypasses credential
		// resolution so the deterministic devChatModel remains usable in tests.
		return pipeline.NewShotPreviewRunner(repo, agentSvc, skillRegistry)
	}
	return pipeline.NewShotPreviewRunnerWithKeys(repo, agentSvc, skillRegistry, keyUser)
}

type devChatModel struct{}

func (m *devChatModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	for _, msg := range input {
		content := msg.Content
		if strings.Contains(content, "# Scene planning") {
			return schema.AssistantMessage(`{"version":"scene-plan/v1","scene_id":"scene-001","asset_tasks":[{"version":"asset-task/v1","task_id":"task-1","asset_id":"asset-1","asset_kind":"model","description":"landscape model","workspace":"tasks","output_path":"assets/asset-1.blend","depends_on":[]}],"environment_plan":{},"action_plan":[]}`, nil), nil
		}
		if strings.Contains(content, "# Editable asset production") {
			return schema.AssistantMessage(`{"version":"asset-manifest/v1","asset_id":"asset-1","asset_kind":"model","source_files":["assets/asset-1.blend"],"blend_file":"assets/asset-1.blend","inspection":{"passed":true}}`, nil), nil
		}
		if strings.Contains(content, "# Shot design") {
			return schema.AssistantMessage(`{"version":"shot-plan/v1","scene_id":"scene-001","shots":[{"shot_id":"shot-1","order":1,"camera":{"type":"perspective"},"action_ids":[],"asset_ids":["asset-1"]}],"constraints":[]}`, nil), nil
		}
		if strings.Contains(content, "# Editable scene assembly") {
			return schema.AssistantMessage(`{"version":"scene-assembly-result/v1","scene_id":"scene-001","blend_file":"tasks/scene.blend","inspection":{"passed":true}}`, nil), nil
		}
	}
	return schema.AssistantMessage(`{"version":"scene-spec/v1","scene_id":"scene-001","summary":"preview shot","style":{"visual_style":"realistic","mood":"peaceful"},"environment":{"location":"studio","lighting":"studio"},"assets":[],"actions":[],"shots":[]}`, nil), nil
}

func (m *devChatModel) Stream(ctx context.Context, input []*schema.Message, options ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, options...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func (m *devChatModel) BindTools(_ []*schema.ToolInfo) error { return nil }

type devCommandExecutor struct {
	wsRoot string
}

func (e devCommandExecutor) Run(ctx context.Context, c productiontools.Command) (productiontools.CommandResult, error) {
	if strings.Contains(c.Name, "ffmpeg") {
		if len(c.Args) > 0 {
			outPath := c.Args[len(c.Args)-1]
			if !filepath.IsAbs(outPath) {
				outPath = filepath.Join(c.Dir, outPath)
			}
			_ = os.MkdirAll(filepath.Dir(outPath), 0755)
			_ = os.WriteFile(outPath, []byte("fake mp4 content"), 0644)
		}
		return productiontools.CommandResult{ExitCode: 0}, nil
	}
	if strings.Contains(c.Name, "ffprobe") {
		return productiontools.CommandResult{
			Stdout:   []byte(`{"format":{"duration":"5.0"},"streams":[{"codec_name":"h264"}]}`),
			ExitCode: 0,
		}, nil
	}
	return (productiontools.OSCommandExecutor{}).Run(ctx, c)
}

func (e devCommandExecutor) Start(ctx context.Context, c productiontools.Command) (productiontools.CommandProcess, error) {
	for i, arg := range c.Args {
		if arg == "--render-output" && i+1 < len(c.Args) {
			prefix := c.Args[i+1]
			if !filepath.IsAbs(prefix) {
				prefix = filepath.Join(c.Dir, prefix)
			}
			_ = os.MkdirAll(filepath.Dir(prefix), 0755)
			_ = os.WriteFile(prefix+"0001.png", []byte("fake frame 1 png"), 0644)
			_ = os.WriteFile(prefix+"0250.png", []byte("fake frame 250 png"), 0644)
		}
	}
	return devProcess{}, nil
}

type devProcess struct{}

func (devProcess) Wait() (productiontools.CommandResult, error) {
	return productiontools.CommandResult{ExitCode: 0}, nil
}
func (devProcess) Kill() error { return nil }
