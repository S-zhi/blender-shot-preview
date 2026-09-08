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
  AssetView,
} from "./types";

export class AssetService {
  static async listAssets(req: ListAssetsRequest): Promise<ListAssetsResponse> {
    if (USE_MOCK_API) {
      const allAssets: AssetView[] = [
        {
          asset_id: "asset_default_001",
          name: "cyberpunk_heroine_eva.blend",
          asset_type: AssetType.MODEL_3D,
          file_format: "blend",
          file_size_bytes: 218 * 1024 * 1024,
          storage_uri: "blender://assets/characters/cyberpunk_heroine_eva.blend",
          thumbnail_uri: "asset_preview_heroine_eva",
          status: AssetStatus.AVAILABLE,
          tags: ["主角人物", "Rigify骨骼", "次世代PBR", "赛博朋克"],
          created_at: new Date(Date.now() - 7200000).toISOString(),
          updated_at: "10分钟前",
          description: "主角角色高模（伊娃），带 Rigify 完整人型骨骼与面部表情驱动，包含4套高分辨率 PBR 材质贴图与次世代布料物理",
          metadata: {
            poly_count: 128450,
            vertex_count: 96200,
            rig_type: "Rigify Humanoid",
            has_face_rig: true,
          },
        },
        {
          asset_id: "asset_default_002",
          name: "cyborg_enforcer_goliath.blend",
          asset_type: AssetType.MODEL_3D,
          file_format: "blend",
          file_size_bytes: 185 * 1024 * 1024,
          storage_uri: "blender://assets/characters/cyborg_enforcer_goliath.blend",
          thumbnail_uri: "asset_preview_cyborg_goliath",
          status: AssetStatus.AVAILABLE,
          tags: ["反派角色", "机械装甲", "武器挂载", "重型骨骼"],
          created_at: new Date(Date.now() - 21600000).toISOString(),
          updated_at: "45分钟前",
          description: "反派机械重装改造人，重型机械义肢与动力装甲，支持全身战术动作骨骼与发光着色器",
          metadata: {
            poly_count: 164200,
            vertex_count: 112000,
            rig_type: "Heavy Mechanoid Rig",
            has_face_rig: false,
          },
        },
        {
          asset_id: "asset_default_003",
          name: "neo_shibuya_night_street.blend",
          asset_type: AssetType.MODEL_3D,
          file_format: "blend",
          file_size_bytes: 348 * 1024 * 1024,
          storage_uri: "blender://assets/scenes/neo_shibuya_night_street.blend",
          thumbnail_uri: "asset_preview_street_night",
          status: AssetStatus.AVAILABLE,
          tags: ["核心场景", "雨夜街道", "SSR反射", "体积光"],
          created_at: new Date(Date.now() - 43200000).toISOString(),
          updated_at: "2小时前",
          description: "雨夜赛博都市核心街景，包含高细节湿漉路面贴图、霓虹招牌反射着色与体积雾灯光预设",
        },
        {
          asset_id: "asset_default_004",
          name: "cinematic_dolly_zoom_vertigo.json",
          asset_type: AssetType.SHOT_PRESET,
          file_format: "json",
          file_size_bytes: 18 * 1024,
          storage_uri: "blender://assets/presets/cinematic_dolly_zoom_vertigo.json",
          thumbnail_uri: "asset_preview_dolly_zoom",
          status: AssetStatus.AVAILABLE,
          tags: ["推拉变焦", "DollyZoom", "眩晕透视", "120帧"],
          created_at: new Date(Date.now() - 86400000).toISOString(),
          updated_at: "4小时前",
          description: "经典希区柯克推拉变焦运镜预设，相机前进同时同步逆向平滑调整视场角 (35mm->85mm)，营造强烈眩晕透视畸变",
          metadata: {
            duration_frames: 120,
            fps: 24,
            fov_range: "35mm - 85mm",
          },
        },
        {
          asset_id: "asset_default_005",
          name: "dynamic_orbit_low_angle_hero.py",
          asset_type: AssetType.SHOT_PRESET,
          file_format: "py",
          file_size_bytes: 12 * 1024,
          storage_uri: "blender://assets/presets/dynamic_orbit_low_angle_hero.py",
          thumbnail_uri: "asset_preview_orbit_hero",
          status: AssetStatus.AVAILABLE,
          tags: ["360环绕", "低仰角", "主角运镜", "曲线平滑"],
          created_at: new Date(Date.now() - 172800000).toISOString(),
          updated_at: "8小时前",
          description: "低角度主角环绕轨迹脚本，以主角胸口为注视点，低仰角 15° 完成 360 度动态匀速圆周运镜",
          metadata: {
            duration_frames: 180,
            fps: 24,
            fov_range: "50mm Fixed",
          },
        },
        {
          asset_id: "asset_default_006",
          name: "rainy_tokyo_cyber_8k.hdr",
          asset_type: AssetType.MATERIAL,
          file_format: "hdr",
          file_size_bytes: 96 * 1024 * 1024,
          storage_uri: "blender://assets/materials/rainy_tokyo_cyber_8k.hdr",
          thumbnail_uri: "asset_preview_hdri_tokyo",
          status: AssetStatus.AVAILABLE,
          tags: ["8K HDRI", "夜景环境光", "PBR反射", "全景照明"],
          created_at: new Date(Date.now() - 259200000).toISOString(),
          updated_at: "16小时前",
          description: "32-bit 浮点高动态范围环境贴图，采集自雨夜都市霓虹灯光，提供全方位物理真实环境反射与天空照明",
        },
        {
          asset_id: "asset_default_007",
          name: "combat_dodge_roll_rootmotion.abc",
          asset_type: AssetType.ANIMATION,
          file_format: "abc",
          file_size_bytes: 145 * 1024 * 1024,
          storage_uri: "blender://assets/animations/combat_dodge_roll_rootmotion.abc",
          thumbnail_uri: "asset_preview_anim_roll",
          status: AssetStatus.AVAILABLE,
          tags: ["Alembic序列", "动作捕捉", "根骨骼位移", "战术动作"],
          created_at: new Date(Date.now() - 345600000).toISOString(),
          updated_at: "24小时前",
          description: "Alembic 格式战术侧滚翻根骨骼位移缓存序列，支持无缝导入并绑定至 Rigify 人型角色",
          metadata: {
            duration_frames: 75,
            fps: 30,
          },
        },
      ];

      let filtered = allAssets;
      if (req.asset_type) {
        filtered = filtered.filter((a) => a.asset_type === req.asset_type);
      }
      if (req.query_keyword) {
        const kw = req.query_keyword.toLowerCase();
        filtered = filtered.filter(
          (a) =>
            a.name.toLowerCase().includes(kw) ||
            (a.description && a.description.toLowerCase().includes(kw)) ||
            (a.tags && a.tags.some((t: string) => t.toLowerCase().includes(kw)))
        );
      }

      return {
        assets: filtered,
        total_count: filtered.length,
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
        total_assets: 7,
        total_models: 3,
        total_presets: 2,
        total_storage_bytes: 992 * 1024 * 1024 + 30720,
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

  static async uploadAssetFile(file: File, name?: string, userId?: string): Promise<AssetView> {
    if (USE_MOCK_API) {
      const isChar =
        file.name.toLowerCase().includes("char") ||
        file.name.toLowerCase().includes("hero") ||
        file.name.toLowerCase().includes("eva") ||
        file.name.toLowerCase().includes("girl");

      return {
        asset_id: `asset_upload_${Date.now()}`,
        name: name || file.name,
        asset_type: AssetType.MODEL_3D,
        file_format: "blend",
        file_size_bytes: file.size || 52 * 1024 * 1024,
        storage_uri: `data/assets/uploads/${file.name}`,
        thumbnail_uri: isChar ? "asset_preview_heroine_eva" : "asset_preview_street_night",
        status: AssetStatus.AVAILABLE,
        tags: isChar
          ? ["主角人物", "Rigify骨骼", "次世代PBR", "本地上传"]
          : ["本地上传", "3D模型"],
        created_at: new Date().toISOString(),
        updated_at: "刚刚",
        description: isChar
          ? `自动解析的人型角色高模（${name || file.name}），已绑定 Rigify 骨骼体系并支持 ARKit 面部表情驱动，面数 138,400`
          : `本地上传的 Blender 工程模型，面数 114,600`,
        metadata: {
          poly_count: isChar ? 138400 : 114600,
          vertex_count: isChar ? 98200 : 88100,
          rig_type: isChar ? "Rigify Humanoid (Blender 4.x)" : "None",
          has_face_rig: isChar,
        },
      };
    }

    const formData = new FormData();
    formData.append("file", file);
    if (name) formData.append("name", name);
    formData.append("user_id", userId || "default_user_001");

    const res = await fetch("/api/v0_1/assets/upload", {
      method: "POST",
      body: formData,
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "上传失败" }));
      throw new Error(err.error || `Upload failed with status ${res.status}`);
    }

    return (await res.json()) as AssetView;
  }
}
