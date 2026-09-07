namespace go handler.v0_1

enum LLMProvider {
    OPENAI = 1
    ANTHROPIC = 2
}

enum LLMKeyOperation {
    SAVE = 1
    UPDATE = 2
    DELETE = 3
}

struct ManageLLMKeyRequest {
    1: required string user_id
    2: required LLMKeyOperation operation
    3: optional string key_id
    4: optional LLMProvider provider
    5: optional string name
    6: optional string api_key
    7: optional string base_url
}

struct ManageLLMKeyResponse {
    1: required string key_id
    2: optional LLMProvider provider
}

service LLMKeyServiceV0_1 {
    ManageLLMKeyResponse ManageLLMKey(1: ManageLLMKeyRequest request)
}
