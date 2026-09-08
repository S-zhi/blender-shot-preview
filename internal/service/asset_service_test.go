package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

func TestAssetService_EmptyByDefault(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()

	// 1. Initially empty
	res, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId: "user_test_default",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(res.Assets) != 0 || res.TotalCount != 0 {
		t.Fatalf("expected 0 assets initially (no mock data), got %d (total %d)", len(res.Assets), res.TotalCount)
	}

	stats, err := svc.GetAssetStats(ctx, &v0_1.GetAssetStatsRequest{
		UserId: "user_test_default",
	})
	if err != nil {
		t.Fatalf("expected no error getting stats, got: %v", err)
	}
	if stats.TotalAssets != 0 || stats.TotalStorageBytes != 0 {
		t.Fatalf("expected 0 stats initially, got total=%d bytes=%d", stats.TotalAssets, stats.TotalStorageBytes)
	}
}

func TestAssetService_RegisterListFilterStats(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()
	userID := "user_cyber_001"

	// 1. Register a 3D Model
	desc1 := "Cyberpunk city model"
	reg1, err := svc.RegisterAsset(ctx, &v0_1.RegisterAssetRequest{
		UserId:        userID,
		Name:          "cyberpunk_city.blend",
		AssetType:     v0_1.AssetType_MODEL_3D,
		FileFormat:    "blend",
		FileSizeBytes: 100 * 1024 * 1024,
		StorageUri:    "blender://assets/models/cyberpunk_city.blend",
		Description:   &desc1,
		Tags:          []string{"cyberpunk", "sci-fi"},
	})
	if err != nil {
		t.Fatalf("failed to register asset 1: %v", err)
	}

	// 2. Register a Shot Preset
	desc2 := "Dolly zoom camera track"
	reg2, err := svc.RegisterAsset(ctx, &v0_1.RegisterAssetRequest{
		UserId:        userID,
		Name:          "dolly_zoom.json",
		AssetType:     v0_1.AssetType_SHOT_PRESET,
		FileFormat:    "json",
		FileSizeBytes: 10 * 1024,
		StorageUri:    "blender://assets/presets/dolly_zoom.json",
		Description:   &desc2,
		Tags:          []string{"camera", "zoom"},
	})
	if err != nil {
		t.Fatalf("failed to register asset 2: %v", err)
	}

	// 3. List all assets
	listRes, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}
	if len(listRes.Assets) != 2 || listRes.TotalCount != 2 {
		t.Fatalf("expected 2 registered assets, got %d", len(listRes.Assets))
	}

	// 4. Filter by type MODEL_3D
	modelType := v0_1.AssetType_MODEL_3D
	modelsRes, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId:    userID,
		AssetType: &modelType,
	})
	if err != nil {
		t.Fatalf("failed to filter by model: %v", err)
	}
	if len(modelsRes.Assets) != 1 || modelsRes.Assets[0].AssetId != reg1.AssetId {
		t.Fatalf("expected 1 model match with id %s, got %d", reg1.AssetId, len(modelsRes.Assets))
	}

	// 5. Search with keyword
	kw := "dolly"
	searchRes, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId:       userID,
		QueryKeyword: &kw,
	})
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(searchRes.Assets) != 1 || searchRes.Assets[0].AssetId != reg2.AssetId {
		t.Fatalf("expected 1 keyword match with id %s", reg2.AssetId)
	}

	// 6. Check stats
	stats, err := svc.GetAssetStats(ctx, &v0_1.GetAssetStatsRequest{
		UserId: userID,
	})
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalAssets != 2 || stats.TotalModels != 1 || stats.TotalPresets != 1 {
		t.Errorf("unexpected stats: total=%d, models=%d, presets=%d", stats.TotalAssets, stats.TotalModels, stats.TotalPresets)
	}
	expectedBytes := int64(100*1024*1024 + 10*1024)
	if stats.TotalStorageBytes != expectedBytes {
		t.Errorf("expected %d bytes, got %d", expectedBytes, stats.TotalStorageBytes)
	}
}

func TestAssetService_RegisterAndDeleteAsset(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()

	name := "test_camera_rig.blend"
	uri := "blender://assets/models/test_camera_rig.blend"
	desc := "Test camera rig asset"
	regRes, err := svc.RegisterAsset(ctx, &v0_1.RegisterAssetRequest{
		UserId:        "user_test_123",
		Name:          name,
		AssetType:     v0_1.AssetType_MODEL_3D,
		FileFormat:    "blend",
		FileSizeBytes: 1024 * 1024,
		StorageUri:    uri,
		Description:   &desc,
		Tags:          []string{"test", "rig"},
	})
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}
	if regRes.AssetId == "" {
		t.Fatalf("expected non-empty asset ID")
	}

	delRes, err := svc.DeleteAsset(ctx, &v0_1.DeleteAssetRequest{
		UserId:  "user_test_123",
		AssetId: regRes.AssetId,
	})
	if err != nil {
		t.Fatalf("failed to delete asset: %v", err)
	}
	if !delRes.Success {
		t.Errorf("expected success deletion, got false")
	}

	delCheck, _ := svc.DeleteAsset(ctx, &v0_1.DeleteAssetRequest{
		UserId:  "user_test_123",
		AssetId: regRes.AssetId,
	})
	if delCheck.Success {
		t.Errorf("expected second deletion to return false, got true")
	}
}

func TestFileAssetStore_Persistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file-asset-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	metaFile := filepath.Join(tempDir, "assets", "metadata.json")

	// 1. Create first store instance and save record
	store1, err := NewFileAssetStore(metaFile)
	if err != nil {
		t.Fatalf("failed to initialize store1: %v", err)
	}

	ctx := context.Background()
	record1 := AssetRecord{
		AssetID:       "asset_persisted_001",
		UserID:        "user_persist",
		Name:          "hero_ship.blend",
		AssetType:     v0_1.AssetType_MODEL_3D,
		FileFormat:    "blend",
		FileSizeBytes: 2048,
		StorageURI:    "blender://assets/hero_ship.blend",
		Tags:          []string{"ship", "vfx"},
	}

	if err := store1.Create(ctx, record1); err != nil {
		t.Fatalf("failed to create record in store1: %v", err)
	}

	// Check file was created on disk
	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		t.Fatalf("expected metadata file %s to exist on disk", metaFile)
	}

	// 2. Instantiate a second store pointing to the same file (simulating server restart)
	store2, err := NewFileAssetStore(metaFile)
	if err != nil {
		t.Fatalf("failed to initialize store2 from disk: %v", err)
	}

	loaded, ok, err := store2.Get(ctx, "user_persist", "asset_persisted_001")
	if err != nil || !ok {
		t.Fatalf("expected to reload record from disk in store2, got ok=%v err=%v", ok, err)
	}
	if loaded.Name != "hero_ship.blend" || loaded.FileSizeBytes != 2048 {
		t.Errorf("reloaded record mismatch: %+v", loaded)
	}

	// 3. Delete from store2 and verify file updated
	deleted, err := store2.Delete(ctx, "user_persist", "asset_persisted_001")
	if err != nil || !deleted {
		t.Fatalf("failed to delete in store2: deleted=%v err=%v", deleted, err)
	}

	// 4. Instantiate a third store to verify deletion persisted
	store3, err := NewFileAssetStore(metaFile)
	if err != nil {
		t.Fatalf("failed to initialize store3: %v", err)
	}
	_, ok3, _ := store3.Get(ctx, "user_persist", "asset_persisted_001")
	if ok3 {
		t.Errorf("expected deleted asset to not be loaded in store3")
	}
}
