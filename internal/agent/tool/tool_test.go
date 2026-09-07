package tool

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestRegistryResolvesEnabledToolProvider(t *testing.T) {
	repository := NewMemoryRepository()
	if err := repository.Create(context.Background(), Definition{
		ID: "order.query", Name: "Order query", Kind: KindRead,
	}); err != nil {
		t.Fatalf("create tool definition: %v", err)
	}
	registry, err := NewRegistry(repository)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	if err := registry.Register("order.query", ProviderFunc(func(context.Context) (tool.BaseTool, error) {
		return echoTool{}, nil
	})); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	resolved, err := registry.Resolve(context.Background(), "order.query")
	if err != nil {
		t.Fatalf("resolve tool: %v", err)
	}
	if resolved.Definition.ID != "order.query" || resolved.Tool == nil {
		t.Fatalf("resolved tool = %#v", resolved)
	}
}

func TestRegistryRejectsDisabledTool(t *testing.T) {
	repository := NewMemoryRepository()
	if err := repository.Create(context.Background(), Definition{
		ID: "refund.create", Name: "Refund", Kind: KindWrite, Status: StatusDisabled,
	}); err != nil {
		t.Fatalf("create tool definition: %v", err)
	}
	registry, err := NewRegistry(repository)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	if err := registry.Register("refund.create", ProviderFunc(func(context.Context) (tool.BaseTool, error) {
		return echoTool{}, nil
	})); err != nil {
		t.Fatalf("register provider: %v", err)
	}

	_, err = registry.Resolve(context.Background(), "refund.create")
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("resolve disabled tool error = %v, want ErrDisabled", err)
	}
}

type echoTool struct{}

func (echoTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "echo", Desc: "Echo input."}, nil
}

func (echoTool) InvokableRun(_ context.Context, arguments string, _ ...tool.Option) (string, error) {
	return arguments, nil
}
