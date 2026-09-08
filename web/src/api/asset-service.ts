import { request } from "./client";
import {
  ListAssetsRequest,
  ListAssetsResponse,
  RegisterAssetRequest,
  RegisterAssetResponse,
  DeleteAssetRequest,
  DeleteAssetResponse,
  GetAssetStatsRequest,
  GetAssetStatsResponse,
} from "./types";

export class AssetService {
  static async listAssets(req: ListAssetsRequest): Promise<ListAssetsResponse> {
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
    const params = new URLSearchParams();
    if (req.user_id) params.append("user_id", req.user_id);

    return request<GetAssetStatsResponse>(`/api/v0_1/assets/stats?${params.toString()}`, {
      method: "GET",
    });
  }

  static async registerAsset(req: RegisterAssetRequest): Promise<RegisterAssetResponse> {
    return request<RegisterAssetResponse>("/api/v0_1/assets", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  static async deleteAsset(req: DeleteAssetRequest): Promise<DeleteAssetResponse> {
    return request<DeleteAssetResponse>("/api/v0_1/assets/delete", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }
}
