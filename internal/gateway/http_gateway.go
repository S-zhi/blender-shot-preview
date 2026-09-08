package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	handlerv0_1 "github.com/S-zhi/blender-shot-preview/internal/handler/v0_1"
	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type HTTPGateway struct {
	shotHandler  *handlerv0_1.ShotPreviewHandler
	keyHandler   *handlerv0_1.LLMKeyHandler
	assetHandler *handlerv0_1.AssetHandler
	mux          *http.ServeMux
}

func NewHTTPGateway(
	shotHandler *handlerv0_1.ShotPreviewHandler,
	keyHandler *handlerv0_1.LLMKeyHandler,
	assetHandler *handlerv0_1.AssetHandler,
) *HTTPGateway {
	gw := &HTTPGateway{
		shotHandler:  shotHandler,
		keyHandler:   keyHandler,
		assetHandler: assetHandler,
		mux:          http.NewServeMux(),
	}
	gw.routes()
	return gw
}

func (g *HTTPGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Global CORS and content type headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	g.mux.ServeHTTP(w, r)
}

func (g *HTTPGateway) routes() {
	g.mux.HandleFunc("/health", g.handleHealth)
	g.mux.HandleFunc("/api/v0_1/shot-preview/task", g.handleShotPreviewTask)
	g.mux.HandleFunc("/api/v0_1/llm-gateway/key", g.handleLLMKey)
	g.mux.HandleFunc("/api/v0_1/assets", g.handleAssets)
	g.mux.HandleFunc("/api/v0_1/assets/upload", g.handleAssetUpload)
	g.mux.HandleFunc("/api/v0_1/assets/stats", g.handleAssetStats)
	g.mux.HandleFunc("/api/v0_1/assets/delete", g.handleAssetDelete)
}

func (g *HTTPGateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (g *HTTPGateway) handleShotPreviewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		g.getShotPreviewTask(w, r)
		return
	}
	if r.Method == http.MethodPost {
		g.createShotPreviewTask(w, r)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (g *HTTPGateway) getShotPreviewTask(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	taskID := q.Get("task_id")
	if taskID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "task_id is required"})
		return
	}
	userID := q.Get("user_id")
	if userID == "" {
		userID = "default_user_001"
	}

	if g.shotHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "shot preview service unavailable"})
		return
	}

	req := api.GetShotPreviewTaskRequest{
		UserId: userID,
		TaskId: taskID,
	}
	res, err := g.shotHandler.GetShotPreviewTask(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) createShotPreviewTask(w http.ResponseWriter, r *http.Request) {
	var req api.CreateShotPreviewTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body: " + err.Error()})
		return
	}

	if g.shotHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "shot preview service unavailable"})
		return
	}

	res, err := g.shotHandler.CreateShotPreviewTask(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) handleLLMKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ManageLLMKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body: " + err.Error()})
		return
	}

	if g.keyHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "llm key service unavailable"})
		return
	}

	res, err := g.keyHandler.ManageLLMKey(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		g.listAssets(w, r)
		return
	}
	if r.Method == http.MethodPost {
		g.registerAsset(w, r)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func (g *HTTPGateway) listAssets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	if userID == "" {
		userID = "default_user_001"
	}

	req := api.ListAssetsRequest{
		UserId: userID,
	}

	if kw := q.Get("query_keyword"); kw != "" {
		req.QueryKeyword = &kw
	}
	if typeStr := q.Get("asset_type"); typeStr != "" {
		if val, err := strconv.Atoi(typeStr); err == nil {
			assetType := api.AssetType(val)
			req.AssetType = &assetType
		}
	}
	if pageNum, err := strconv.Atoi(q.Get("page_num")); err == nil && pageNum > 0 {
		pn := int32(pageNum)
		req.PageNum = &pn
	}
	if pageSize, err := strconv.Atoi(q.Get("page_size")); err == nil && pageSize > 0 {
		ps := int32(pageSize)
		req.PageSize = &ps
	}

	res, err := g.assetHandler.ListAssets(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) registerAsset(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.UserId) == "" {
		req.UserId = "default_user_001"
	}

	res, err := g.assetHandler.RegisterAsset(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) handleAssetDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DeleteAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	if strings.TrimSpace(req.UserId) == "" {
		req.UserId = "default_user_001"
	}

	res, err := g.assetHandler.DeleteAsset(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (g *HTTPGateway) handleAssetUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 128 MB max upload
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file field is required in multipart form"})
		return
	}
	defer file.Close()

	userID := r.FormValue("user_id")
	if strings.TrimSpace(userID) == "" {
		userID = "default_user_001"
	}

	assetName := strings.TrimSpace(r.FormValue("name"))
	if assetName == "" {
		assetName = header.Filename
	}

	// Prepare storage directory
	uploadDir := filepath.Join("data", "assets", "uploads")
	_ = os.MkdirAll(uploadDir, 0755)

	targetFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	targetPath := filepath.Join(uploadDir, targetFilename)

	dst, err := os.Create(targetPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save file: " + err.Error()})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to write file content: " + err.Error()})
		return
	}

	// Run automatic Blender inspection & character feature extraction
	regReq, spec, err := service.InspectAndBuildRegisterRequest(userID, assetName, targetPath, written)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to analyze asset: " + err.Error()})
		return
	}

	if g.assetHandler == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset handler unavailable"})
		return
	}

	res, err := g.assetHandler.RegisterAsset(r.Context(), regReq)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register asset: " + err.Error()})
		return
	}

	// Return registered view along with extracted character specs
	responsePayload := map[string]any{
		"asset_id":       res.AssetId,
		"name":           regReq.Name,
		"file_format":    regReq.FileFormat,
		"file_size":      regReq.FileSizeBytes,
		"storage_uri":    regReq.StorageUri,
		"character_spec": spec,
		"status":         res.Status,
	}

	writeJSON(w, http.StatusOK, responsePayload)
}

func (g *HTTPGateway) handleAssetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "default_user_001"
	}

	req := api.GetAssetStatsRequest{
		UserId: userID,
	}

	res, err := g.assetHandler.GetAssetStats(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
