// Package agent provides the application boundary for configured Eino agents.
package agent

import (
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/components/model"
)

type AgentStatus string

const (
	AgentStatusEnabled  AgentStatus = "enabled"
	AgentStatusDisabled AgentStatus = "disabled"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
	RunStatusTimeout   RunStatus = "timeout"
)

type EventType string

const (
	EventAgentStarted  EventType = "agent_started"
	EventMessageDelta  EventType = "message_delta"
	EventToolStarted   EventType = "tool_call_started"
	EventToolFinished  EventType = "tool_call_finished"
	EventAgentComplete EventType = "agent_completed"
	EventAgentFailed   EventType = "agent_failed"
	EventAgentCanceled EventType = "agent_cancelled"
)

// AgentDefinition is persisted configuration. Runtime Eino objects are always
// constructed from this definition and are never stored here.
type AgentDefinition struct {
	ID              string
	Name            string
	Description     string
	SystemPrompt    string
	ModelProfile    string
	SkillIDs        []string
	InputSchema     json.RawMessage
	OutputSchema    json.RawMessage
	MaxSteps        int
	MaxOutputTokens int
	Timeout         time.Duration
	Status          AgentStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AgentRequest struct {
	AgentID   string
	SessionID string
	Input     string
	UserID    string
	TenantID  string

	// SkillIDs is reserved for request-scoped skill selection. Dynamic skills
	// remain disabled until the authorization policy in issue #6 is implemented.
	SkillIDs []string

	// Model allows callers to explicitly bind a ChatModel at the request boundary.
	// When provided, it takes precedence over the agent definition's ModelProfile.
	Model model.BaseChatModel
}

type AgentResult struct {
	RunID   string
	Status  RunStatus
	Output  string
	Error   string
	Started time.Time
	Ended   time.Time
}

type AgentRun struct {
	ID             string
	AgentID        string
	SessionID      string
	UserID         string
	TenantID       string
	Input          string
	Status         RunStatus
	Output         string
	Error          string
	ConfigSnapshot json.RawMessage
	StartedAt      time.Time
	FinishedAt     *time.Time
}

type AgentEvent struct {
	RunID     string
	Sequence  int
	Type      EventType
	Content   string
	ToolName  string
	ErrorCode string
	CreatedAt time.Time
}

type CreateAgentRequest struct {
	ID              string
	Name            string
	Description     string
	SystemPrompt    string
	ModelProfile    string
	SkillIDs        []string
	InputSchema     json.RawMessage
	OutputSchema    json.RawMessage
	MaxSteps        int
	MaxOutputTokens int
	Timeout         time.Duration
}

type UpdateAgentRequest struct {
	ID              string
	Name            string
	Description     string
	SystemPrompt    string
	ModelProfile    string
	SkillIDs        []string
	InputSchema     json.RawMessage
	OutputSchema    json.RawMessage
	MaxSteps        int
	MaxOutputTokens int
	Timeout         time.Duration
}
