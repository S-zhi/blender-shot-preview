namespace go handler.v0_1

enum AssetType {
    MODEL_3D = 1,      // 3D模型 (.blend, .fbx, .obj)
    SHOT_PRESET = 2,   // 镜头轨迹预设 (.json, .py)
    MATERIAL = 3,      // 材质与贴图 (.png, .exr, .hdr)
    ANIMATION = 4      // 动画序列 (.bvh, .abc)
}

enum AssetStatus {
    AVAILABLE = 1,     // 可用就绪
    PROCESSING = 2,    // 解析/预渲染中
    ARCHIVED = 3       // 已归档
}

struct AssetView {
    1: required string asset_id
    2: required string name
    3: required AssetType asset_type
    4: required string file_format      // 例如: blend, fbx, json
    5: required i64 file_size_bytes     // 文件字节大小
    6: required string storage_uri      // 本地存储路径或 OSS URI
    7: optional string thumbnail_uri    // 缩略图路径
    8: required AssetStatus status
    9: optional list<string> tags       // 标签，如 ["cyberpunk", "wide-angle"]
    10: required string created_at
    11: required string updated_at
    12: optional string description
}

// 1. 获取资产列表
struct ListAssetsRequest {
    1: required string user_id
    2: optional AssetType asset_type
    3: optional string query_keyword    // 模糊搜索关键词
    4: optional i32 page_size
    5: optional i32 page_num
}

struct ListAssetsResponse {
    1: required list<AssetView> assets
    2: required i64 total_count
}

// 2. 上传/注册新资产
struct RegisterAssetRequest {
    1: required string user_id
    2: required string name
    3: required AssetType asset_type
    4: required string file_format
    5: required i64 file_size_bytes
    6: required string storage_uri
    7: optional string description
    8: optional list<string> tags
}

struct RegisterAssetResponse {
    1: required string asset_id
    2: required AssetStatus status
}

// 3. 删除资产
struct DeleteAssetRequest {
    1: required string user_id
    2: required string asset_id
}

struct DeleteAssetResponse {
    1: required bool success
}

// 4. 资产看板概览指标统计
struct GetAssetStatsRequest {
    1: required string user_id
}

struct GetAssetStatsResponse {
    1: required i64 total_assets
    2: required i64 total_models
    3: required i64 total_presets
    4: required i64 total_storage_bytes
}

service AssetServiceV0_1 {
    ListAssetsResponse ListAssets(1: ListAssetsRequest request)
    RegisterAssetResponse RegisterAsset(1: RegisterAssetRequest request)
    DeleteAssetResponse DeleteAsset(1: DeleteAssetRequest request)
    GetAssetStatsResponse GetAssetStats(1: GetAssetStatsRequest request)
}
