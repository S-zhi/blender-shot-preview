package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	agenttool "github.com/S-zhi/blender-shot-preview/internal/agent/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	maximumCallsPerTool      = 4
	maximumConsecutiveErrors = 3
	defaultToolTimeout       = 30 * time.Second
	maximumReadAttempts      = 3
)

type ToolKind = agenttool.Kind

const (
	ToolKindRead         = agenttool.KindRead
	ToolKindWrite        = agenttool.KindWrite
	ToolKindIrreversible = agenttool.KindIrreversible
)

// ToolExecution contains the stable identifiers that a Skill passes to its
// business Service. A write Service must use IdempotencyKey when persisting a
// side effect or forwarding a call to another service.
type ToolExecution struct {
	RunID          string
	ToolCallID     string
	IdempotencyKey string
	ToolName       string
	Attempt        int
}

type toolExecutionKey struct{}
type agentRunIDKey struct{}

func ToolExecutionFromContext(ctx context.Context) (ToolExecution, bool) {
	execution, ok := ctx.Value(toolExecutionKey{}).(ToolExecution)
	return execution, ok
}

func withRunID(ctx context.Context, runID string) context.Context {
	return context.WithValue(ctx, agentRunIDKey{}, runID)
}

func runIDFromContext(ctx context.Context) string {
	runID, _ := ctx.Value(agentRunIDKey{}).(string)
	return runID
}

// ToolError lets a Skill explicitly declare whether a read failure is safe to
// retry. Write and irreversible tools are never retried by this generic layer.
type ToolError struct {
	Code      string
	Retryable bool
	Cause     error
}

func (e *ToolError) Error() string {
	if e.Cause == nil {
		return e.Code
	}
	return fmt.Sprintf("%s: %v", e.Code, e.Cause)
}

func (e *ToolError) Unwrap() error { return e.Cause }

type toolExecutionGuard struct {
	mu                  sync.Mutex
	totalLimit          int
	totalCalls          int
	callsByTool         map[string]int
	consecutiveFailures int
	nextCall            int
}

func newToolExecutionGuard(totalLimit int) *toolExecutionGuard {
	return &toolExecutionGuard{
		totalLimit:  totalLimit,
		callsByTool: make(map[string]int),
	}
}

func (g *toolExecutionGuard) wrap(base tool.BaseTool, kind ToolKind) (tool.BaseTool, error) {
	invokable, ok := base.(tool.InvokableTool)
	if !ok {
		return nil, fmt.Errorf("%w: skill must implement InvokableTool", ErrInvalidDefinition)
	}
	return &guardedTool{base: invokable, kind: kind, guard: g}, nil
}

type guardedTool struct {
	base  tool.InvokableTool
	kind  ToolKind
	guard *toolExecutionGuard
}

func (t *guardedTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.base.Info(ctx)
}

func (t *guardedTool) InvokableRun(ctx context.Context, arguments string, options ...tool.Option) (string, error) {
	info, err := t.base.Info(ctx)
	if err != nil {
		return "", err
	}
	callID, err := t.guard.begin(info.Name)
	if err != nil {
		return "", err
	}
	execution := ToolExecution{
		RunID:          runIDFromContext(ctx),
		ToolCallID:     callID,
		IdempotencyKey: fmt.Sprintf("agent:%s:tool:%s", runIDFromContext(ctx), callID),
		ToolName:       info.Name,
	}
	for attempt := 1; attempt <= t.maxAttempts(); attempt++ {
		execution.Attempt = attempt
		callContext, cancel := context.WithTimeout(context.WithValue(ctx, toolExecutionKey{}, execution), defaultToolTimeout)
		result, callErr := t.base.InvokableRun(callContext, arguments, options...)
		cancel()
		if callErr == nil {
			t.guard.finish(true)
			return result, nil
		}
		if !t.shouldRetry(callErr, attempt) {
			t.guard.finish(false)
			return "", callErr
		}
	}
	t.guard.finish(false)
	return "", errors.New("unreachable tool retry state")
}

func (t *guardedTool) maxAttempts() int {
	if t.kind == ToolKindRead {
		return maximumReadAttempts
	}
	return 1
}

func (t *guardedTool) shouldRetry(err error, attempt int) bool {
	if t.kind != ToolKindRead || attempt >= maximumReadAttempts {
		return false
	}
	var toolErr *ToolError
	return errors.As(err, &toolErr) && toolErr.Retryable
}

func (g *toolExecutionGuard) begin(toolName string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.consecutiveFailures >= maximumConsecutiveErrors {
		return "", ErrConsecutiveFailures
	}
	if g.totalCalls >= g.totalLimit {
		return "", ErrToolCallLimit
	}
	if g.callsByTool[toolName] >= maximumCallsPerTool {
		return "", ErrRepeatedToolLimit
	}
	g.totalCalls++
	g.callsByTool[toolName]++
	g.nextCall++
	return fmt.Sprintf("%s-%d", toolName, g.nextCall), nil
}

func (g *toolExecutionGuard) finish(success bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if success {
		g.consecutiveFailures = 0
		return
	}
	g.consecutiveFailures++
}
