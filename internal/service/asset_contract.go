package service

import (
	"context"
	"time"

	"github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

// AssetRecord represents an internal storage model for assets
type AssetRecord struct {
	AssetID       string
	UserID        string
	Name          string
	AssetType     v0_1.AssetType
	FileFormat    string
	FileSizeBytes int64
	StorageURI    string
	ThumbnailURI  string
	Status        v0_1.AssetStatus
	Tags          []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Description   string
}

// AssetStore defines the persistence interface for asset metadata
type AssetStore interface {
	List(ctx context.Context, userID string, assetType *v0_1.AssetType, keyword string, pageNum, pageSize int) ([]AssetRecord, int64, error)
	Get(ctx context.Context, userID, assetID string) (AssetRecord, bool, error)
	Create(ctx context.Context, record AssetRecord) error
	Delete(ctx context.Context, userID, assetID string) (bool, error)
	GetStats(ctx context.Context, userID string) (total, models, presets, totalBytes int64, err error)
}

// AssetService defines the domain interface for asset management
type AssetService interface {
	ListAssets(ctx context.Context, req *v0_1.ListAssetsRequest) (*v0_1.ListAssetsResponse, error)
	RegisterAsset(ctx context.Context, req *v0_1.RegisterAssetRequest) (*v0_1.RegisterAssetResponse, error)
	DeleteAsset(ctx context.Context, req *v0_1.DeleteAssetRequest) (*v0_1.DeleteAssetResponse, error)
	GetAssetStats(ctx context.Context, req *v0_1.GetAssetStatsRequest) (*v0_1.GetAssetStatsResponse, error)
}
