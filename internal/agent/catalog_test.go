package agent

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestProductionAgentDefinitions(t *testing.T) {
	t.Parallel()

	definitions := ProductionAgentDefinitions()
	if len(definitions) != 5 {
		t.Fatalf("catalog definitions = %d, want 5", len(definitions))
	}

	type expectation struct {
		inputRequired  []string
		outputRequired []string
		maxSteps       int
		maxTokens      int
		timeout        time.Duration
		toolIDs        []string
	}
	expected := map[string]expectation{
		IntentAgentID: {
			inputRequired: []string{"prompt"}, outputRequired: []string{"version", "scene_id", "summary", "style", "environment", "assets", "actions", "shots"},
			maxSteps: 6, maxTokens: 2048, timeout: 40 * time.Minute,
		},
		ScenePlannerAgentID: {
			inputRequired: []string{"version", "scene_id", "summary", "style", "environment", "assets", "actions", "shots"}, outputRequired: []string{"version", "scene_id", "asset_tasks", "environment_plan", "action_plan"},
			maxSteps: 10, maxTokens: 3072, timeout: 40 * time.Minute,
		},
		AssetCreatorAgentID: {
			inputRequired: []string{"version", "task_id", "asset_id", "asset_kind", "description", "workspace", "output_path"}, outputRequired: []string{"version", "asset_id", "asset_kind", "source_files", "blend_file", "inspection"},
			maxSteps: 28, maxTokens: 4096, timeout: 40 * time.Minute,
			toolIDs: []string{
				ToolIDAssetSearch, ToolIDBlenderImportReference, ToolIDBlenderCreateModel,
				ToolIDBlenderCreateMaterial, ToolIDBlenderCreateRig, ToolIDBlenderInspectAsset,
			},
		},
		ShotDesignerAgentID: {
			inputRequired: []string{"scene_spec", "asset_manifests"}, outputRequired: []string{"version", "scene_id", "shots", "constraints"},
			maxSteps: 10, maxTokens: 3072, timeout: 40 * time.Minute,
		},
		SceneAssemblyAgentID: {
			inputRequired: []string{"asset_manifests", "shot_plan", "output_path"}, outputRequired: []string{"version", "scene_id", "blend_file", "inspection"},
			maxSteps: 24, maxTokens: 4096, timeout: 40 * time.Minute,
			toolIDs: []string{
				ToolIDBlenderCreateProject, ToolIDBlenderImportAsset, ToolIDBlenderApplyScenePatch,
				ToolIDBlenderApplyShotPlan, ToolIDBlenderSaveProject, ToolIDBlenderInspectScene,
			},
		},
	}

	seen := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		want, exists := expected[definition.ID]
		if !exists {
			t.Fatalf("unexpected catalog agent %q", definition.ID)
		}
		if _, duplicate := seen[definition.ID]; duplicate {
			t.Fatalf("duplicate catalog agent %q", definition.ID)
		}
		seen[definition.ID] = struct{}{}
		if definition.ModelProfile != ProductionModelProfile || definition.Status != AgentStatusEnabled {
			t.Fatalf("%s model/status = %q/%q", definition.ID, definition.ModelProfile, definition.Status)
		}
		if definition.SystemPrompt == "" {
			t.Fatalf("%s has an empty skill instruction", definition.ID)
		}
		if definition.MaxSteps != want.maxSteps || definition.MaxSteps < 1 || definition.MaxSteps > maximumMaxSteps {
			t.Fatalf("%s max steps = %d", definition.ID, definition.MaxSteps)
		}
		if definition.MaxOutputTokens != want.maxTokens || definition.MaxOutputTokens < 1 || definition.MaxOutputTokens > maximumMaxOutputTokens {
			t.Fatalf("%s max output tokens = %d", definition.ID, definition.MaxOutputTokens)
		}
		if definition.Timeout != want.timeout || definition.Timeout < time.Second || definition.Timeout > maximumTimeout {
			t.Fatalf("%s timeout = %s", definition.ID, definition.Timeout)
		}
		assertRequiredFields(t, definition.ID+" input", definition.InputSchema, want.inputRequired)
		assertRequiredFields(t, definition.ID+" output", definition.OutputSchema, want.outputRequired)
		if !reflect.DeepEqual(definition.SkillIDs, want.toolIDs) {
			t.Fatalf("%s tool allowlist = %#v, want %#v", definition.ID, definition.SkillIDs, want.toolIDs)
		}
		allowlist, ok := ProductionAgentToolAllowlist(definition.ID)
		if !ok || !reflect.DeepEqual(allowlist, want.toolIDs) {
			t.Fatalf("%s exported tool allowlist = %#v, found=%t", definition.ID, allowlist, ok)
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("catalog IDs = %#v, want %#v", seen, expected)
	}
}

func TestProductionAgentDefinitionsReturnIndependentValues(t *testing.T) {
	t.Parallel()

	first := ProductionAgentDefinitions()
	first[0].InputSchema[0] = 'x'
	first[2].SkillIDs[0] = "unexpected"
	second := ProductionAgentDefinitions()
	if second[0].InputSchema[0] == 'x' || second[2].SkillIDs[0] == "unexpected" {
		t.Fatal("catalog callers can mutate a later catalog result")
	}
	if _, ok := ProductionAgentToolAllowlist("unknown-agent"); ok {
		t.Fatal("unknown agent unexpectedly has a tool allowlist")
	}
}

func assertRequiredFields(t *testing.T, name string, raw json.RawMessage, want []string) {
	t.Helper()
	var contract struct {
		Type     string   `json:"type"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("decode %s contract: %v", name, err)
	}
	if contract.Type != "object" {
		t.Fatalf("%s root type = %q, want object", name, contract.Type)
	}
	if !reflect.DeepEqual(contract.Required, want) {
		t.Fatalf("%s required = %#v, want %#v", name, contract.Required, want)
	}
}
