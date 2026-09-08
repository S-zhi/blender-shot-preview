import { request, USE_MOCK_API } from "./client";
import {
  ListAssetsRequest,
  ListAssetsResponse,
  RegisterAssetRequest,
  RegisterAssetResponse,
  DeleteAssetRequest,
  DeleteAssetResponse,
  GetAssetStatsRequest,
  GetAssetStatsResponse,
  AssetType,
  AssetStatus,
} from "./types";

export class AssetService {
  static async listAssets(req: ListAssetsRequest): Promise<ListAssetsResponse> {
    if (USE_MOCK_API) {
      return {
        assets: [
          {
            asset_id: "mock_1",
            name: "cyberpunk_street_night.blend",
            asset_type: AssetType.MODEL_3D,
            file_format: "blend",
            file_size_bytes: 348 * 1024 * 1024,
            storage_uri: "blender://assets/models/cyberpunk_street_night.blend",
            status: AssetStatus.AVAILABLE,
            tags: ["赛博朋克", "雨夜", "SSR光影"],
            created_at: new Date().toISOString(),
            updated_at: "10分钟前",
            description: "包含高细节湿漉路面贴图、霓虹招牌顶点着色与体积雾灯光设置",
          },
        ],
        total_count: 1,
      };
    }

    const params = new URLSearchParams();
    if (req.user_id) params.append("user_id", req.user_id);
    if (req.asset_type) params.append("asset_type", String(req.asset_type));
    if (req.query_keyword) params.append("query_keyword", req.query_keyword);
    if (req.page_num) params.append("page_num", String(req.page_num));
    if (req.page_size) params.append("page_size", String(req.page_size));

    return request<ListAssetsResponse>(`/api/v0_1/assets?${params.toString()}`, {
      method: "GET",
    });
  }

  static async getAssetStats(req: GetAssetStatsRequest): Promise<GetAssetStatsResponse> {
    if (USE_MOCK_API) {
      return {
        total_assets: 6,
        total_models: 2,
        total_presets: 2,
        total_storage_bytes: 841 * 1024 * 1024,
      };
    }

    const params = new URLSearchParams();
    if (req.user_id) params.append("user_id", req.user_id);

    return request<GetAssetStatsResponse>(`/api/v0_1/assets/stats?${params.toString()}`, {
      method: "GET",
    });
  }

  static async registerAsset(req: RegisterAssetRequest): Promise<RegisterAssetResponse> {
    if (USE_MOCK_API) {
      return {
        asset_id: `asset_${Date.now()}`,
        status: AssetStatus.AVAILABLE,
      };
    }

    return request<RegisterAssetResponse>("/api/v0_1/assets", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  static async deleteAsset(req: DeleteAssetRequest): Promise<DeleteAssetResponse> {
    if (USE_MOCK_API) {
      return { success: true };
    }

    return request<DeleteAssetResponse>("/api/v0_1/assets/delete", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }
}
