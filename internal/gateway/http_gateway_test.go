package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

func TestHTTPGateway_Routes(t *testing.T) {
	assetSvc := service.NewAssetService(service.NewInMemoryAssetStore())
	assetHandler := handlerv0_1.NewAssetHandler(assetSvc)

	gw := NewHTTPGateway(nil, nil, assetHandler)

	// 1. Test /health
	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	wHealth := httptest.NewRecorder()
	gw.ServeHTTP(wHealth, reqHealth)
	if wHealth.Code != http.StatusOK {
		t.Errorf("expected 200 from /health, got %d", wHealth.Code)
	}

	// 2. Test POST /api/v0_1/assets (Register asset first)
	newAsset := api.RegisterAssetRequest{
		UserId:        "default_user_001",
		Name:          "http_test_model.blend",
		AssetType:     api.AssetType_MODEL_3D,
		FileFormat:    "blend",
		FileSizeBytes: 1024,
		StorageUri:    "blender://test",
	}
	body, _ := json.Marshal(newAsset)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v0_1/assets", bytes.NewReader(body))
	wPost := httptest.NewRecorder()
	gw.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Errorf("expected 200 from POST /api/v0_1/assets, got %d: %s", wPost.Code, wPost.Body.String())
	}

	// 3. Test GET /api/v0_1/assets (Verify registered asset is returned)
	reqList := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets?user_id=default_user_001", nil)
	wList := httptest.NewRecorder()
	gw.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/v0_1/assets, got %d: %s", wList.Code, wList.Body.String())
	}
	var listRes api.ListAssetsResponse
	if err := json.Unmarshal(wList.Body.Bytes(), &listRes); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if len(listRes.Assets) != 1 || listRes.Assets[0].Name != "http_test_model.blend" {
		t.Errorf("expected 1 registered asset from gateway list, got %d", len(listRes.Assets))
	}

	// 4. Test GET /api/v0_1/assets/stats
	reqStats := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets/stats?user_id=default_user_001", nil)
	wStats := httptest.NewRecorder()
	gw.ServeHTTP(wStats, reqStats)
	if wStats.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/v0_1/assets/stats, got %d", wStats.Code)
	}

	// 5. Test POST /api/v0_1/assets/upload (multipart .blend upload)
	var b bytes.Buffer
	wUploadWriter := multipart.NewWriter(&b)
	part, err := wUploadWriter.CreateFormFile("file", "heroine_alita_rigged.blend")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("BLENDER_v401\x00Armature_Rigify\x00Bone_head\x00"))
	_ = wUploadWriter.WriteField("user_id", "default_user_001")
	_ = wUploadWriter.WriteField("name", "heroine_alita_rigged.blend")
	_ = wUploadWriter.Close()

	reqUpload := httptest.NewRequest(http.MethodPost, "/api/v0_1/assets/upload", &b)
	reqUpload.Header.Set("Content-Type", wUploadWriter.FormDataContentType())
	wUpload := httptest.NewRecorder()
	gw.ServeHTTP(wUpload, reqUpload)
	if wUpload.Code != http.StatusOK {
		t.Errorf("expected 200 from POST /api/v0_1/assets/upload, got %d: %s", wUpload.Code, wUpload.Body.String())
	}
}

func TestHTTPGateway_ShotPreviewTask(t *testing.T) {
	stub := &stubShotPreviewService{
		task: service.TaskView{
			TaskID:          "task-test-1",
			Status:          service.TaskStatusRunning,
			WorkflowID:      service.ShotPreviewWorkflowID,
			WorkflowVersion: service.ShotPreviewWorkflowVersion,
			Nodes: []service.NodeView{
				{ID: "Initialize", Status: service.NodeStatusSucceeded},
			},
		},
	}
	shotHandler := handlerv0_1.NewShotPreviewHandler(stub)
	gw := NewHTTPGateway(shotHandler, nil, nil)

	// 1. Test POST /api/v0_1/shot-preview/task
	createReq := api.CreateShotPreviewTaskRequest{
		UserId: "default_user_001",
		Prompt: "establish shot of mountains",
	}
	createBody, _ := json.Marshal(createReq)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v0_1/shot-preview/task", bytes.NewReader(createBody))
	wPost := httptest.NewRecorder()
	gw.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Fatalf("expected 200 from POST /api/v0_1/shot-preview/task, got %d: %s", wPost.Code, wPost.Body.String())
	}
	var createRes api.CreateShotPreviewTaskResponse
	if err := json.Unmarshal(wPost.Body.Bytes(), &createRes); err != nil {
		t.Fatalf("failed to decode create task response: %v", err)
	}
	if createRes.TaskId != "task-test-1" {
		t.Fatalf("expected task-test-1, got %s", createRes.TaskId)
	}

	// 2. Test GET /api/v0_1/shot-preview/task?task_id=task-test-1&user_id=default_user_001
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v0_1/shot-preview/task?task_id=task-test-1&user_id=default_user_001", nil)
	wGet := httptest.NewRecorder()
	gw.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("expected 200 from GET /api/v0_1/shot-preview/task, got %d: %s", wGet.Code, wGet.Body.String())
	}
	var getRes api.GetShotPreviewTaskResponse
	if err := json.Unmarshal(wGet.Body.Bytes(), &getRes); err != nil {
		t.Fatalf("failed to decode get task response: %v", err)
	}
	if getRes.Task == nil || getRes.Task.TaskId != "task-test-1" {
		t.Fatalf("expected task-test-1 in get response, got %+v", getRes.Task)
	}
}

type stubShotPreviewService struct {
	task service.TaskView
}

func (s *stubShotPreviewService) CreateTask(_ context.Context, req service.CreateTaskRequest) (service.CreateTaskResult, error) {
	return service.CreateTaskResult{TaskID: "task-test-1", RequestID: req.RequestID, Status: service.CreateStatusAccepted}, nil
}
func (s *stubShotPreviewService) GetTask(_ context.Context, _ service.GetTaskRequest) (service.TaskView, error) {
	return s.task, nil
}
func (s *stubShotPreviewService) CancelTask(_ context.Context, _ service.CancelTaskRequest) (service.TaskView, error) {
	return s.task, nil
}
func (s *stubShotPreviewService) RetryTask(_ context.Context, _ service.RetryTaskRequest) (service.TaskView, error) {
	return s.task, nil
}
func (s *stubShotPreviewService) ConfirmStep(_ context.Context, _ service.ConfirmStepRequest) error {
	return nil
}
func (s *stubShotPreviewService) AdjustStep(_ context.Context, _ service.AdjustStepRequest) error {
	return nil
}
func (s *stubShotPreviewService) SubscribeEvents(_ context.Context, _ string) (<-chan pipeline.PipelineEvent, func(), error) {
	ch := make(chan pipeline.PipelineEvent, 1)
	ch <- pipeline.PipelineEvent{
		TaskID: "task-test-1",
		Type:   pipeline.EventTaskSucceeded,
		Status: string(service.TaskStatusSucceeded),
	}
	return ch, func() { close(ch) }, nil
}

func TestHTTPGateway_StreamAndConfirm(t *testing.T) {
	stub := &stubShotPreviewService{
		task: service.TaskView{
			TaskID: "task-test-1",
			Status: service.TaskStatusRunning,
			Nodes: []service.NodeView{
				{ID: "Intent", Status: service.NodeStatusWaitingConfirmation},
			},
		},
	}
	shotHandler := handlerv0_1.NewShotPreviewHandler(stub)
	gw := NewHTTPGateway(shotHandler, nil, nil)

	// Test GET /api/v0_1/shot-preview/task/stream
	reqStream := httptest.NewRequest(http.MethodGet, "/api/v0_1/shot-preview/task/stream?task_id=task-test-1", nil)
	wStream := httptest.NewRecorder()
	gw.ServeHTTP(wStream, reqStream)
	if wStream.Code != http.StatusOK {
		t.Fatalf("expected 200 from stream, got %d: %s", wStream.Code, wStream.Body.String())
	}
	if !bytes.Contains(wStream.Body.Bytes(), []byte("task_snapshot")) {
		t.Fatalf("expected task_snapshot event in stream body: %s", wStream.Body.String())
	}

	// Test POST /api/v0_1/shot-preview/task/node/confirm
	confirmBody := []byte(`{"task_id":"task-test-1","node_id":"Intent"}`)
	reqConfirm := httptest.NewRequest(http.MethodPost, "/api/v0_1/shot-preview/task/node/confirm", bytes.NewReader(confirmBody))
	wConfirm := httptest.NewRecorder()
	gw.ServeHTTP(wConfirm, reqConfirm)
	if wConfirm.Code != http.StatusOK {
		t.Fatalf("expected 200 from confirm, got %d: %s", wConfirm.Code, wConfirm.Body.String())
	}

	// Test POST /api/v0_1/shot-preview/task/node/adjust
	adjustBody := []byte(`{"task_id":"task-test-1","node_id":"Intent","output_json":"{}"}`)
	reqAdjust := httptest.NewRequest(http.MethodPost, "/api/v0_1/shot-preview/task/node/adjust", bytes.NewReader(adjustBody))
	wAdjust := httptest.NewRecorder()
	gw.ServeHTTP(wAdjust, reqAdjust)
	if wAdjust.Code != http.StatusOK {
		t.Fatalf("expected 200 from adjust, got %d: %s", wAdjust.Code, wAdjust.Body.String())
	}
}

func TestHTTPGateway_AssetUpload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "http-gateway-upload-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	metaFile := filepath.Join(tempDir, "metadata.json")
	store, err := service.NewFileAssetStore(metaFile)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	assetHandler := handlerv0_1.NewAssetHandler(service.NewAssetService(store))
	gw := NewHTTPGateway(nil, nil, assetHandler)
	gw.SetAssetsDir(tempDir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "scene_drag_drop.blend")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fileContent := []byte("BLENDER_SCENE_MOCK_BINARY_DATA")
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	_ = writer.WriteField("user_id", "default_user_001")
	_ = writer.WriteField("description", "拖拽上传测试工程")
	_ = writer.WriteField("tags", "拖拽,测试")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v0_1/assets/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	gw.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from upload, got %d: %s", w.Code, w.Body.String())
	}

	var res api.RegisterAssetResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if res.AssetId == "" {
		t.Fatalf("expected non-empty asset ID")
	}

	// Verify physical file was created in tempDir
	targetFile := filepath.Join(tempDir, "scene_drag_drop.blend")
	savedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("expected target file %s to exist on disk: %v", targetFile, err)
	}
	if string(savedContent) != string(fileContent) {
		t.Errorf("saved content mismatch, got %s", string(savedContent))
	}

	// Verify asset is indexed in store
	record, ok, err := store.Get(context.Background(), "default_user_001", res.AssetId)
	if err != nil || !ok {
		t.Fatalf("expected asset %s in store: %v", res.AssetId, err)
	}
	if record.Name != "scene_drag_drop.blend" || record.FileSizeBytes != int64(len(fileContent)) {
		t.Errorf("unexpected record attributes: %+v", record)
	}
	if record.AssetType != api.AssetType_MODEL_3D {
		t.Errorf("expected MODEL_3D, got %v", record.AssetType)
	}
}

func TestHTTPGateway_Authentication(t *testing.T) {
	assetSvc := service.NewAssetService(service.NewInMemoryAssetStore())
	assetHandler := handlerv0_1.NewAssetHandler(assetSvc)

	gw := NewHTTPGateway(nil, nil, assetHandler)
	testToken := "bspe_secret_auth_token_999"
	gw.SetAccessToken(testToken)

	// 1. Without auth -> Protected route /api/v0_1/assets should return 401
	reqUnauth := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets", nil)
	wUnauth := httptest.NewRecorder()
	gw.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", wUnauth.Code)
	}

	// 2. /health should be whitelisted and return 200
	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	wHealth := httptest.NewRecorder()
	gw.ServeHTTP(wHealth, reqHealth)
	if wHealth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /health, got %d", wHealth.Code)
	}

	// 3. /api/v0_1/auth/check without cookie should return 401
	reqCheckUnauth := httptest.NewRequest(http.MethodGet, "/api/v0_1/auth/check", nil)
	wCheckUnauth := httptest.NewRecorder()
	gw.ServeHTTP(wCheckUnauth, reqCheckUnauth)
	if wCheckUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 from /auth/check when unauthenticated, got %d", wCheckUnauth.Code)
	}

	// 4. POST /api/v0_1/auth/login with wrong token -> 401
	wrongLoginBody := []byte(`{"token":"invalid_token"}`)
	reqWrongLogin := httptest.NewRequest(http.MethodPost, "/api/v0_1/auth/login", bytes.NewReader(wrongLoginBody))
	wWrongLogin := httptest.NewRecorder()
	gw.ServeHTTP(wWrongLogin, reqWrongLogin)
	if wWrongLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 from login with wrong token, got %d", wWrongLogin.Code)
	}

	// 5. POST /api/v0_1/auth/login with correct token -> 200 and Set-Cookie
	correctLoginBody := []byte(`{"token":"` + testToken + `"}`)
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v0_1/auth/login", bytes.NewReader(correctLoginBody))
	wLogin := httptest.NewRecorder()
	gw.ServeHTTP(wLogin, reqLogin)
	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	cookies := wLogin.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "bspe_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("expected bspe_session cookie in login response")
	}

	// 6. Request /api/v0_1/auth/check WITH session cookie -> 200
	reqCheckAuthed := httptest.NewRequest(http.MethodGet, "/api/v0_1/auth/check", nil)
	reqCheckAuthed.AddCookie(sessionCookie)
	wCheckAuthed := httptest.NewRecorder()
	gw.ServeHTTP(wCheckAuthed, reqCheckAuthed)
	if wCheckAuthed.Code != http.StatusOK {
		t.Fatalf("expected 200 from /auth/check with cookie, got %d", wCheckAuthed.Code)
	}

	// 7. Request protected route /api/v0_1/assets WITH session cookie -> 200
	reqAssetsAuthed := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets", nil)
	reqAssetsAuthed.AddCookie(sessionCookie)
	wAssetsAuthed := httptest.NewRecorder()
	gw.ServeHTTP(wAssetsAuthed, reqAssetsAuthed)
	if wAssetsAuthed.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/v0_1/assets with cookie, got %d", wAssetsAuthed.Code)
	}

	// 8. Request protected route WITH Authorization Header -> 200
	reqAssetsBearer := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets", nil)
	reqAssetsBearer.Header.Set("Authorization", "Bearer "+testToken)
	wAssetsBearer := httptest.NewRecorder()
	gw.ServeHTTP(wAssetsBearer, reqAssetsBearer)
	if wAssetsBearer.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/v0_1/assets with Bearer token, got %d", wAssetsBearer.Code)
	}

	// 9. Logout -> clears session cookie
	reqLogout := httptest.NewRequest(http.MethodPost, "/api/v0_1/auth/logout", nil)
	reqLogout.AddCookie(sessionCookie)
	wLogout := httptest.NewRecorder()
	gw.ServeHTTP(wLogout, reqLogout)
	if wLogout.Code != http.StatusOK {
		t.Fatalf("expected 200 from logout, got %d", wLogout.Code)
	}
	logoutCookies := wLogout.Result().Cookies()
	var clearedCookie *http.Cookie
	for _, c := range logoutCookies {
		if c.Name == "bspe_session" {
			clearedCookie = c
			break
		}
	}
	if clearedCookie == nil || clearedCookie.MaxAge > 0 {
		t.Fatalf("expected cleared session cookie with negative max age")
	}
}

