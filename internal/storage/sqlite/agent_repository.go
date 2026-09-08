package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
)

type AgentDefinitionRepository struct{ db *sql.DB }

func NewAgentDefinitionRepository(db *sql.DB) *AgentDefinitionRepository {
	return &AgentDefinitionRepository{db: db}
}

func (r *AgentDefinitionRepository) Create(ctx context.Context, definition agent.AgentDefinition) error {
	payload, err := json.Marshal(definition)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO agent_definitions(id,definition_json,created_at,updated_at) VALUES(?,?,?,?)`, definition.ID, string(payload), definition.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"), definition.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"))
	if err != nil {
		return fmt.Errorf("create agent definition: %w", err)
	}
	return nil
}

func (r *AgentDefinitionRepository) Get(ctx context.Context, id string) (agent.AgentDefinition, error) {
	var payload string
	if err := r.db.QueryRowContext(ctx, `SELECT definition_json FROM agent_definitions WHERE id=?`, id).Scan(&payload); err == sql.ErrNoRows {
		return agent.AgentDefinition{}, agent.ErrAgentNotFound
	} else if err != nil {
		return agent.AgentDefinition{}, err
	}
	var definition agent.AgentDefinition
	if err := json.Unmarshal([]byte(payload), &definition); err != nil {
		return agent.AgentDefinition{}, err
	}
	return definition, nil
}

func (r *AgentDefinitionRepository) List(ctx context.Context) ([]agent.AgentDefinition, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT definition_json FROM agent_definitions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var definitions []agent.AgentDefinition
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var definition agent.AgentDefinition
		if err := json.Unmarshal([]byte(payload), &definition); err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	return definitions, rows.Err()
}

func (r *AgentDefinitionRepository) Update(ctx context.Context, definition agent.AgentDefinition) error {
	payload, err := json.Marshal(definition)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE agent_definitions SET definition_json=?, updated_at=? WHERE id=?`, string(payload), definition.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"), definition.ID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return agent.ErrAgentNotFound
	}
	return nil
}

type AgentRunRepository struct{ db *sql.DB }

func NewAgentRunRepository(db *sql.DB) *AgentRunRepository { return &AgentRunRepository{db: db} }

func (r *AgentRunRepository) Create(ctx context.Context, run agent.AgentRun) error {
	payload, err := json.Marshal(run)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO agent_runs(run_id,run_json) VALUES(?,?)`, run.ID, string(payload))
	return err
}

func (r *AgentRunRepository) Get(ctx context.Context, id string) (agent.AgentRun, error) {
	var payload string
	if err := r.db.QueryRowContext(ctx, `SELECT run_json FROM agent_runs WHERE run_id=?`, id).Scan(&payload); err == sql.ErrNoRows {
		return agent.AgentRun{}, agent.ErrRunNotFound
	} else if err != nil {
		return agent.AgentRun{}, err
	}
	var run agent.AgentRun
	if err := json.Unmarshal([]byte(payload), &run); err != nil {
		return agent.AgentRun{}, err
	}
	return run, nil
}

func (r *AgentRunRepository) Update(ctx context.Context, run agent.AgentRun) error {
	payload, err := json.Marshal(run)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE agent_runs SET run_json=? WHERE run_id=?`, string(payload), run.ID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return agent.ErrRunNotFound
	}
	return nil
}

func (r *AgentRunRepository) AppendEvent(ctx context.Context, event agent.AgentEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT OR REPLACE INTO agent_events(run_id,sequence,event_json) VALUES(?,?,?)`, event.RunID, event.Sequence, string(payload))
	return err
}

func (r *AgentRunRepository) ListEvents(ctx context.Context, runID string) ([]agent.AgentEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT event_json FROM agent_events WHERE run_id=? ORDER BY sequence`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []agent.AgentEvent
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var event agent.AgentEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
