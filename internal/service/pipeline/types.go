// Package pipeline provides the durable orchestration boundary for shot
// production. It intentionally does not depend on a particular agent catalog,
// database, or Blender implementation.
package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrTaskNotFound        = errors.New("pipeline task not found")
	ErrInvalidWorkflow     = errors.New("invalid pipeline workflow")
	ErrInvalidSubmission   = errors.New("invalid pipeline submission")
	ErrIdempotencyConflict = errors.New("pipeline idempotency conflict")
	ErrTaskNotFailed       = errors.New("pipeline task is not failed")
	ErrTaskAlreadyTerminal = errors.New("pipeline task is already terminal")
	ErrMissingInvoker      = errors.New("pipeline node invoker is not configured")
)

// TaskStatus is deliberately independent from the RPC acceptance status. A
// request can be accepted while its asynchronous production task is pending.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

func (s TaskStatus) Terminal() bool {
	return s == TaskStatusSucceeded || s == TaskStatusFailed || s == TaskStatusCancelled
}

type NodeStatus string

const (
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusSucceeded NodeStatus = "succeeded"
	NodeStatusFailed    NodeStatus = "failed"
	NodeStatusCancelled NodeStatus = "cancelled"
)

func (s NodeStatus) Terminal() bool {
	return s == NodeStatusSucceeded || s == NodeStatusFailed || s == NodeStatusCancelled
}

// NodeID gives dependency edges a distinct type from arbitrary strings.
type NodeID string

// InvocationKind selects the narrow runtime dependency used by a node.
type InvocationKind string

const (
	InvocationAgent InvocationKind = "agent"
	InvocationStep  InvocationKind = "step"
	InvocationTool  InvocationKind = "tool"
)

// Invocation describes a runtime capability without persisting a function or
// implementation-specific runtime object.
type Invocation struct {
	Kind   InvocationKind
	Target string
}

// Snapshot is an immutable-at-boundary JSON value. Repositories clone it on
// read/write so callers cannot mutate stored task history through a byte slice.
type Snapshot struct {
	JSON       json.RawMessage
	CapturedAt time.Time
}

// NewSnapshot captures a typed value as JSON for a task or node boundary.
func NewSnapshot[T any](value T) (Snapshot, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return Snapshot{}, fmt.Errorf("encode snapshot: %w", err)
	}
	return newSnapshot(encoded, time.Now().UTC())
}

// DecodeSnapshot decodes a snapshot into an explicitly chosen destination
// type. It keeps JSON at the persistence boundary without forcing callers to
// use untyped maps in their application code.
func DecodeSnapshot[T any](snapshot Snapshot) (T, error) {
	var value T
	if err := json.Unmarshal(snapshot.JSON, &value); err != nil {
		return value, fmt.Errorf("decode snapshot: %w", err)
	}
	return value, nil
}

func newSnapshot(raw json.RawMessage, capturedAt time.Time) (Snapshot, error) {
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	if !json.Valid(raw) {
		return Snapshot{}, fmt.Errorf("%w: snapshot must contain valid JSON", ErrInvalidSubmission)
	}
	if capturedAt.IsZero() {
		capturedAt = time.Now().UTC()
	}
	return Snapshot{JSON: append(json.RawMessage(nil), raw...), CapturedAt: capturedAt.UTC()}, nil
}

// NodeSpec defines a node before submission. Its input is copied into a
// timestamped node snapshot during Submit.
type NodeSpec struct {
	ID          NodeID
	DependsOn   []NodeID
	Input       json.RawMessage
	Invocation  Invocation
	MaxAttempts int
}

// Workflow is a dependency DAG. Nodes with all dependencies satisfied may run
// concurrently.
type Workflow struct {
	Nodes []NodeSpec
}

func (w Workflow) Validate() error {
	if len(w.Nodes) == 0 {
		return fmt.Errorf("%w: at least one node is required", ErrInvalidWorkflow)
	}
	known := make(map[NodeID]NodeSpec, len(w.Nodes))
	for _, node := range w.Nodes {
		if strings.TrimSpace(string(node.ID)) == "" {
			return fmt.Errorf("%w: node ID is required", ErrInvalidWorkflow)
		}
		if _, exists := known[node.ID]; exists {
			return fmt.Errorf("%w: duplicate node %q", ErrInvalidWorkflow, node.ID)
		}
		if node.Invocation.Kind != InvocationAgent && node.Invocation.Kind != InvocationStep && node.Invocation.Kind != InvocationTool {
			return fmt.Errorf("%w: node %q has unsupported invocation kind %q", ErrInvalidWorkflow, node.ID, node.Invocation.Kind)
		}
		if strings.TrimSpace(node.Invocation.Target) == "" {
			return fmt.Errorf("%w: node %q invocation target is required", ErrInvalidWorkflow, node.ID)
		}
		if node.MaxAttempts < 0 {
			return fmt.Errorf("%w: node %q max attempts cannot be negative", ErrInvalidWorkflow, node.ID)
		}
		if len(node.Input) > 0 && !json.Valid(node.Input) {
			return fmt.Errorf("%w: node %q input must be valid JSON", ErrInvalidWorkflow, node.ID)
		}
		known[node.ID] = node
	}
	for _, node := range w.Nodes {
		seenDependency := make(map[NodeID]struct{}, len(node.DependsOn))
		for _, dependency := range node.DependsOn {
			if _, exists := known[dependency]; !exists {
				return fmt.Errorf("%w: node %q depends on unknown node %q", ErrInvalidWorkflow, node.ID, dependency)
			}
			if dependency == node.ID {
				return fmt.Errorf("%w: node %q cannot depend on itself", ErrInvalidWorkflow, node.ID)
			}
			if _, duplicate := seenDependency[dependency]; duplicate {
				return fmt.Errorf("%w: node %q has duplicate dependency %q", ErrInvalidWorkflow, node.ID, dependency)
			}
			seenDependency[dependency] = struct{}{}
		}
	}
	visiting := make(map[NodeID]bool, len(w.Nodes))
	visited := make(map[NodeID]bool, len(w.Nodes))
	var visit func(NodeID) error
	visit = func(id NodeID) error {
		if visiting[id] {
			return fmt.Errorf("%w: dependency cycle includes %q", ErrInvalidWorkflow, id)
		}
		if visited[id] {
			return nil
		}
		visiting[id] = true
		for _, dependency := range known[id].DependsOn {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for _, node := range w.Nodes {
		if err := visit(node.ID); err != nil {
			return err
		}
	}
	return nil
}

// FanOutNodes is the intentionally small abstraction for asset work that is
// discovered before submission. A planner can create one child per asset while
// retaining a stable parent dependency and auditable per-item input snapshots.
func FanOutNodes(parent NodeID, template NodeSpec, inputs []json.RawMessage) ([]NodeSpec, error) {
	if strings.TrimSpace(string(parent)) == "" || strings.TrimSpace(string(template.ID)) == "" {
		return nil, fmt.Errorf("%w: fan-out parent and template ID are required", ErrInvalidWorkflow)
	}
	children := make([]NodeSpec, 0, len(inputs))
	for index, input := range inputs {
		if len(input) > 0 && !json.Valid(input) {
			return nil, fmt.Errorf("%w: fan-out input %d must be valid JSON", ErrInvalidWorkflow, index)
		}
		child := cloneNodeSpec(template)
		child.ID = NodeID(fmt.Sprintf("%s-%d", template.ID, index+1))
		child.DependsOn = append([]NodeID{parent}, child.DependsOn...)
		child.Input = append(json.RawMessage(nil), input...)
		children = append(children, child)
	}
	return children, nil
}

// Node records the durable execution state and the exact input/output snapshots
// for one workflow node.
type Node struct {
	ID          NodeID
	DependsOn   []NodeID
	Invocation  Invocation
	Input       Snapshot
	Output      *Snapshot
	Status      NodeStatus
	Attempts    int
	MaxAttempts int
	Error       string
	StartedAt   *time.Time
	FinishedAt  *time.Time
}

// Task is the persisted asynchronous production request.
type Task struct {
	ID              string
	IdempotencyKey  string
	Input           Snapshot
	Nodes           []Node
	Status          TaskStatus
	CancelRequested bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      *time.Time
}

func (t Task) Node(id NodeID) (Node, bool) {
	for _, node := range t.Nodes {
		if node.ID == id {
			return cloneNode(node), true
		}
	}
	return Node{}, false
}

// Submission accepts only data that can be persisted and recovered later.
type Submission struct {
	IdempotencyKey string
	Input          Snapshot
	Workflow       Workflow
}

// DependencyOutputs are passed by value to invokers and represent the exact
// completed outputs visible to a downstream node.
type DependencyOutputs map[NodeID]Snapshot

type AgentRequest struct {
	TaskID       string
	NodeID       NodeID
	AgentID      string
	Attempt      int
	TaskInput    Snapshot
	Input        Snapshot
	Dependencies DependencyOutputs
}

type StepRequest struct {
	TaskID       string
	NodeID       NodeID
	Step         string
	Attempt      int
	TaskInput    Snapshot
	Input        Snapshot
	Dependencies DependencyOutputs
}

type ToolRequest struct {
	TaskID       string
	NodeID       NodeID
	Tool         string
	Attempt      int
	TaskInput    Snapshot
	Input        Snapshot
	Dependencies DependencyOutputs
}

// AgentInvoker is the only LLM/agent-facing dependency of the runner.
type AgentInvoker interface {
	InvokeAgent(ctx context.Context, request AgentRequest) (json.RawMessage, error)
}

// StepInvoker is for deterministic application steps such as validation or
// encoding orchestration. It deliberately has no model dependency.
type StepInvoker interface {
	InvokeStep(ctx context.Context, request StepRequest) (json.RawMessage, error)
}

// ToolInvoker is for deterministic external tool calls such as Blender. Tool
// implementations remain outside this package.
type ToolInvoker interface {
	InvokeTool(ctx context.Context, request ToolRequest) (json.RawMessage, error)
}

type AgentInvokerFunc func(context.Context, AgentRequest) (json.RawMessage, error)

func (f AgentInvokerFunc) InvokeAgent(ctx context.Context, request AgentRequest) (json.RawMessage, error) {
	return f(ctx, request)
}

type StepInvokerFunc func(context.Context, StepRequest) (json.RawMessage, error)

func (f StepInvokerFunc) InvokeStep(ctx context.Context, request StepRequest) (json.RawMessage, error) {
	return f(ctx, request)
}

type ToolInvokerFunc func(context.Context, ToolRequest) (json.RawMessage, error)

func (f ToolInvokerFunc) InvokeTool(ctx context.Context, request ToolRequest) (json.RawMessage, error) {
	return f(ctx, request)
}

func cloneNodeSpec(spec NodeSpec) NodeSpec {
	spec.DependsOn = append([]NodeID(nil), spec.DependsOn...)
	spec.Input = append(json.RawMessage(nil), spec.Input...)
	return spec
}

func cloneSnapshot(snapshot Snapshot) Snapshot {
	snapshot.JSON = append(json.RawMessage(nil), snapshot.JSON...)
	return snapshot
}

func cloneNode(node Node) Node {
	node.DependsOn = append([]NodeID(nil), node.DependsOn...)
	node.Input = cloneSnapshot(node.Input)
	if node.Output != nil {
		output := cloneSnapshot(*node.Output)
		node.Output = &output
	}
	if node.StartedAt != nil {
		value := *node.StartedAt
		node.StartedAt = &value
	}
	if node.FinishedAt != nil {
		value := *node.FinishedAt
		node.FinishedAt = &value
	}
	return node
}

func cloneTask(task Task) Task {
	task.Input = cloneSnapshot(task.Input)
	nodes := task.Nodes
	task.Nodes = make([]Node, len(nodes))
	for index, node := range nodes {
		task.Nodes[index] = cloneNode(node)
	}
	if task.FinishedAt != nil {
		value := *task.FinishedAt
		task.FinishedAt = &value
	}
	return task
}
