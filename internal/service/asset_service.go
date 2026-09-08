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
	return &InMemoryAssetStore{
		assets: make(map[string]AssetRecord),
	}
}

func (s *InMemoryAssetStore) List(_ context.Context, userID string, assetType *v0_1.AssetType, keyword string, pageNum, pageSize int) ([]AssetRecord, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []AssetRecord
	kw := strings.ToLower(strings.TrimSpace(keyword))

	for _, a := range s.assets {
		if userID != "" && a.UserID != userID {
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
	if userID != "" && record.UserID != userID {
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
	if userID != "" && record.UserID != userID {
		return false, nil
	}
	delete(s.assets, assetID)
	return true, nil
}

func (s *InMemoryAssetStore) GetStats(_ context.Context, userID string) (total, models, presets, totalBytes int64, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, a := range s.assets {
		if userID != "" && a.UserID != userID {
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
