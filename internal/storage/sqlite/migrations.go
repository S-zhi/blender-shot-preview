package sqlite

import (
	"database/sql"
	"fmt"
)

func migrate(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY, user_id TEXT NOT NULL, title TEXT NOT NULL,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_user_updated ON conversations(user_id, updated_at DESC);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY, conversation_id TEXT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
			role TEXT NOT NULL, content TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT '',
			task_id TEXT NOT NULL DEFAULT '', thoughts_json TEXT NOT NULL DEFAULT '[]', nodes_json TEXT NOT NULL DEFAULT '[]',
			artifacts_json TEXT NOT NULL DEFAULT '[]', waiting_node_json TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS pipeline_tasks (
			task_id TEXT PRIMARY KEY, idempotency_key TEXT NOT NULL UNIQUE, user_id TEXT NOT NULL DEFAULT '',
			conversation_id TEXT NOT NULL DEFAULT '', status TEXT NOT NULL, task_json TEXT NOT NULL,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_pipeline_tasks_status ON pipeline_tasks(status);`,
		`CREATE TABLE IF NOT EXISTS pipeline_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT, task_id TEXT NOT NULL, sequence INTEGER NOT NULL,
			event_type TEXT NOT NULL, payload_json TEXT NOT NULL, created_at TEXT NOT NULL,
			UNIQUE(task_id, sequence)
		);`,
		`CREATE TABLE IF NOT EXISTS agent_definitions (
			id TEXT PRIMARY KEY, definition_json TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS agent_runs (
			run_id TEXT PRIMARY KEY, run_json TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS agent_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT, run_id TEXT NOT NULL, sequence INTEGER NOT NULL,
			event_json TEXT NOT NULL, UNIQUE(run_id, sequence)
		);`,
		`CREATE TABLE IF NOT EXISTS assets (
			asset_id TEXT PRIMARY KEY, user_id TEXT NOT NULL, name TEXT NOT NULL, asset_type INTEGER NOT NULL,
			file_format TEXT NOT NULL, file_size_bytes INTEGER NOT NULL, storage_uri TEXT NOT NULL,
			thumbnail_uri TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL, tags_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL, description TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("sqlite migration: %w", err)
		}
	}
	return nil
}
