package v0_1

import (
	"context"
	"strings"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

// AssetHandler implements AssetServiceV0_1 Kitex interface
type AssetHandler struct {
	service service.AssetService
}

func NewAssetHandler(svc service.AssetService) *AssetHandler {
	return &AssetHandler{service: svc}
}

func (h *AssetHandler) ListAssets(ctx context.Context, request *api.ListAssetsRequest) (*api.ListAssetsResponse, error) {
	if request == nil {
		return nil, kerrors.NewBizStatusError(400, "request is required")
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, kerrors.NewBizStatusError(400, "user_id is required")
	}
	if h == nil || h.service == nil {
		return nil, kerrors.NewBizStatusError(500, "asset service unavailable")
	}

	res, err := h.service.ListAssets(ctx, request)
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, err.Error())
	}
	return res, nil
}

func (h *AssetHandler) RegisterAsset(ctx context.Context, request *api.RegisterAssetRequest) (*api.RegisterAssetResponse, error) {
	if request == nil {
		return nil, kerrors.NewBizStatusError(400, "request is required")
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, kerrors.NewBizStatusError(400, "user_id is required")
	}
	if strings.TrimSpace(request.GetName()) == "" {
		return nil, kerrors.NewBizStatusError(400, "name is required")
	}
	if strings.TrimSpace(request.GetStorageUri()) == "" {
		return nil, kerrors.NewBizStatusError(400, "storage_uri is required")
	}
	if h == nil || h.service == nil {
		return nil, kerrors.NewBizStatusError(500, "asset service unavailable")
	}

	res, err := h.service.RegisterAsset(ctx, request)
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, err.Error())
	}
	return res, nil
}

func (h *AssetHandler) DeleteAsset(ctx context.Context, request *api.DeleteAssetRequest) (*api.DeleteAssetResponse, error) {
	if request == nil {
		return nil, kerrors.NewBizStatusError(400, "request is required")
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, kerrors.NewBizStatusError(400, "user_id is required")
	}
	if strings.TrimSpace(request.GetAssetId()) == "" {
		return nil, kerrors.NewBizStatusError(400, "asset_id is required")
	}
	if h == nil || h.service == nil {
		return nil, kerrors.NewBizStatusError(500, "asset service unavailable")
	}

	res, err := h.service.DeleteAsset(ctx, request)
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, err.Error())
	}
	return res, nil
}

func (h *AssetHandler) GetAssetStats(ctx context.Context, request *api.GetAssetStatsRequest) (*api.GetAssetStatsResponse, error) {
	if request == nil {
		return nil, kerrors.NewBizStatusError(400, "request is required")
	}
	userID := strings.TrimSpace(request.GetUserId())
	if userID == "" {
		return nil, kerrors.NewBizStatusError(400, "user_id is required")
	}
	if h == nil || h.service == nil {
		return nil, kerrors.NewBizStatusError(500, "asset service unavailable")
	}

	res, err := h.service.GetAssetStats(ctx, request)
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, err.Error())
	}
	return res, nil
}
