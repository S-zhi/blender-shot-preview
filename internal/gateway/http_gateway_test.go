package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
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

