namespace go handler.v0_1

enum TaskStatus {
    ACCEPTED = 1
    REJECTED = 2
}

struct CreateShotPreviewTaskRequest {
    1: required string user_id
    2: required string prompt
    3: optional string conversation_id
    4: optional string request_id
}

struct CreateShotPreviewTaskResponse {
    1: required string task_id
    2: required TaskStatus status
    3: optional string request_id
}

service ShotPreviewServiceV0_1 {
    CreateShotPreviewTaskResponse CreateShotPreviewTask(1: CreateShotPreviewTaskRequest request)
}
