package agent

import "errors"

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
