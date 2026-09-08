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
			Name:          "cyberpunk_heroine_eva.blend",
			AssetType:     v0_1.AssetType_MODEL_3D,
			FileFormat:    "blend",
			FileSizeBytes: 218 * 1024 * 1024,
			StorageURI:    "blender://assets/characters/cyberpunk_heroine_eva.blend",
			ThumbnailURI:  "asset_preview_heroine_eva",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"主角人物", "Rigify骨骼", "次世代PBR", "赛博朋克"},
			CreatedAt:     now.Add(-2 * time.Hour),
			UpdatedAt:     now.Add(-10 * time.Minute),
			Description:   "主角角色高模（伊娃），带 Rigify 完整人型骨骼与面部表情驱动，包含4套高分辨率 PBR 材质贴图与次世代布料物理",
		},
		{
			AssetID:       "asset_default_002",
			UserID:        "default_user_001",
			Name:          "cyborg_enforcer_goliath.blend",
			AssetType:     v0_1.AssetType_MODEL_3D,
			FileFormat:    "blend",
			FileSizeBytes: 185 * 1024 * 1024,
			StorageURI:    "blender://assets/characters/cyborg_enforcer_goliath.blend",
			ThumbnailURI:  "asset_preview_cyborg_goliath",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"反派角色", "机械装甲", "武器挂载", "重型骨骼"},
			CreatedAt:     now.Add(-6 * time.Hour),
			UpdatedAt:     now.Add(-45 * time.Minute),
			Description:   "反派机械重装改造人，重型机械义肢与动力装甲，支持全身战术动作骨骼与发光着色器",
		},
		{
			AssetID:       "asset_default_003",
			UserID:        "default_user_001",
			Name:          "neo_shibuya_night_street.blend",
			AssetType:     v0_1.AssetType_MODEL_3D,
			FileFormat:    "blend",
			FileSizeBytes: 348 * 1024 * 1024,
			StorageURI:    "blender://assets/scenes/neo_shibuya_night_street.blend",
			ThumbnailURI:  "asset_preview_street_night",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"核心场景", "雨夜街道", "SSR反射", "体积光"},
			CreatedAt:     now.Add(-12 * time.Hour),
			UpdatedAt:     now.Add(-2 * time.Hour),
			Description:   "雨夜赛博都市核心街景，包含高细节湿漉路面贴图、霓虹招牌反射着色与体积雾灯光预设",
		},
		{
			AssetID:       "asset_default_004",
			UserID:        "default_user_001",
			Name:          "cinematic_dolly_zoom_vertigo.json",
			AssetType:     v0_1.AssetType_SHOT_PRESET,
			FileFormat:    "json",
			FileSizeBytes: 18 * 1024,
			StorageURI:    "blender://assets/presets/cinematic_dolly_zoom_vertigo.json",
			ThumbnailURI:  "asset_preview_dolly_zoom",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"推拉变焦", "DollyZoom", "眩晕透视", "120帧"},
			CreatedAt:     now.Add(-24 * time.Hour),
			UpdatedAt:     now.Add(-4 * time.Hour),
			Description:   "经典希区柯克推拉变焦运镜预设，相机前进同时同步逆向平滑调整视场角 (35mm->85mm)，营造强烈眩晕透视畸变",
		},
		{
			AssetID:       "asset_default_005",
			UserID:        "default_user_001",
			Name:          "dynamic_orbit_low_angle_hero.py",
			AssetType:     v0_1.AssetType_SHOT_PRESET,
			FileFormat:    "py",
			FileSizeBytes: 12 * 1024,
			StorageURI:    "blender://assets/presets/dynamic_orbit_low_angle_hero.py",
			ThumbnailURI:  "asset_preview_orbit_hero",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"360环绕", "低仰角", "主角运镜", "曲线平滑"},
			CreatedAt:     now.Add(-48 * time.Hour),
			UpdatedAt:     now.Add(-8 * time.Hour),
			Description:   "低角度主角环绕轨迹脚本，以主角胸口为注视点，低仰角 15° 完成 360 度动态匀速圆周运镜",
		},
		{
			AssetID:       "asset_default_006",
			UserID:        "default_user_001",
			Name:          "rainy_tokyo_cyber_8k.hdr",
			AssetType:     v0_1.AssetType_MATERIAL,
			FileFormat:    "hdr",
			FileSizeBytes: 96 * 1024 * 1024,
			StorageURI:    "blender://assets/materials/rainy_tokyo_cyber_8k.hdr",
			ThumbnailURI:  "asset_preview_hdri_tokyo",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"8K HDRI", "夜景环境光", "PBR反射", "全景照明"},
			CreatedAt:     now.Add(-72 * time.Hour),
			UpdatedAt:     now.Add(-16 * time.Hour),
			Description:   "32-bit 浮点高动态范围环境贴图，采集自雨夜都市霓虹灯光，提供全方位物理真实环境反射与天空照明",
		},
		{
			AssetID:       "asset_default_007",
			UserID:        "default_user_001",
			Name:          "combat_dodge_roll_rootmotion.abc",
			AssetType:     v0_1.AssetType_ANIMATION,
			FileFormat:    "abc",
			FileSizeBytes: 145 * 1024 * 1024,
			StorageURI:    "blender://assets/animations/combat_dodge_roll_rootmotion.abc",
			ThumbnailURI:  "asset_preview_anim_roll",
			Status:        v0_1.AssetStatus_AVAILABLE,
			Tags:          []string{"Alembic序列", "动作捕捉", "根骨骼位移", "战术动作"},
			CreatedAt:     now.Add(-96 * time.Hour),
			UpdatedAt:     now.Add(-24 * time.Hour),
			Description:   "Alembic 格式战术侧滚翻根骨骼位移缓存序列，支持无缝导入并绑定至 Rigify 人型角色",
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
