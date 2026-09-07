package skill

import (
	"context"
	"testing"

	agenttool "github.com/S-zhi/blender-shot-preview/internal/agent/tool"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestRegistryResolvesEveryToolBoundToSkill(t *testing.T) {
	toolsRepository := agenttool.NewMemoryRepository()
	for _, definition := range []agenttool.Definition{
		{ID: "order.query", Name: "Order query", Kind: agenttool.KindRead},
		{ID: "logistics.query", Name: "Logistics query", Kind: agenttool.KindRead},
	} {
		if err := toolsRepository.Create(context.Background(), definition); err != nil {
			t.Fatalf("create tool definition: %v", err)
		}
	}
	tools, err := agenttool.NewRegistry(toolsRepository)
	if err != nil {
		t.Fatalf("new tool registry: %v", err)
	}
	for _, toolID := range []string{"order.query", "logistics.query"} {
		if err := tools.Register(toolID, agenttool.ProviderFunc(func(context.Context) (einotool.BaseTool, error) {
			return testTool{}, nil
		})); err != nil {
			t.Fatalf("register tool provider: %v", err)
		}
	}

	skillsRepository := NewMemoryRepository()
	if err := skillsRepository.Create(context.Background(), Definition{
		ID: "order_lookup", Name: "Order lookup", ToolIDs: []string{"order.query", "logistics.query"},
	}); err != nil {
		t.Fatalf("create skill definition: %v", err)
	}
	registry, err := NewRegistry(skillsRepository, tools)
	if err != nil {
		t.Fatalf("new skill registry: %v", err)
	}

	resolved, err := registry.Resolve(context.Background(), "order_lookup")
	if err != nil {
		t.Fatalf("resolve skill: %v", err)
	}
	if len(resolved) != 2 || resolved[0].Definition.ID != "order.query" || resolved[1].Definition.ID != "logistics.query" {
		t.Fatalf("resolved tools = %#v", resolved)
	}
}

type testTool struct{}

func (testTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "test", Desc: "Test tool."}, nil
}

func (testTool) InvokableRun(_ context.Context, arguments string, _ ...einotool.Option) (string, error) {
	return arguments, nil
}
