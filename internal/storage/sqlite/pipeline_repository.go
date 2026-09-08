package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

type PipelineRepository struct{ db *sql.DB }

func NewPipelineRepository(db *sql.DB) *PipelineRepository { return &PipelineRepository{db: db} }

func (r *PipelineRepository) CreateIfAbsent(ctx context.Context, task pipeline.Task) (pipeline.Task, bool, error) {
	if r == nil || r.db == nil {
		return pipeline.Task{}, false, fmt.Errorf("sqlite pipeline repository is unavailable")
	}
	payload, err := json.Marshal(task)
	if err != nil {
		return pipeline.Task{}, false, fmt.Errorf("encode pipeline task: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return pipeline.Task{}, false, err
	}
	defer tx.Rollback()
	var id, existingPayload string
	err = tx.QueryRowContext(ctx, `SELECT task_id, task_json FROM pipeline_tasks WHERE idempotency_key = ?`, task.IdempotencyKey).Scan(&id, &existingPayload)
	if err == nil {
		var existing pipeline.Task
		if json.Unmarshal([]byte(existingPayload), &existing) != nil {
			return pipeline.Task{}, false, fmt.Errorf("decode stored pipeline task %q", id)
		}
		if !bytes.Equal(existing.Input.JSON, task.Input.JSON) {
			return pipeline.Task{}, false, pipeline.ErrIdempotencyConflict
		}
		return existing, false, nil
	}
	if err != sql.ErrNoRows {
		return pipeline.Task{}, false, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO pipeline_tasks(task_id,idempotency_key,user_id,conversation_id,status,task_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`,
		task.ID, task.IdempotencyKey, taskUser(task), taskConversation(task), task.Status, string(payload), task.CreatedAt.UTC().Format(time.RFC3339Nano), task.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return pipeline.Task{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return pipeline.Task{}, false, err
	}
	return task, true, nil
}

func (r *PipelineRepository) Get(ctx context.Context, taskID string) (pipeline.Task, error) {
	var payload string
	err := r.db.QueryRowContext(ctx, `SELECT task_json FROM pipeline_tasks WHERE task_id = ?`, taskID).Scan(&payload)
	if err == sql.ErrNoRows {
		return pipeline.Task{}, pipeline.ErrTaskNotFound
	}
	if err != nil {
		return pipeline.Task{}, err
	}
	var task pipeline.Task
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return pipeline.Task{}, fmt.Errorf("decode pipeline task: %w", err)
	}
	return task, nil
}

func (r *PipelineRepository) Update(ctx context.Context, task pipeline.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE pipeline_tasks SET status=?, task_json=?, updated_at=? WHERE task_id=?`, task.Status, string(payload), task.UpdatedAt.UTC().Format(time.RFC3339Nano), task.ID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return pipeline.ErrTaskNotFound
	}
	return err
}

func (r *PipelineRepository) ListIncomplete(ctx context.Context) ([]pipeline.Task, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT task_json FROM pipeline_tasks WHERE status NOT IN ('succeeded','failed','cancelled') ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []pipeline.Task
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var task pipeline.Task
		if err := json.Unmarshal([]byte(payload), &task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].CreatedAt.Before(tasks[j].CreatedAt) })
	return tasks, nil
}

func (r *PipelineRepository) AppendEvent(ctx context.Context, event pipeline.PipelineEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO pipeline_events(task_id,sequence,event_type,payload_json,created_at) SELECT ?,COALESCE(MAX(sequence),0)+1,?,?,? FROM pipeline_events WHERE task_id=?`, event.TaskID, event.Type, string(payload), event.Timestamp.UTC().Format(time.RFC3339Nano), event.TaskID)
	return err
}

// These fields are also present in the task input snapshot. Keeping them in
// dedicated columns makes the database inspectable without coupling Task to a
// storage-specific shape.
func taskUser(task pipeline.Task) string {
	var v struct {
		UserID string `json:"user_id"`
	}
	_ = json.Unmarshal(task.Input.JSON, &v)
	return v.UserID
}

func taskConversation(task pipeline.Task) string {
	var v struct {
		ConversationID string `json:"conversation_id"`
	}
	_ = json.Unmarshal(task.Input.JSON, &v)
	return v.ConversationID
}
