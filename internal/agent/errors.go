package agent

import (
	"errors"
	"fmt"
)

var (
	ErrAgentNotFound       = errors.New("agent not found")
	ErrAgentDisabled       = errors.New("agent is disabled")
	ErrRunNotFound         = errors.New("agent run not found")
	ErrInvalidDefinition   = errors.New("invalid agent definition")
	ErrSkillNotFound       = errors.New("skill not found")
	ErrDynamicSkillBlocked = errors.New("request-scoped skills are not enabled")
	ErrNotImplemented      = errors.New("not implemented")
	ErrRunCancelled        = errors.New("agent run cancelled")
	ErrToolCallLimit       = errors.New("tool call limit exceeded")
	ErrRepeatedToolLimit   = errors.New("repeated tool call limit exceeded")
	ErrConsecutiveFailures = errors.New("consecutive tool failures exceeded")
)

// LLMErrorCode identifies an operator-actionable model connectivity failure.
// The code is safe to expose to API clients; the wrapped cause must never
// contain credentials or raw provider response bodies.
type LLMErrorCode string

const (
	LLMNotConfigured         LLMErrorCode = "LLM_NOT_CONFIGURED"
	LLMCredentialUnavailable LLMErrorCode = "LLM_CREDENTIAL_UNAVAILABLE"
	LLMConnectionFailed      LLMErrorCode = "LLM_CONNECTION_FAILED"
	LLMTimeout               LLMErrorCode = "LLM_TIMEOUT"
	LLMAuthFailed            LLMErrorCode = "LLM_AUTH_FAILED"
	LLMRateLimited           LLMErrorCode = "LLM_RATE_LIMITED"
	LLMProviderError         LLMErrorCode = "LLM_PROVIDER_ERROR"
	LLMInvalidResponse       LLMErrorCode = "LLM_INVALID_RESPONSE"
)

// LLMError is returned at the model boundary so callers can distinguish a
// missing credential from a transient provider/network failure.
type LLMError struct {
	Code  LLMErrorCode
	Cause error
}

func (e *LLMError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Cause)
}

func (e *LLMError) Unwrap() error { return e.Cause }

func (e *LLMError) ErrorCode() string {
	if e == nil {
		return ""
	}
	return string(e.Code)
}
