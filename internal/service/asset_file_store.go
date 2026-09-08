package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

// FileAssetStore provides a persistent, thread-safe implementation of AssetStore backed by a JSON file.
type FileAssetStore struct {
	mu       sync.RWMutex
	filePath string
	assets   map[string]AssetRecord // key: asset_id
}

// NewFileAssetStore creates a new FileAssetStore backed by the specified file path.
// If the parent directory does not exist, it will be created.
// If the metadata file already exists, it loads existing assets into memory.
func NewFileAssetStore(filePath string) (*FileAssetStore, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create asset storage directory: %w", err)
	}

	store := &FileAssetStore{
		filePath: filePath,
		assets:   make(map[string]AssetRecord),
	}

	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		var records []AssetRecord
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, fmt.Errorf("failed to parse asset metadata from %s: %w", filePath, err)
		}
		for _, r := range records {
			store.assets[r.AssetID] = r
		}
	}

	return store, nil
}

// saveLocked writes the in-memory assets to disk atomically using a temp file. Caller must hold mu.Lock().
func (s *FileAssetStore) saveLocked() error {
	records := make([]AssetRecord, 0, len(s.assets))
	for _, r := range s.assets {
		records = append(records, r)
	}

	// Sort deterministically by CreatedAt desc
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal asset records: %w", err)
	}

	tmpFile := fmt.Sprintf("%s.tmp.%d", s.filePath, os.Getpid())
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp asset file: %w", err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to atomic commit asset file: %w", err)
	}

	return nil
}

func (s *FileAssetStore) List(_ context.Context, userID string, assetType *v0_1.AssetType, keyword string, pageNum, pageSize int) ([]AssetRecord, int64, error) {
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

	// Order by CreatedAt desc
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

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

func (s *FileAssetStore) Get(_ context.Context, userID, assetID string) (AssetRecord, bool, error) {
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

func (s *FileAssetStore) Create(_ context.Context, record AssetRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.assets[record.AssetID] = record
	return s.saveLocked()
}

func (s *FileAssetStore) Delete(_ context.Context, userID, assetID string) (bool, error) {
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
	if err := s.saveLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *FileAssetStore) GetStats(_ context.Context, userID string) (total, models, presets, totalBytes int64, err error) {
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
