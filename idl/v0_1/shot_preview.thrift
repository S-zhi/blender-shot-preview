namespace go handler.v0_1

enum TaskStatus {
    ACCEPTED = 1
    REJECTED = 2
}

enum ProductionTaskStatus {
    PENDING = 1
    RUNNING = 2
    SUCCEEDED = 3
    FAILED = 4
    CANCELLED = 5
}

struct TaskFailure {
    1: required string code
    2: required string message
    3: required bool retryable
}

struct TaskNodeView {
    1: required string node_id
    2: required ProductionTaskStatus status
    3: required i32 attempts
    4: optional string started_at
    5: optional string finished_at
    6: optional TaskFailure failure
}

struct TaskArtifactView {
    1: required string type
    2: required string uri
    3: required string name
}

struct ShotPreviewTaskView {
    1: required string task_id
    2: required ProductionTaskStatus status
    3: required string workflow_id
    4: required string workflow_version
    5: required list<TaskNodeView> nodes
    6: required list<TaskArtifactView> artifacts
    7: optional TaskFailure failure
    8: required string created_at
    9: required string updated_at
    10: optional string finished_at
}

struct CreateShotPreviewTaskRequest {
    1: required string user_id
    2: required string prompt
    3: optional string conversation_id
    4: optional string request_id
    5: optional string workflow_id
    6: optional string workflow_version
}

struct CreateShotPreviewTaskResponse {
    1: required string task_id
    2: required TaskStatus status
    3: optional string request_id
    4: optional bool replayed
}

struct GetShotPreviewTaskRequest {
    1: required string user_id
    2: required string task_id
}

struct GetShotPreviewTaskResponse {
    1: required ShotPreviewTaskView task
}

struct CancelShotPreviewTaskRequest {
    1: required string user_id
    2: required string task_id
}

struct CancelShotPreviewTaskResponse {
    1: required ShotPreviewTaskView task
}

struct RetryShotPreviewTaskRequest {
    1: required string user_id
    2: required string task_id
}

struct RetryShotPreviewTaskResponse {
    1: required ShotPreviewTaskView task
}

service ShotPreviewServiceV0_1 {
    CreateShotPreviewTaskResponse CreateShotPreviewTask(1: CreateShotPreviewTaskRequest request)
    GetShotPreviewTaskResponse GetShotPreviewTask(1: GetShotPreviewTaskRequest request)
    CancelShotPreviewTaskResponse CancelShotPreviewTask(1: CancelShotPreviewTaskRequest request)
    RetryShotPreviewTaskResponse RetryShotPreviewTask(1: RetryShotPreviewTaskRequest request)
}
