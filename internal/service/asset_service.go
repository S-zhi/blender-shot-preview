package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

// InMemoryAssetStore provides a thread-safe in-memory implementation of AssetStore
type InMemoryAssetStore struct {
	mu     sync.RWMutex
	assets map[string]AssetRecord // key: asset_id
}

func NewInMemoryAssetStore() *InMemoryAssetStore {
	store := &InMemoryAssetStore{
		assets: make(map[string]AssetRecord),
	}
	now := time.Now()
	defaults := []AssetRecord{
		{
			AssetID:       "asset_default_001",
			UserID:        "default_user_001",
			Name:          "cyberpunk_street_night.blend",
			AssetType:     v0_1.AssetType_MODEL_3D,
			FileFormat:    "blend",
			FileSizeBytes: 348 * 1024 * 1024,
			StorageURI:    "blender://assets/models/cyberpunk_street_night.blend",
			ThumbnailURI:  "blender://thumbnails/cyberpunk_street_night.png",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"赛博朋克", "雨夜", "SSR光影"},
			CreatedAt:     now.Add(-2 * time.Hour),
			UpdatedAt:     now.Add(-10 * time.Minute),
			Description:   "包含高细节湿漉路面贴图、霓虹招牌顶点着色与体积雾灯光设置",
		},
		{
			AssetID:       "asset_default_002",
			UserID:        "default_user_001",
			Name:          "dolly_zoom_rack_focus_35to85.json",
			AssetType:     v0_1.AssetType_SHOT_PRESET,
			FileFormat:    "json",
			FileSizeBytes: 12 * 1024,
			StorageURI:    "blender://assets/presets/dolly_zoom_rack_focus_35to85.json",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"推拉变焦", "平移对焦", "180帧"},
			CreatedAt:     now.Add(-5 * time.Hour),
			UpdatedAt:     now.Add(-1 * time.Hour),
			Description:   "Blender 摄像机平滑变焦焦点转移曲线数据，定焦前景水珠到背景人脸",
		},
		{
			AssetID:       "asset_default_003",
			UserID:        "default_user_001",
			Name:          "mechanical_watch_center.blend",
			AssetType:     v0_1.AssetType_MODEL_3D,
			FileFormat:    "blend",
			FileSizeBytes: 142 * 1024 * 1024,
			StorageURI:    "blender://assets/models/mechanical_watch_center.blend",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"工业产品", "精密零件", "金属材质"},
			CreatedAt:     now.Add(-24 * time.Hour),
			UpdatedAt:     now.Add(-12 * time.Hour),
			Description:   "高精度机械机芯齿轮结构模型，带独立中心对焦旋转轴",
		},
		{
			AssetID:       "asset_default_004",
			UserID:        "default_user_001",
			Name:          "orbit_360_smooth_yaw.py",
			AssetType:     v0_1.AssetType_SHOT_PRESET,
			FileFormat:    "py",
			FileSizeBytes: 8 * 1024,
			StorageURI:    "blender://assets/presets/orbit_360_smooth_yaw.py",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"圆周环绕", "DampedTrack", "恒速"},
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-24 * time.Hour),
			Description:   "Python 脚本驱动的 360 度圆周运镜轨迹生成算法",
		},
		{
			AssetID:       "asset_default_005",
			UserID:        "default_user_001",
			Name:          "tokyo_night_rain_4k.hdr",
			AssetType:     v0_1.AssetType_MATERIAL,
			FileFormat:    "hdr",
			FileSizeBytes: 68 * 1024 * 1024,
			StorageURI:    "blender://assets/materials/tokyo_night_rain_4k.hdr",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"HDRI", "夜景环境光", "4K"},
			CreatedAt:     now.Add(-72 * time.Hour),
			UpdatedAt:     now.Add(-36 * time.Hour),
			Description:   "32-bit 浮点高动态范围环境贴图，提供高真实度反射与环境照明",
		},
		{
			AssetID:       "asset_default_006",
			UserID:        "default_user_001",
			Name:          "fpv_drone_canyon_run.abc",
			AssetType:     v0_1.AssetType_ANIMATION,
			FileFormat:    "abc",
			FileSizeBytes: 215 * 1024 * 1024,
			StorageURI:    "blender://assets/animations/fpv_drone_canyon_run.abc",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"FPV无人机", "大景深", "动态模糊"},
			CreatedAt:     now.Add(-96 * time.Hour),
			UpdatedAt:     now.Add(-48 * time.Hour),
			Description:   "Alembic 格式无人机快速俯冲与穿越山谷相机运动缓存序列",
		},
	}
	for _, a := range defaults {
		store.assets[a.AssetID] = a
	}
	return store
}

func (s *InMemoryAssetStore) List(_ context.Context, userID string, assetType *v0_1.AssetType, keyword string, pageNum, pageSize int) ([]AssetRecord, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []AssetRecord
	kw := strings.ToLower(strings.TrimSpace(keyword))

	for _, a := range s.assets {
		if userID != "" && a.UserID != userID && a.UserID != "default_user_001" {
			continue
		}
		if assetType != nil && a.AssetType != *assetType {
			continue
		}
		if kw != "" {
			nameMatch := strings.Contains(strings.ToLower(a.Name), kw)
			descMatch := strings.Contains(strings.ToLower(a.Description), kw)
			tagMatch := false
			for _, t := range a.Tags {
				if strings.Contains(strings.ToLower(t), kw) {
					tagMatch = true
					break
				}
			}
			if !nameMatch && !descMatch && !tagMatch {
				continue
			}
		}
		filtered = append(filtered, a)
	}

	total := int64(len(filtered))
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	start := (pageNum - 1) * pageSize
	if start >= len(filtered) {
		return []AssetRecord{}, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total, nil
}

func (s *InMemoryAssetStore) Get(_ context.Context, userID, assetID string) (AssetRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.assets[assetID]
	if !ok {
		return AssetRecord{}, false, nil
	}
	if userID != "" && record.UserID != userID && record.UserID != "default_user_001" {
		return AssetRecord{}, false, nil
	}
	return record, true, nil
}

func (s *InMemoryAssetStore) Create(_ context.Context, record AssetRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.assets[record.AssetID] = record
	return nil
}

func (s *InMemoryAssetStore) Delete(_ context.Context, userID, assetID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.assets[assetID]
	if !ok {
		return false, nil
	}
	if userID != "" && record.UserID != userID && record.UserID != "default_user_001" {
		return false, nil
	}
	delete(s.assets, assetID)
	return true, nil
}

func (s *InMemoryAssetStore) GetStats(_ context.Context, userID string) (total, models, presets, totalBytes int64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, a := range s.assets {
		if userID != "" && a.UserID != userID && a.UserID != "default_user_001" {
			continue
		}
		total++
		totalBytes += a.FileSizeBytes
		switch a.AssetType {
		case v0_1.AssetType_MODEL_3D:
			models++
		case v0_1.AssetType_SHOT_PRESET:
			presets++
		}
	}
	return total, models, presets, totalBytes, nil
}

// AssetServiceImpl implements AssetService
type AssetServiceImpl struct {
	store AssetStore
}

func NewAssetService(store AssetStore) *AssetServiceImpl {
	if store == nil {
		store = NewInMemoryAssetStore()
	}
	return &AssetServiceImpl{store: store}
}

func (s *AssetServiceImpl) ListAssets(ctx context.Context, req *v0_1.ListAssetsRequest) (*v0_1.ListAssetsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	pageNum := 1
	if req.PageNum != nil && *req.PageNum > 0 {
		pageNum = int(*req.PageNum)
	}
	pageSize := 50
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = int(*req.PageSize)
	}
	kw := ""
	if req.QueryKeyword != nil {
		kw = *req.QueryKeyword
	}

	records, total, err := s.store.List(ctx, req.UserId, req.AssetType, kw, pageNum, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	views := make([]*v0_1.AssetView, len(records))
	for i, r := range records {
		thumb := r.ThumbnailURI
		desc := r.Description
		views[i] = &v0_1.AssetView{
			AssetId:       r.AssetID,
			Name:          r.Name,
			AssetType:     r.AssetType,
			FileFormat:    r.FileFormat,
			FileSizeBytes: r.FileSizeBytes,
			StorageUri:    r.StorageURI,
			ThumbnailUri:  &thumb,
			Status:        r.Status,
			Tags:          r.Tags,
			CreatedAt:     r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
			Description:   &desc,
		}
	}

	return &v0_1.ListAssetsResponse{
		Assets:     views,
		TotalCount: total,
	}, nil
}

func (s *AssetServiceImpl) RegisterAsset(ctx context.Context, req *v0_1.RegisterAssetRequest) (*v0_1.RegisterAssetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("asset name cannot be empty")
	}
	if strings.TrimSpace(req.StorageUri) == "" {
		return nil, fmt.Errorf("storage_uri cannot be empty")
	}

	now := time.Now()
	assetID := fmt.Sprintf("asset_%d", now.UnixNano())
	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}

	record := AssetRecord{
		AssetID:       assetID,
		UserID:        req.UserId,
		Name:          req.Name,
		AssetType:     req.AssetType,
		FileFormat:    req.FileFormat,
		FileSizeBytes: req.FileSizeBytes,
		StorageURI:    req.StorageUri,
		Status:        v0_1.AssetStatus_AVAILABLE,
		Tags:          req.Tags,
		CreatedAt:     now,
		UpdatedAt:     now,
		Description:   desc,
	}

	if err := s.store.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to register asset: %w", err)
	}

	return &v0_1.RegisterAssetResponse{
		AssetId: assetID,
		Status:  v0_1.AssetStatus_AVAILABLE,
	}, nil
}

func (s *AssetServiceImpl) DeleteAsset(ctx context.Context, req *v0_1.DeleteAssetRequest) (*v0_1.DeleteAssetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.AssetId) == "" {
		return nil, fmt.Errorf("asset_id cannot be empty")
	}

	ok, err := s.store.Delete(ctx, req.UserId, req.AssetId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete asset: %w", err)
	}

	return &v0_1.DeleteAssetResponse{
		Success: ok,
	}, nil
}

func (s *AssetServiceImpl) GetAssetStats(ctx context.Context, req *v0_1.GetAssetStatsRequest) (*v0_1.GetAssetStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	total, models, presets, totalBytes, err := s.store.GetStats(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset stats: %w", err)
	}

	return &v0_1.GetAssetStatsResponse{
		TotalAssets:       total,
		TotalModels:       models,
		TotalPresets:      presets,
		TotalStorageBytes: totalBytes,
	}, nil
}
