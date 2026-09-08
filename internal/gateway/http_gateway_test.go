package gateway

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
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

	// 2. Test GET /api/v0_1/assets
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
	if len(listRes.Assets) == 0 {
		t.Errorf("expected seeded assets from gateway list")
	}

	// 3. Test GET /api/v0_1/assets/stats
	reqStats := httptest.NewRequest(http.MethodGet, "/api/v0_1/assets/stats?user_id=default_user_001", nil)
	wStats := httptest.NewRecorder()
	gw.ServeHTTP(wStats, reqStats)
	if wStats.Code != http.StatusOK {
		t.Errorf("expected 200 from /api/v0_1/assets/stats, got %d", wStats.Code)
	}

	// 4. Test POST /api/v0_1/assets
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

	// 5. Test POST /api/v0_1/assets/upload (multipart .blend upload)
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("file", "heroine_alita_rigged.blend")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("BLENDER_v401\x00Armature_Rigify\x00Bone_head\x00"))
	_ = w.WriteField("user_id", "default_user_001")
	_ = w.WriteField("name", "heroine_alita_rigged.blend")
	_ = w.Close()

	reqUpload := httptest.NewRequest(http.MethodPost, "/api/v0_1/assets/upload", &b)
	reqUpload.Header.Set("Content-Type", w.FormDataContentType())
	wUpload := httptest.NewRecorder()
	gw.ServeHTTP(wUpload, reqUpload)
	if wUpload.Code != http.StatusOK {
		t.Errorf("expected 200 from POST /api/v0_1/assets/upload, got %d: %s", wUpload.Code, wUpload.Body.String())
	}
}
