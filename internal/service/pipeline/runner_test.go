package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRunnerHonorsDependenciesAndRunsReadyNodesConcurrently(t *testing.T) {
	started := make(chan NodeID, 2)
	release := make(chan struct{})
	var (
		mu       sync.Mutex
		calls    []NodeID
		seenDeps map[NodeID]DependencyOutputs
	)
	runner := NewRunner(NewMemoryRepository(), nil, StepInvokerFunc(func(_ context.Context, request StepRequest) (json.RawMessage, error) {
		mu.Lock()
		calls = append(calls, request.NodeID)
		if seenDeps == nil {
			seenDeps = make(map[NodeID]DependencyOutputs)
		}
		seenDeps[request.NodeID] = request.Dependencies
		mu.Unlock()
		if request.NodeID == "B" || request.NodeID == "C" {
			started <- request.NodeID
			<-release
		}
		return json.RawMessage(fmt.Sprintf(`{"node":%q}`, request.NodeID)), nil
	}), nil)

	task, created, err := runner.Submit(context.Background(), Submission{
		IdempotencyKey: "dependency-parallel",
		Input:          snapshot(t, map[string]string{"request": "one"}),
		Workflow: Workflow{Nodes: []NodeSpec{
			stepNode("A", nil, 1),
			stepNode("B", []NodeID{"A"}, 1),
			stepNode("C", []NodeID{"A"}, 1),
			stepNode("D", []NodeID{"B", "C"}, 1),
		}},
	})
	if err != nil || !created {
		t.Fatalf("Submit() = task=%#v created=%v err=%v", task, created, err)
	}

	got := map[NodeID]bool{}
	for range 2 {
		select {
		case id := <-started:
			got[id] = true
		case <-time.After(time.Second):
			t.Fatal("ready branches did not start concurrently")
		}
	}
	if !got["B"] || !got["C"] {
		t.Fatalf("parallel starts = %#v, want B and C", got)
	}
	close(release)
	completed := waitForTask(t, runner, task.ID, func(task Task) bool { return task.Status == TaskStatusSucceeded })
	if node, ok := completed.Node("D"); !ok || node.Status != NodeStatusSucceeded {
		t.Fatalf("D = %#v, want succeeded", node)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 4 || calls[0] != "A" || calls[len(calls)-1] != "D" {
		t.Fatalf("execution order = %#v, want A then B/C then D", calls)
	}
	if _, exists := seenDeps["B"]["A"]; !exists {
		t.Fatalf("B dependencies = %#v, want A output", seenDeps["B"])
	}
	if len(seenDeps["D"]) != 2 {
		t.Fatalf("D dependencies = %#v, want B and C outputs", seenDeps["D"])
	}
}

func TestRunnerDeduplicatesIdempotentSubmission(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	runner := NewRunner(NewMemoryRepository(), nil, StepInvokerFunc(func(context.Context, StepRequest) (json.RawMessage, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return json.RawMessage(`{"ok":true}`), nil
	}), nil)
	submission := Submission{
		IdempotencyKey: "same-request",
		Input:          snapshot(t, map[string]string{"request": "one"}),
		Workflow:       Workflow{Nodes: []NodeSpec{stepNode("only", nil, 1)}},
	}
	first, created, err := runner.Submit(context.Background(), submission)
	if err != nil || !created {
		t.Fatalf("first Submit() = created=%v err=%v", created, err)
	}
	second, created, err := runner.Submit(context.Background(), submission)
	if err != nil || created {
		t.Fatalf("second Submit() = created=%v err=%v", created, err)
	}
	if first.ID != second.ID {
		t.Fatalf("task IDs = %q and %q, want idempotent ID", first.ID, second.ID)
	}
	waitForTask(t, runner, first.ID, func(task Task) bool { return task.Status == TaskStatusSucceeded })
	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("step calls = %d, want 1", calls)
	}
}

func TestRunnerRetriesFailedNode(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	runner := NewRunner(NewMemoryRepository(), nil, StepInvokerFunc(func(context.Context, StepRequest) (json.RawMessage, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if calls == 1 {
			return nil, errors.New("temporary failure")
		}
		return json.RawMessage(`{"ok":true}`), nil
	}), nil)
	task, _, err := runner.Submit(context.Background(), Submission{
		IdempotencyKey: "retry",
		Input:          snapshot(t, map[string]string{"request": "one"}),
		Workflow:       Workflow{Nodes: []NodeSpec{stepNode("retry", nil, 2)}},
	})
	if err != nil {
		t.Fatalf("Submit(): %v", err)
	}
	completed := waitForTask(t, runner, task.ID, func(task Task) bool { return task.Status == TaskStatusSucceeded })
	node, exists := completed.Node("retry")
	if !exists || node.Attempts != 2 || node.Status != NodeStatusSucceeded {
		t.Fatalf("retry node = %#v, want succeeded after two attempts", node)
	}
	mu.Lock()
	defer mu.Unlock()
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestRunnerCanRetryFailedTask(t *testing.T) {
	var (
		mu      sync.Mutex
		allowed bool
	)
	runner := NewRunner(NewMemoryRepository(), nil, StepInvokerFunc(func(context.Context, StepRequest) (json.RawMessage, error) {
		mu.Lock()
		defer mu.Unlock()
		if !allowed {
			return nil, errors.New("operator action required")
		}
		return json.RawMessage(`{"ok":true}`), nil
	}), nil)
	task, _, err := runner.Submit(context.Background(), Submission{
		IdempotencyKey: "manual-retry",
		Input:          snapshot(t, map[string]string{"request": "one"}),
		Workflow:       Workflow{Nodes: []NodeSpec{stepNode("retry", nil, 1)}},
	})
	if err != nil {
		t.Fatalf("Submit(): %v", err)
	}
	waitForTask(t, runner, task.ID, func(task Task) bool { return task.Status == TaskStatusFailed })
	mu.Lock()
	allowed = true
	mu.Unlock()
	if _, err := runner.Retry(context.Background(), task.ID); err != nil {
		t.Fatalf("Retry(): %v", err)
	}
	completed := waitForTask(t, runner, task.ID, func(task Task) bool { return task.Status == TaskStatusSucceeded })
	node, exists := completed.Node("retry")
	if !exists || node.Status != NodeStatusSucceeded || node.Attempts != 1 {
		t.Fatalf("retried node = %#v, want fresh successful attempt", node)
	}
}

func TestRunnerCancelsRunningTask(t *testing.T) {
	entered := make(chan struct{})
	runner := NewRunner(NewMemoryRepository(), nil, StepInvokerFunc(func(ctx context.Context, _ StepRequest) (json.RawMessage, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}), nil)
	task, _, err := runner.Submit(context.Background(), Submission{
		IdempotencyKey: "cancel",
		Input:          snapshot(t, map[string]string{"request": "one"}),
		Workflow:       Workflow{Nodes: []NodeSpec{stepNode("blocking", nil, 1)}},
	})
	if err != nil {
		t.Fatalf("Submit(): %v", err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("step did not start")
	}
	if err := runner.Cancel(context.Background(), task.ID); err != nil {
		t.Fatalf("Cancel(): %v", err)
	}
	cancelled := waitForTask(t, runner, task.ID, func(task Task) bool { return task.Status == TaskStatusCancelled })
	node, exists := cancelled.Node("blocking")
	if !exists || node.Status != NodeStatusCancelled {
		t.Fatalf("cancelled node = %#v, want cancelled", node)
	}
}

func TestRunnerRecoversPersistedRunningNode(t *testing.T) {
	repository := NewMemoryRepository()
	input := snapshot(t, map[string]string{"request": "recover"})
	now := time.Now().UTC()
	stored, created, err := repository.CreateIfAbsent(context.Background(), Task{
		ID: "recovered", IdempotencyKey: "recover", Input: input, Status: TaskStatusRunning,
		CreatedAt: now, UpdatedAt: now,
		Nodes: []Node{{
			ID: "interrupted", Invocation: Invocation{Kind: InvocationStep, Target: "test"},
			Input: input, Status: NodeStatusRunning, Attempts: 1, MaxAttempts: 2, StartedAt: &now,
		}},
	})
	if err != nil || !created || stored.ID != "recovered" {
		t.Fatalf("seed task = %#v created=%v err=%v", stored, created, err)
	}
	var calls int
	runner := NewRunner(repository, nil, StepInvokerFunc(func(_ context.Context, request StepRequest) (json.RawMessage, error) {
		calls++
		if request.Attempt != 2 {
			return nil, fmt.Errorf("attempt = %d, want resumed attempt 2", request.Attempt)
		}
		return json.RawMessage(`{"recovered":true}`), nil
	}), nil)
	if err := runner.Recover(context.Background()); err != nil {
		t.Fatalf("Recover(): %v", err)
	}
	completed := waitForTask(t, runner, "recovered", func(task Task) bool { return task.Status == TaskStatusSucceeded })
	node, exists := completed.Node("interrupted")
	if !exists || node.Status != NodeStatusSucceeded || node.Attempts != 2 || calls != 1 {
		t.Fatalf("recovered node = %#v calls=%d", node, calls)
	}
}

func stepNode(id NodeID, dependencies []NodeID, attempts int) NodeSpec {
	return NodeSpec{
		ID: id, DependsOn: dependencies, MaxAttempts: attempts,
		Invocation: Invocation{Kind: InvocationStep, Target: "test"},
	}
}

func snapshot(t *testing.T, value any) Snapshot {
	t.Helper()
	snapshot, err := NewSnapshot(value)
	if err != nil {
		t.Fatalf("NewSnapshot(): %v", err)
	}
	return snapshot
}

func waitForTask(t *testing.T, runner *Runner, taskID string, predicate func(Task) bool) Task {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		task, err := runner.Get(context.Background(), taskID)
		if err != nil {
			t.Fatalf("Get(%q): %v", taskID, err)
		}
		if predicate(task) {
			return task
		}
		time.Sleep(5 * time.Millisecond)
	}
	task, err := runner.Get(context.Background(), taskID)
	t.Fatalf("task %q did not reach expected state, last=%#v err=%v", taskID, task, err)
	return Task{}
}
