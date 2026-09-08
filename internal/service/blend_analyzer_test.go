package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectBlendFileAndRegister(t *testing.T) {
	tmpDir := t.TempDir()
	blendPath := filepath.Join(tmpDir, "cyberpunk_heroine_alita.blend")

	// Create fake .blend file with header and Armature content
	content := []byte("BLENDER_v401\x00\x00\x00Armature_Rigify_ORG-spine\x00Bone_head\x00")
	if err := os.WriteFile(blendPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	spec, err := InspectBlendFile(blendPath)
	if err != nil {
		t.Fatalf("InspectBlendFile failed: %v", err)
	}

	if !spec.IsCharacter {
		t.Errorf("expected IsCharacter to be true")
	}
	if spec.RigType != "Rigify Humanoid (Blender 4.x)" {
		t.Errorf("expected Rigify Humanoid, got %s", spec.RigType)
	}
	if spec.PolyCount <= 0 {
		t.Errorf("expected positive poly count")
	}

	store := NewInMemoryAssetStore()
	record, err := AnalyzeAndRegisterUploadedAsset(context.Background(), store, "user_001", "cyberpunk_heroine_alita.blend", blendPath, int64(len(content)))
	if err != nil {
		t.Fatalf("AnalyzeAndRegisterUploadedAsset failed: %v", err)
	}

	if record.AssetID == "" {
		t.Errorf("expected non-empty asset ID")
	}
	if record.ThumbnailURI == "" {
		t.Errorf("expected generated thumbnail URI")
	}
}
