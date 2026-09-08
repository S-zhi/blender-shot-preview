package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service"
)

type ConversationStore struct{ db *sql.DB }

func NewConversationStore(db *sql.DB) *ConversationStore { return &ConversationStore{db: db} }

func (s *ConversationStore) List(ctx context.Context, userID string) ([]service.Conversation, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,user_id,title,created_at,updated_at FROM conversations WHERE user_id=? ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	var conversations []service.Conversation
	for rows.Next() {
		var c service.Conversation
		var created, updated string
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt, err = parseTime(created)
		if err != nil {
			return nil, err
		}
		c.UpdatedAt, err = parseTime(updated)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for index := range conversations {
		conversations[index].Messages, err = s.messages(ctx, conversations[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return conversations, nil
}

func (s *ConversationStore) Get(ctx context.Context, userID, conversationID string) (service.Conversation, error) {
	var c service.Conversation
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,user_id,title,created_at,updated_at FROM conversations WHERE id=? AND user_id=?`, conversationID, userID).Scan(&c.ID, &c.UserID, &c.Title, &created, &updated)
	if err == sql.ErrNoRows {
		return service.Conversation{}, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return service.Conversation{}, err
	}
	c.CreatedAt, err = parseTime(created)
	if err != nil {
		return c, err
	}
	c.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return c, err
	}
	c.Messages, err = s.messages(ctx, c.ID)
	return c, err
}

func (s *ConversationStore) Ensure(ctx context.Context, userID, conversationID, title string) (service.Conversation, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	title = strings.TrimSpace(title)
	if title == "" {
		title = "新的分镜会话"
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO conversations(id,user_id,title,created_at,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET title=excluded.title,updated_at=excluded.updated_at WHERE conversations.user_id=excluded.user_id`, conversationID, userID, title, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano))
	if err != nil {
		return service.Conversation{}, err
	}
	return s.Get(ctx, userID, conversationID)
}

func (s *ConversationStore) CreateMessage(ctx context.Context, conversationID string, message service.Message) error {
	thoughts, _ := json.Marshal(message.Thoughts)
	nodes, _ := json.Marshal(message.Nodes)
	artifacts, _ := json.Marshal(message.Artifacts)
	waiting := ""
	if message.WaitingNode != nil {
		b, _ := json.Marshal(message.WaitingNode)
		waiting = string(b)
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO messages(id,conversation_id,role,content,status,task_id,thoughts_json,nodes_json,artifacts_json,waiting_node_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, message.ID, conversationID, message.Role, message.Content, message.Status, message.TaskID, string(thoughts), string(nodes), string(artifacts), waiting, message.CreatedAt.UTC().Format(time.RFC3339Nano), message.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err == nil {
		_, err = s.db.ExecContext(ctx, `UPDATE conversations SET updated_at=? WHERE id=?`, message.UpdatedAt.UTC().Format(time.RFC3339Nano), conversationID)
	}
	return err
}

func (s *ConversationStore) UpdateMessage(ctx context.Context, message service.Message) error {
	thoughts, _ := json.Marshal(message.Thoughts)
	nodes, _ := json.Marshal(message.Nodes)
	artifacts, _ := json.Marshal(message.Artifacts)
	waiting := ""
	if message.WaitingNode != nil {
		b, _ := json.Marshal(message.WaitingNode)
		waiting = string(b)
	}
	_, err := s.db.ExecContext(ctx, `UPDATE messages SET content=?,status=?,task_id=?,thoughts_json=?,nodes_json=?,artifacts_json=?,waiting_node_json=?,updated_at=? WHERE id=?`, message.Content, message.Status, message.TaskID, string(thoughts), string(nodes), string(artifacts), waiting, message.UpdatedAt.UTC().Format(time.RFC3339Nano), message.ID)
	return err
}

func (s *ConversationStore) UpdateTask(ctx context.Context, taskID string, task service.TaskView) error {
	var id, content, status, thoughtsJSON string
	err := s.db.QueryRowContext(ctx, `SELECT id,content,status,thoughts_json FROM messages WHERE task_id=? ORDER BY created_at DESC LIMIT 1`, taskID).Scan(&id, &content, &status, &thoughtsJSON)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	var thoughts []string
	_ = json.Unmarshal([]byte(thoughtsJSON), &thoughts)
	status = "thought"
	switch task.Status {
	case service.TaskStatusSucceeded:
		status = "done"
		content = "### 🎬 分镜预览工作流已全部执行完成\n\n已成功生成分镜视频制品，可点击下方按钮下载。"
		thoughts = append(thoughts, "🎉 全流程执行成功")
	case service.TaskStatusFailed:
		status = "error"
		message := "执行过程中发生异常"
		if task.Failure != nil && task.Failure.Message != "" {
			message = task.Failure.Message
		}
		content = "### ❌ 分镜生成任务失败\n\n错误详情：" + message
		thoughts = append(thoughts, "✖ 任务失败: "+message)
	case service.TaskStatusCancelled:
		status = "error"
		content = "### ⚠️ 分镜生成任务已取消"
		thoughts = append(thoughts, "任务已取消")
	default:
		for _, node := range task.Nodes {
			if node.Status == service.NodeStatusWaitingConfirmation {
				status = "waiting_confirmation"
				thoughts = append(thoughts, "⏸ 等待节点确认: "+node.ID)
				break
			}
		}
	}
	newThoughts, _ := json.Marshal(thoughts)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `UPDATE messages SET content=?,status=?,thoughts_json=?,nodes_json=?,artifacts_json=?,updated_at=? WHERE id=?`, content, status, string(newThoughts), mustJSON(task.Nodes), mustJSON(task.Artifacts), now, id)
	return err
}

func mustJSON(value any) string { data, _ := json.Marshal(value); return string(data) }

func (s *ConversationStore) Delete(ctx context.Context, userID, conversationID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM conversations WHERE id=? AND user_id=?`, conversationID, userID)
	return err
}

func (s *ConversationStore) messages(ctx context.Context, conversationID string) ([]service.Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,role,content,status,task_id,thoughts_json,nodes_json,artifacts_json,waiting_node_json,created_at,updated_at FROM messages WHERE conversation_id=? ORDER BY created_at`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []service.Message
	for rows.Next() {
		var m service.Message
		var thoughts, nodes, artifacts, waiting, created, updated string
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Status, &m.TaskID, &thoughts, &nodes, &artifacts, &waiting, &created, &updated); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(thoughts), &m.Thoughts)
		_ = json.Unmarshal([]byte(nodes), &m.Nodes)
		_ = json.Unmarshal([]byte(artifacts), &m.Artifacts)
		if waiting != "" {
			m.WaitingNode = &service.NodeView{}
			_ = json.Unmarshal([]byte(waiting), m.WaitingNode)
		}
		m.CreatedAt, err = parseTime(created)
		if err != nil {
			return nil, err
		}
		m.UpdatedAt, err = parseTime(updated)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
