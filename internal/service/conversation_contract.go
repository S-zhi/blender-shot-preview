package service

import (
	"context"
	"time"
)

// Conversation and Message are the durable history boundary used by the web
// client. The persistence implementation deliberately lives outside service.
type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	ID          string         `json:"id"`
	Role        string         `json:"role"`
	Content     string         `json:"content"`
	Status      string         `json:"status,omitempty"`
	TaskID      string         `json:"task_id,omitempty"`
	Thoughts    []string       `json:"thoughts,omitempty"`
	Nodes       []NodeView     `json:"nodes,omitempty"`
	Artifacts   []ArtifactView `json:"artifacts,omitempty"`
	WaitingNode *NodeView      `json:"waiting_node,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ConversationStore is intentionally small so it can be backed by SQLite,
// another database, or a test double without changing the HTTP contract.
type ConversationStore interface {
	List(ctx context.Context, userID string) ([]Conversation, error)
	Get(ctx context.Context, userID, conversationID string) (Conversation, error)
	Ensure(ctx context.Context, userID, conversationID, title string) (Conversation, error)
	CreateMessage(ctx context.Context, conversationID string, message Message) error
	UpdateMessage(ctx context.Context, message Message) error
	UpdateTask(ctx context.Context, taskID string, task TaskView) error
	Delete(ctx context.Context, userID, conversationID string) error
}
