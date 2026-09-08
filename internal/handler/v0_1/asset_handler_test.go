package v0_1

import (
	"context"
	"testing"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

func TestAssetHandler_Validation(t *testing.T) {
	handler := NewAssetHandler(service.NewAssetService(nil))
	ctx := context.Background()

	// 1. Nil request
	_, err := handler.ListAssets(ctx, nil)
	if err == nil {
		t.Errorf("expected error for nil request")
	}

	// 2. Empty user_id
	_, err = handler.ListAssets(ctx, &api.ListAssetsRequest{})
	if err == nil {
		t.Errorf("expected error for empty user_id")
	}

	// 3. Register asset missing name
	_, err = handler.RegisterAsset(ctx, &api.RegisterAssetRequest{
		UserId: "test_user",
	})
	if err == nil {
		t.Errorf("expected error for missing name in RegisterAsset")
	}

	// 4. Delete asset missing asset_id
	_, err = handler.DeleteAsset(ctx, &api.DeleteAssetRequest{
		UserId: "test_user",
	})
	if err == nil {
		t.Errorf("expected error for missing asset_id in DeleteAsset")
	}

	// 5. Register an asset and verify GetAssetStats
	desc := "Handler test model"
	regRes, err := handler.RegisterAsset(ctx, &api.RegisterAssetRequest{
		UserId:        "default_user_001",
		Name:          "scene.blend",
		AssetType:     api.AssetType_MODEL_3D,
		FileFormat:    "blend",
		FileSizeBytes: 1024,
		StorageUri:    "blender://assets/scene.blend",
		Description:   &desc,
	})
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}
	if regRes.AssetId == "" {
		t.Fatalf("expected valid asset_id")
	}

	stats, err := handler.GetAssetStats(ctx, &api.GetAssetStatsRequest{
		UserId: "default_user_001",
	})
	if err != nil {
		t.Fatalf("expected successful stats, got error: %v", err)
	}
	if stats.TotalAssets != 1 {
		t.Errorf("expected 1 total asset, got %d", stats.TotalAssets)
	}
}
