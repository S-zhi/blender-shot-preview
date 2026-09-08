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
  AssetView,
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

  static async uploadAsset(
    file: File,
    meta: {
      user_id?: string;
      name?: string;
      asset_type?: number;
      description?: string;
      tags?: string[];
    } = {}
  ): Promise<RegisterAssetResponse> {
    const formData = new FormData();
    formData.append("file", file);
    if (meta.user_id) formData.append("user_id", meta.user_id);
    if (meta.name) formData.append("name", meta.name);
    if (meta.asset_type !== undefined) formData.append("asset_type", String(meta.asset_type));
    if (meta.description) formData.append("description", meta.description);
    if (meta.tags && meta.tags.length > 0) {
      formData.append("tags", JSON.stringify(meta.tags));
    }

    const response = await fetch("/api/v0_1/assets/upload", {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      let errText = `Upload failed with status ${response.status}`;
      try {
        const errJson = await response.json();
        if (errJson.error) errText = errJson.error;
      } catch {
        // ignore
      }
      throw new Error(errText);
    }

    return (await response.json()) as RegisterAssetResponse;
  }

  static async uploadAssetFile(file: File, name?: string, userId?: string): Promise<AssetView> {
    const res = await this.uploadAsset(file, {
      user_id: userId || "default_user_001",
      name: name || file.name,
    });

    // Return an AssetView projection for caller compatibility
    return {
      asset_id: res.asset_id,
      name: name || file.name,
      asset_type: 1, // MODEL_3D
      file_format: file.name.split(".").pop() || "blend",
      file_size_bytes: file.size,
      storage_uri: `blender://assets/${file.name}`,
      status: res.status,
      tags: ["拖拽导入", file.name.split(".").pop() || "blend"],
      created_at: new Date().toISOString(),
      updated_at: "刚刚",
    };
  }
}
