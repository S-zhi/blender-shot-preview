package pipeline

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type execution struct {
	cancel context.CancelFunc
}

// Runner turns submitted workflows into asynchronous executions. Each batch of
// ready nodes runs concurrently; the repository remains the source of truth so
// a new Runner can recover an interrupted process.
type Runner struct {
	repository Repository
	agents     AgentInvoker
	steps      StepInvoker
	tools      ToolInvoker
	now        func() time.Time
	newTaskID  func() (string, error)

	mu           sync.Mutex
	active       map[string]execution
	taskLocks    map[string]*sync.Mutex
	subscribers  map[string][]chan PipelineEvent
	confirmChans map[string]chan struct{}
}

func NewRunner(repository Repository, agents AgentInvoker, steps StepInvoker, tools ToolInvoker) *Runner {
	return &Runner{
		repository:   repository,
		agents:       agents,
		steps:        steps,
		tools:        tools,
		now:          time.Now,
		newTaskID:    newTaskID,
		active:       make(map[string]execution),
		taskLocks:    make(map[string]*sync.Mutex),
		subscribers:  make(map[string][]chan PipelineEvent),
		confirmChans: make(map[string]chan struct{}),
	}
}

// Submit atomically deduplicates by idempotency key and starts the task in the
// background. The supplied context covers persistence only; execution is
// intentionally detached so an RPC response cancellation cannot stop a task.
func (r *Runner) Submit(ctx context.Context, submission Submission) (Task, bool, error) {
	if r.repository == nil {
		return Task{}, false, fmt.Errorf("%w: repository is required", ErrInvalidSubmission)
	}
	submission.IdempotencyKey = strings.TrimSpace(submission.IdempotencyKey)
	if submission.IdempotencyKey == "" {
		return Task{}, false, fmt.Errorf("%w: idempotency key is required", ErrInvalidSubmission)
	}
	if err := submission.Workflow.Validate(); err != nil {
		return Task{}, false, err
	}
	input, err := newSnapshot(submission.Input.JSON, submission.Input.CapturedAt)
	if err != nil {
		return Task{}, false, err
	}
	taskID, err := r.newTaskID()
	if err != nil {
		return Task{}, false, fmt.Errorf("create pipeline task ID: %w", err)
	}
	now := r.now().UTC()
	task := Task{
		ID: taskID, IdempotencyKey: submission.IdempotencyKey, Input: input,
		Status: TaskStatusPending, RequireConfirmation: submission.RequireConfirmation, CreatedAt: now, UpdatedAt: now,
		Nodes: make([]Node, 0, len(submission.Workflow.Nodes)),
	}
	for _, spec := range submission.Workflow.Nodes {
		nodeInput, err := newSnapshot(spec.Input, now)
		if err != nil {
			return Task{}, false, fmt.Errorf("node %q input: %w", spec.ID, err)
		}
		maxAttempts := spec.MaxAttempts
		if maxAttempts == 0 {
			maxAttempts = 1
		}
		task.Nodes = append(task.Nodes, Node{
			ID: spec.ID, DependsOn: append([]NodeID(nil), spec.DependsOn...), Invocation: spec.Invocation,
			Input: nodeInput, Status: NodeStatusPending, MaxAttempts: maxAttempts,
		})
	}
	stored, created, err := r.repository.CreateIfAbsent(ctx, task)
	if err != nil {
		return Task{}, false, err
	}
	r.broadcast(PipelineEvent{
		TaskID:    stored.ID,
		Type:      EventTaskStarted,
		Status:    string(stored.Status),
		Timestamp: now,
	})
	if !stored.Status.Terminal() {
		r.start(stored.ID)
	}
	return stored, created, nil
}

func (r *Runner) Get(ctx context.Context, taskID string) (Task, error) {
	if r.repository == nil {
		return Task{}, ErrTaskNotFound
	}
	return r.repository.Get(ctx, strings.TrimSpace(taskID))
}

// Cancel is idempotent after a task has reached cancelled state. It persists
// the cancellation before signalling active invocations, which makes a late
// completion unable to overwrite the terminal result.
func (r *Runner) Cancel(ctx context.Context, taskID string) error {
	if r.repository == nil {
		return ErrTaskNotFound
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ErrTaskNotFound
	}
	lock := r.taskLock(taskID)
	lock.Lock()
	task, err := r.repository.Get(ctx, taskID)
	if err == nil {
		if task.Status == TaskStatusCancelled {
			lock.Unlock()
			return nil
		}
		if task.Status.Terminal() {
			lock.Unlock()
			return ErrTaskAlreadyTerminal
		}
		now := r.now().UTC()
		task.CancelRequested = true
		task.Status = TaskStatusCancelled
		task.UpdatedAt = now
		task.FinishedAt = timePointer(now)
		for index := range task.Nodes {
			if !task.Nodes[index].Status.Terminal() {
				task.Nodes[index].Status = NodeStatusCancelled
				task.Nodes[index].FinishedAt = timePointer(now)
			}
		}
		err = r.repository.Update(ctx, task)
	}
	lock.Unlock()
	if err != nil {
		return err
	}
	r.mu.Lock()
	active, exists := r.active[taskID]
	for key, ch := range r.confirmChans {
		if strings.HasPrefix(key, taskID+":") {
			select {
			case ch <- struct{}{}:
			default:
			}
			delete(r.confirmChans, key)
		}
	}
	r.mu.Unlock()
	if exists {
		active.cancel()
	}
	return nil
}

// Retry restarts the failed portion of a task while retaining successful node
// outputs. It is intended for an explicit operator retry after automatic node
// attempts have been exhausted.
func (r *Runner) Retry(ctx context.Context, taskID string) (Task, error) {
	if r.repository == nil {
		return Task{}, ErrTaskNotFound
	}
	taskID = strings.TrimSpace(taskID)
	var retried Task
	err := r.withTask(taskID, func(task *Task) error {
		if task.Status != TaskStatusFailed {
			return ErrTaskNotFailed
		}
		now := r.now().UTC()
		task.Status = TaskStatusPending
		task.CancelRequested = false
		task.FinishedAt = nil
		task.UpdatedAt = now
		for index := range task.Nodes {
			node := &task.Nodes[index]
			if node.Status == NodeStatusSucceeded {
				continue
			}
			node.Status = NodeStatusPending
			node.Attempts = 0
			node.Error = ""
			node.ErrorCode = ""
			node.StartedAt = nil
			node.FinishedAt = nil
			node.Output = nil
		}
		retried = cloneTask(*task)
		return nil
	})
	if err != nil {
		return Task{}, err
	}
	r.start(taskID)
	return retried, nil
}

// Recover requeues persisted running nodes as pending, then schedules every
// unfinished task. Call it during process startup after constructing a Runner.
func (r *Runner) Recover(ctx context.Context) error {
	if r.repository == nil {
		return fmt.Errorf("%w: repository is required", ErrInvalidSubmission)
	}
	tasks, err := r.repository.ListIncomplete(ctx)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if r.isActive(task.ID) {
			continue
		}
		lock := r.taskLock(task.ID)
		lock.Lock()
		current, getErr := r.repository.Get(ctx, task.ID)
		if getErr == nil && !current.Status.Terminal() {
			changed := false
			for index := range current.Nodes {
				if current.Nodes[index].Status == NodeStatusRunning {
					if current.Nodes[index].Attempts >= current.Nodes[index].MaxAttempts {
						now := r.now().UTC()
						current.Nodes[index].Status = NodeStatusFailed
						current.Nodes[index].Error = "interrupted during final allowed attempt"
						current.Nodes[index].FinishedAt = timePointer(now)
					} else {
						current.Nodes[index].Status = NodeStatusPending
						current.Nodes[index].StartedAt = nil
						current.Nodes[index].FinishedAt = nil
					}
					changed = true
				}
			}
			if current.Status == TaskStatusRunning {
				current.Status = TaskStatusPending
				changed = true
			}
			if changed {
				current.UpdatedAt = r.now().UTC()
				getErr = r.repository.Update(ctx, current)
			}
		}
		lock.Unlock()
		if getErr != nil {
			return getErr
		}
		r.start(task.ID)
	}
	return nil
}

func (r *Runner) start(taskID string) {
	r.mu.Lock()
	if _, exists := r.active[taskID]; exists {
		r.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.active[taskID] = execution{cancel: cancel}
	r.mu.Unlock()
	go func() {
		recheck := false
		defer func() {
			cancel()
			r.mu.Lock()
			delete(r.active, taskID)
			r.mu.Unlock()
			if !recheck {
				return
			}
			// A manual retry can change a failed task back to pending while this
			// worker is between its final status read and active-map cleanup.
			// Re-check after releasing the active slot so that retry is never lost.
			if task, err := r.Get(context.Background(), taskID); err == nil && !task.Status.Terminal() {
				r.start(taskID)
			}
		}()
		recheck = r.run(ctx, taskID)
	}()
}

// run requests one final state check only when it observed a terminal task.
// Repository errors stop the worker without a hot restart loop; Recover can
// retry the persisted in-flight state once storage is healthy again.
func (r *Runner) run(ctx context.Context, taskID string) bool {
	for {
		task, err := r.Get(context.Background(), taskID)
		if err != nil {
			return false
		}
		if task.Status.Terminal() {
			return true
		}
		if task.CancelRequested || ctx.Err() != nil {
			return false
		}
		if task.Status == TaskStatusPending {
			if err := r.setTaskRunning(taskID); err != nil {
				return false
			}
			continue
		}
		ready := readyNodes(task)
		if len(ready) == 0 {
			if err := r.completeIfBlocked(taskID); err != nil {
				return false
			}
			continue
		}
		started, err := r.startNodes(taskID, ready)
		if err != nil || len(started) == 0 {
			if err != nil {
				return false
			}
			continue
		}
		var wait sync.WaitGroup
		persistErrors := make(chan error, len(started))
		for _, node := range started {
			node := node
			wait.Add(1)
			go func() {
				defer wait.Done()
				if err := r.executeNode(ctx, taskID, node); err != nil {
					persistErrors <- err
				}
			}()
		}
		wait.Wait()
		close(persistErrors)
		if len(persistErrors) > 0 {
			return false
		}
	}
}

func (r *Runner) setTaskRunning(taskID string) error {
	return r.withTask(taskID, func(task *Task) error {
		if task.Status == TaskStatusPending && !task.CancelRequested {
			task.Status = TaskStatusRunning
			task.UpdatedAt = r.now().UTC()
		}
		return nil
	})
}

func (r *Runner) startNodes(taskID string, ready []NodeID) ([]Node, error) {
	started := make([]Node, 0, len(ready))
	err := r.withTask(taskID, func(task *Task) error {
		if task.CancelRequested || task.Status.Terminal() {
			return nil
		}
		now := r.now().UTC()
		for _, id := range ready {
			for index := range task.Nodes {
				node := &task.Nodes[index]
				if node.ID != id || node.Status != NodeStatusPending || !dependenciesSucceeded(*node, *task) {
					continue
				}
				node.Status = NodeStatusRunning
				node.Attempts++
				node.StartedAt = timePointer(now)
				node.FinishedAt = nil
				node.Error = ""
				started = append(started, cloneNode(*node))
				break
			}
		}
		if len(started) > 0 {
			task.UpdatedAt = now
		}
		return nil
	})
	if err == nil {
		for _, n := range started {
			r.broadcast(PipelineEvent{
				TaskID:    taskID,
				Type:      EventNodeStarted,
				NodeID:    n.ID,
				Status:    string(n.Status),
				Input:     string(n.Input.JSON),
				Timestamp: r.now().UTC(),
			})
		}
	}
	return started, err
}

func (r *Runner) executeNode(ctx context.Context, taskID string, node Node) error {
	task, err := r.Get(context.Background(), taskID)
	if err != nil {
		return err
	}
	output, invokeErr := r.invoke(ctx, task, node)
	if invokeErr != nil {
		return r.finishNode(taskID, node.ID, output, invokeErr)
	}

	if task.RequireConfirmation && nodeRequiresConfirmation(node) {
		if err := r.setNodeWaitingConfirmation(taskID, node.ID, output); err != nil {
			return err
		}
		confirmChan := r.getConfirmChan(taskID, node.ID)
		select {
		case <-confirmChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return r.finishNode(taskID, node.ID, output, nil)
}

func nodeRequiresConfirmation(node Node) bool {
	return node.Invocation.Kind == InvocationAgent || node.ID == NodeInspect
}

func (r *Runner) setNodeWaitingConfirmation(taskID string, nodeID NodeID, output Snapshot) error {
	var inputJSON string
	err := r.withTask(taskID, func(task *Task) error {
		for i := range task.Nodes {
			node := &task.Nodes[i]
			if node.ID == nodeID {
				now := r.now().UTC()
				node.Status = NodeStatusWaitingConfirmation
				copyOut := cloneSnapshot(output)
				node.Output = &copyOut
				inputJSON = string(node.Input.JSON)
				task.UpdatedAt = now
				return nil
			}
		}
		return errors.New("node not found")
	})
	if err != nil {
		return err
	}
	r.broadcast(PipelineEvent{
		TaskID:    taskID,
		Type:      EventNodeWaitingConfirm,
		NodeID:    nodeID,
		Status:    string(NodeStatusWaitingConfirmation),
		Input:     inputJSON,
		Output:    string(output.JSON),
		Timestamp: r.now().UTC(),
	})
	return nil
}

func (r *Runner) ConfirmNode(ctx context.Context, taskID string, nodeID NodeID, adjustedOutput *Snapshot) error {
	var confirmedOutput Snapshot
	err := r.withTask(taskID, func(task *Task) error {
		if task.Status.Terminal() || task.CancelRequested {
			return ErrTaskAlreadyTerminal
		}
		for i := range task.Nodes {
			node := &task.Nodes[i]
			if node.ID == nodeID {
				if node.Status != NodeStatusWaitingConfirmation {
					return errors.New("node is not waiting for confirmation")
				}
				now := r.now().UTC()
				node.FinishedAt = timePointer(now)
				node.Status = NodeStatusSucceeded
				if adjustedOutput != nil && len(adjustedOutput.JSON) > 0 {
					copyOut := cloneSnapshot(*adjustedOutput)
					node.Output = &copyOut
				}
				if node.Output != nil {
					confirmedOutput = cloneSnapshot(*node.Output)
				}
				task.UpdatedAt = now
				return nil
			}
		}
		return errors.New("node not found")
	})
	if err != nil {
		return err
	}

	r.signalConfirm(taskID, nodeID)

	r.broadcast(PipelineEvent{
		TaskID:    taskID,
		Type:      EventNodeSucceeded,
		NodeID:    nodeID,
		Status:    string(NodeStatusSucceeded),
		Output:    string(confirmedOutput.JSON),
		Timestamp: r.now().UTC(),
	})
	return nil
}

func (r *Runner) AdjustNodeOutput(ctx context.Context, taskID string, nodeID NodeID, newOutput Snapshot) error {
	err := r.withTask(taskID, func(task *Task) error {
		if task.Status.Terminal() || task.CancelRequested {
			return ErrTaskAlreadyTerminal
		}
		for i := range task.Nodes {
			node := &task.Nodes[i]
			if node.ID == nodeID {
				copyOut := cloneSnapshot(newOutput)
				node.Output = &copyOut
				task.UpdatedAt = r.now().UTC()
				return nil
			}
		}
		return errors.New("node not found")
	})
	if err != nil {
		return err
	}
	r.broadcast(PipelineEvent{
		TaskID:    taskID,
		Type:      EventNodeOutputAdjusted,
		NodeID:    nodeID,
		Output:    string(newOutput.JSON),
		Timestamp: r.now().UTC(),
	})
	return nil
}

func (r *Runner) getConfirmChan(taskID string, nodeID NodeID) chan struct{} {
	key := taskID + ":" + string(nodeID)
	r.mu.Lock()
	defer r.mu.Unlock()
	ch, ok := r.confirmChans[key]
	if !ok {
		ch = make(chan struct{}, 1)
		r.confirmChans[key] = ch
	}
	return ch
}

func (r *Runner) signalConfirm(taskID string, nodeID NodeID) {
	key := taskID + ":" + string(nodeID)
	r.mu.Lock()
	defer r.mu.Unlock()
	if ch, ok := r.confirmChans[key]; ok {
		select {
		case ch <- struct{}{}:
		default:
		}
		delete(r.confirmChans, key)
	}
}

func (r *Runner) Subscribe(taskID string) (<-chan PipelineEvent, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch := make(chan PipelineEvent, 64)
	r.subscribers[taskID] = append(r.subscribers[taskID], ch)
	unsubscribe := func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		subs := r.subscribers[taskID]
		for i, sub := range subs {
			if sub == ch {
				r.subscribers[taskID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
		if len(r.subscribers[taskID]) == 0 {
			delete(r.subscribers, taskID)
		}
	}
	return ch, unsubscribe
}

func (r *Runner) broadcast(event PipelineEvent) {
	if events, ok := r.repository.(EventRepository); ok {
		_ = events.AppendEvent(context.Background(), event)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if subs, ok := r.subscribers[event.TaskID]; ok {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (r *Runner) invoke(ctx context.Context, task Task, node Node) (Snapshot, error) {
	dependencies := completedDependencies(node, task)
	var (
		output json.RawMessage
		err    error
	)
	switch node.Invocation.Kind {
	case InvocationAgent:
		if r.agents == nil {
			return Snapshot{}, fmt.Errorf("%w: agent %q", ErrMissingInvoker, node.Invocation.Target)
		}
		output, err = r.agents.InvokeAgent(ctx, AgentRequest{TaskID: task.ID, NodeID: node.ID, AgentID: node.Invocation.Target, Attempt: node.Attempts, TaskInput: cloneSnapshot(task.Input), Input: cloneSnapshot(node.Input), Dependencies: dependencies})
	case InvocationStep:
		if r.steps == nil {
			return Snapshot{}, fmt.Errorf("%w: step %q", ErrMissingInvoker, node.Invocation.Target)
		}
		output, err = r.steps.InvokeStep(ctx, StepRequest{TaskID: task.ID, NodeID: node.ID, Step: node.Invocation.Target, Attempt: node.Attempts, TaskInput: cloneSnapshot(task.Input), Input: cloneSnapshot(node.Input), Dependencies: dependencies})
	case InvocationTool:
		if r.tools == nil {
			return Snapshot{}, fmt.Errorf("%w: tool %q", ErrMissingInvoker, node.Invocation.Target)
		}
		output, err = r.tools.InvokeTool(ctx, ToolRequest{TaskID: task.ID, NodeID: node.ID, Tool: node.Invocation.Target, Attempt: node.Attempts, TaskInput: cloneSnapshot(task.Input), Input: cloneSnapshot(node.Input), Dependencies: dependencies})
	default:
		return Snapshot{}, fmt.Errorf("%w: invocation kind %q", ErrInvalidWorkflow, node.Invocation.Kind)
	}
	if err != nil {
		return Snapshot{}, err
	}
	snapshot, err := newSnapshot(output, r.now().UTC())
	if err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (r *Runner) finishNode(taskID string, nodeID NodeID, output Snapshot, invokeErr error) error {
	var (
		finalNodeStatus NodeStatus
		nodeError       string
		nodeErrorCode   string
		outSnapshot     Snapshot
	)
	err := r.withTask(taskID, func(task *Task) error {
		if task.CancelRequested || task.Status.Terminal() {
			return nil
		}
		for index := range task.Nodes {
			node := &task.Nodes[index]
			if node.ID != nodeID || (node.Status != NodeStatusRunning && node.Status != NodeStatusWaitingConfirmation) {
				continue
			}
			now := r.now().UTC()
			node.FinishedAt = timePointer(now)
			if invokeErr == nil {
				copyOfOutput := cloneSnapshot(output)
				node.Output = &copyOfOutput
				node.Status = NodeStatusSucceeded
				node.Error = ""
				node.ErrorCode = ""
				finalNodeStatus = NodeStatusSucceeded
				outSnapshot = cloneSnapshot(output)
			} else {
				node.Error = strings.TrimSpace(invokeErr.Error())
				nodeError = node.Error
				if coded, ok := invokeErr.(interface{ ErrorCode() string }); ok {
					nodeErrorCode = coded.ErrorCode()
				}
				node.ErrorCode = nodeErrorCode
				if node.Attempts < node.MaxAttempts {
					node.Status = NodeStatusPending
					node.FinishedAt = nil
					finalNodeStatus = NodeStatusPending
				} else {
					node.Status = NodeStatusFailed
					finalNodeStatus = NodeStatusFailed
				}
			}
			task.UpdatedAt = now
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	if finalNodeStatus == NodeStatusSucceeded {
		r.broadcast(PipelineEvent{
			TaskID:    taskID,
			Type:      EventNodeSucceeded,
			NodeID:    nodeID,
			Status:    string(NodeStatusSucceeded),
			Output:    string(outSnapshot.JSON),
			Timestamp: r.now().UTC(),
		})
	} else if finalNodeStatus == NodeStatusFailed {
		r.broadcast(PipelineEvent{
			TaskID:    taskID,
			Type:      EventNodeFailed,
			NodeID:    nodeID,
			Status:    string(NodeStatusFailed),
			Error:     nodeError,
			ErrorCode: nodeErrorCode,
			Timestamp: r.now().UTC(),
		})
	}
	return nil
}

func (r *Runner) completeIfBlocked(taskID string) error {
	var (
		finalStatus TaskStatus
		failedError string
		failedCode  string
	)
	err := r.withTask(taskID, func(task *Task) error {
		if task.Status.Terminal() || task.CancelRequested {
			return nil
		}
		for _, node := range task.Nodes {
			if node.Status == NodeStatusRunning || node.Status == NodeStatusWaitingConfirmation {
				return nil
			}
		}
		now := r.now().UTC()
		for _, node := range task.Nodes {
			if node.Status == NodeStatusFailed {
				task.Status = TaskStatusFailed
				task.FinishedAt = timePointer(now)
				finalStatus = TaskStatusFailed
				failedError = node.Error
				failedCode = node.ErrorCode
				for index := range task.Nodes {
					if task.Nodes[index].Status == NodeStatusPending {
						task.Nodes[index].Status = NodeStatusCancelled
						task.Nodes[index].FinishedAt = timePointer(now)
					}
				}
				task.UpdatedAt = now
				return nil
			}
		}
		allSucceeded := true
		for _, node := range task.Nodes {
			if node.Status != NodeStatusSucceeded {
				allSucceeded = false
				break
			}
		}
		if allSucceeded {
			task.Status = TaskStatusSucceeded
			task.FinishedAt = timePointer(now)
			task.UpdatedAt = now
			finalStatus = TaskStatusSucceeded
		}
		return nil
	})
	if err != nil {
		return err
	}
	if finalStatus == TaskStatusSucceeded {
		r.broadcast(PipelineEvent{
			TaskID: taskID,
			Type:   EventTaskSucceeded,
			Status: string(TaskStatusSucceeded),
			Artifacts: []ArtifactSnapshot{
				{
					Type: "video",
					Name: "shot-preview.mp4",
					URI:  "/api/v0_1/shot-preview/artifacts/download?task_id=" + taskID + "&name=shot-preview.mp4",
				},
			},
			Timestamp: r.now().UTC(),
		})
	} else if finalStatus == TaskStatusFailed {
		r.broadcast(PipelineEvent{
			TaskID:    taskID,
			Type:      EventTaskFailed,
			Status:    string(TaskStatusFailed),
			Error:     failedError,
			ErrorCode: failedCode,
			Timestamp: r.now().UTC(),
		})
	}
	return nil
}

func (r *Runner) withTask(taskID string, operation func(*Task) error) error {
	lock := r.taskLock(taskID)
	lock.Lock()
	defer lock.Unlock()
	task, err := r.repository.Get(context.Background(), taskID)
	if err != nil {
		return err
	}
	if err := operation(&task); err != nil {
		return err
	}
	return r.repository.Update(context.Background(), task)
}

func (r *Runner) taskLock(taskID string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	lock, exists := r.taskLocks[taskID]
	if !exists {
		lock = &sync.Mutex{}
		r.taskLocks[taskID] = lock
	}
	return lock
}

func (r *Runner) isActive(taskID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.active[taskID]
	return exists
}

func readyNodes(task Task) []NodeID {
	ready := make([]NodeID, 0, len(task.Nodes))
	for _, node := range task.Nodes {
		if node.Status == NodeStatusPending && dependenciesSucceeded(node, task) {
			ready = append(ready, node.ID)
		}
	}
	return ready
}

func dependenciesSucceeded(node Node, task Task) bool {
	for _, dependency := range node.DependsOn {
		matched, exists := task.Node(dependency)
		if !exists || matched.Status != NodeStatusSucceeded {
			return false
		}
	}
	return true
}

func completedDependencies(node Node, task Task) DependencyOutputs {
	outputs := make(DependencyOutputs, len(node.DependsOn))
	for _, dependency := range node.DependsOn {
		if matched, exists := task.Node(dependency); exists && matched.Output != nil {
			outputs[dependency] = cloneSnapshot(*matched.Output)
		}
	}
	return outputs
}

func timePointer(value time.Time) *time.Time { return &value }

func newTaskID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
