package service

import (
	"context"
	"testing"

	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

func TestAssetService_ListAssets(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()

	// 1. List all assets
	res, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId: "default_user_001",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(res.Assets) == 0 {
		t.Fatalf("expected seeded assets, got 0")
	}
	if res.TotalCount != int64(len(res.Assets)) {
		t.Errorf("expected total count %d, got %d", len(res.Assets), res.TotalCount)
	}

	// 2. Filter by asset type: MODEL_3D
	modelType := v0_1.AssetType_MODEL_3D
	resModels, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId:    "default_user_001",
		AssetType: &modelType,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for _, a := range resModels.Assets {
		if a.AssetType != modelType {
			t.Errorf("expected asset type %v, got %v", modelType, a.AssetType)
		}
	}

	// 3. Search with keyword
	kw := "cyberpunk"
	resSearch, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId:       "default_user_001",
		QueryKeyword: &kw,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(resSearch.Assets) == 0 {
		t.Errorf("expected at least 1 match for keyword %q", kw)
	}
}

func TestAssetService_RegisterAndDeleteAsset(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()

	// 1. Register new asset
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

	// 2. Verify it's queryable
	listRes, err := svc.ListAssets(ctx, &v0_1.ListAssetsRequest{
		UserId: "user_test_123",
	})
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}
	found := false
	for _, a := range listRes.Assets {
		if a.AssetId == regRes.AssetId {
			found = true
			if a.Name != name {
				t.Errorf("expected name %q, got %q", name, a.Name)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected to find registered asset %q", regRes.AssetId)
	}

	// 3. Delete the asset
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

	// 4. Verify deleted
	delCheck, _ := svc.DeleteAsset(ctx, &v0_1.DeleteAssetRequest{
		UserId:  "user_test_123",
		AssetId: regRes.AssetId,
	})
	if delCheck.Success {
		t.Errorf("expected second deletion to return false, got true")
	}
}

func TestAssetService_GetAssetStats(t *testing.T) {
	svc := NewAssetService(NewInMemoryAssetStore())
	ctx := context.Background()

	stats, err := svc.GetAssetStats(ctx, &v0_1.GetAssetStatsRequest{
		UserId: "default_user_001",
	})
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalAssets <= 0 {
		t.Errorf("expected positive total assets, got %d", stats.TotalAssets)
	}
	if stats.TotalModels <= 0 {
		t.Errorf("expected positive total models, got %d", stats.TotalModels)
	}
	if stats.TotalPresets <= 0 {
		t.Errorf("expected positive total presets, got %d", stats.TotalPresets)
	}
	if stats.TotalStorageBytes <= 0 {
		t.Errorf("expected positive storage bytes, got %d", stats.TotalStorageBytes)
	}
}
