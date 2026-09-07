# blender-shot-preview

## LLM credential RPC

The Kitex server registers `LLMKeyServiceV0_1.ManageLLMKey` alongside the existing
shot preview service. Both OpenAI and Anthropic use the same RPC:

- `operation`: `SAVE` (1), `UPDATE` (2), or `DELETE` (3).
- `provider`: `OPENAI` (1) or `ANTHROPIC` (2); required for save.
- `user_id`: credential owner, supplied by a trusted caller after authentication.
- Save requires `name` and `api_key`; `base_url` is optional.
- Update requires `key_id` and at least one of `name`, `api_key`, `base_url`.
- Delete requires `key_id`. The provider cannot be changed on update.

Responses contain `key_id` and the provider if supplied, never the API key.
Gateway errors use Kitex business status codes 400, 403, 404, and 500.
The RPC is Thrift, not an OpenAI/Anthropic HTTP inference endpoint.

Set `LLM_GATEWAY_MASTER_KEY` to a securely supplied 32-byte key encoded as 64 hex
characters, then run `make run`. A missing or invalid key fails startup.
The current server uses encrypted, in-memory credential storage: restarting the
process loses all credentials. PostgreSQL remains a future `CredentialStore`
implementation. This version trusts `user_id`; deploy behind an authenticated
RPC boundary that validates it, not directly on an untrusted public network.

Run `make generate` with Kitex v0.16.0 to regenerate both IDLs. Verify with
`go test ./...` and `go vet ./...`.
