package sqlite

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	"github.com/S-zhi/blender-shot-preview/internal/service/pipeline"
)

func TestSQLitePersistsPipelineAndConversationAcrossHandles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "shot-preview.sqlite")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := NewPipelineRepository(db)
	input, err := pipeline.NewSnapshot(map[string]string{"user_id": "user-1", "conversation_id": "conv-1", "prompt": "test"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	task := pipeline.Task{ID: "task-1", IdempotencyKey: "user-1:req-1", Input: input, Status: pipeline.TaskStatusPending, CreatedAt: now, UpdatedAt: now, Nodes: []pipeline.Node{{ID: "Initialize", Status: pipeline.NodeStatusPending, MaxAttempts: 1, Invocation: pipeline.Invocation{Kind: pipeline.InvocationStep, Target: "Initialize"}, Input: input}}}
	if _, created, err := repo.CreateIfAbsent(ctx, task); err != nil || !created {
		t.Fatalf("create task: created=%v err=%v", created, err)
	}
	if err := repo.Update(ctx, task); err != nil {
		t.Fatal(err)
	}
	conversation := NewConversationStore(db)
	if _, err := conversation.Ensure(ctx, "user-1", "conv-1", "test conversation"); err != nil {
		t.Fatal(err)
	}
	if err := conversation.CreateMessage(ctx, "conv-1", service.Message{ID: "msg-1", Role: "user", Content: "test", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := conversation.CreateMessage(ctx, "conv-1", service.Message{ID: "msg-2", Role: "assistant", Status: "thought", TaskID: "task-1", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo = NewPipelineRepository(db)
	loaded, err := repo.Get(ctx, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != task.ID || loaded.Input.JSON == nil {
		t.Fatalf("loaded task = %#v", loaded)
	}
	conversations := NewConversationStore(db)
	history, err := conversations.List(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || len(history[0].Messages) != 2 {
		t.Fatalf("history = %#v", history)
	}
	if got, _ := json.Marshal(history[0]); !json.Valid(got) {
		t.Fatal("history is not valid JSON")
	}
}

func TestSQLiteBackedShotPreviewServicePersistsEveryQuestion(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "shot-preview.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	runner := pipeline.NewRunner(
		NewPipelineRepository(db),
		pipeline.AgentInvokerFunc(func(_ context.Context, _ pipeline.AgentRequest) (json.RawMessage, error) {
			return json.RawMessage(`{"ok":true}`), nil
		}),
		pipeline.StepInvokerFunc(func(_ context.Context, _ pipeline.StepRequest) (json.RawMessage, error) {
			return json.RawMessage(`{"ok":true}`), nil
		}),
		pipeline.ToolInvokerFunc(func(_ context.Context, _ pipeline.ToolRequest) (json.RawMessage, error) {
			return json.RawMessage(`{"ok":true}`), nil
		}),
	)
	serviceImpl := service.NewShotPreviewServiceWithConversations(runner, NewConversationStore(db))
	result, err := serviceImpl.CreateTask(context.Background(), service.CreateTaskRequest{UserID: "user-1", Prompt: "persist me", ConversationID: "conv-1", RequestID: "req-1"})
	if err != nil || result.TaskID == "" {
		t.Fatalf("CreateTask = %#v, err=%v", result, err)
	}
	history, err := NewConversationStore(db).List(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || len(history[0].Messages) != 2 {
		t.Fatalf("history = %#v", history)
	}
	if history[0].Messages[0].Content != "persist me" || history[0].Messages[1].TaskID != result.TaskID {
		t.Fatalf("messages = %#v", history[0].Messages)
	}
}
