package agent

import (
	_ "embed"
	"encoding/json"
	"strings"
	"time"
)

const (
	// ProductionModelProfile is registered by the application bootstrap.
	ProductionModelProfile = "scene-production"

	IntentAgentID        = "intent-agent"
	ScenePlannerAgentID  = "scene-planner-agent"
	AssetCreatorAgentID  = "asset-creator-agent"
	ShotDesignerAgentID  = "shot-designer-agent"
	SceneAssemblyAgentID = "scene-assembly-agent"

	ToolIDAssetSearch            = "asset.search"
	ToolIDAssetInspect           = "asset.inspect"
	ToolIDBlenderCreateModel     = "blender.create_model"
	ToolIDBlenderCreateMaterial  = "blender.create_material"
	ToolIDBlenderCreateRig       = "blender.create_rig"
	ToolIDBlenderImportReference = "blender.import_reference"
	ToolIDBlenderInspectAsset    = "blender.inspect_asset"

	ToolIDBlenderCreateProject   = "blender.create_project"
	ToolIDBlenderImportAsset     = "blender.import_asset"
	ToolIDBlenderApplyScenePatch = "blender.apply_scene_patch"
	ToolIDBlenderApplyShotPlan   = "blender.apply_shot_plan"
	ToolIDBlenderSaveProject     = "blender.save_project"
	ToolIDBlenderInspectScene    = "blender.inspect_scene"
)

// Versioned contracts are embedded so deployed catalog definitions and their
// contract revision cannot drift apart.
//
//go:embed catalog/contracts/v1/intent-request.schema.json
var intentRequestSchema []byte

//go:embed catalog/contracts/v1/scene-spec.schema.json
var sceneSpecSchema []byte

//go:embed catalog/contracts/v1/scene-plan.schema.json
var scenePlanSchema []byte

//go:embed catalog/contracts/v1/asset-task.schema.json
var assetTaskSchema []byte

//go:embed catalog/contracts/v1/asset-manifest.schema.json
var assetManifestSchema []byte

//go:embed catalog/contracts/v1/shot-design-request.schema.json
var shotDesignRequestSchema []byte

//go:embed catalog/contracts/v1/shot-plan.schema.json
var shotPlanSchema []byte

//go:embed catalog/contracts/v1/scene-assembly-request.schema.json
var sceneAssemblyRequestSchema []byte

//go:embed catalog/contracts/v1/scene-assembly-result.schema.json
var sceneAssemblyResultSchema []byte

// Skill instruction assets are distinct from executable Go tools. SkillIDs
// below is only the static tool allowlist used by the runtime registry.
//
//go:embed catalog/skills/intent-agent.md
var intentAgentSkill string

//go:embed catalog/skills/scene-planner-agent.md
var scenePlannerAgentSkill string

//go:embed catalog/skills/asset-creator-agent.md
var assetCreatorAgentSkill string

//go:embed catalog/skills/shot-designer-agent.md
var shotDesignerAgentSkill string

//go:embed catalog/skills/scene-assembly-agent.md
var sceneAssemblyAgentSkill string

// ProductionAgentDefinitions returns the complete static scene-production
// catalog. Every invocation produces independent definitions and schema bytes.
// Callers must register ProductionModelProfile and the named executable tool
// IDs before creating these definitions with DefinitionService.
func ProductionAgentDefinitions() []AgentDefinition {
	return []AgentDefinition{
		{
			ID: IntentAgentID, Name: "Intent Agent",
			Description:  "Converts an initial creative request into a versioned SceneSpec.",
			SystemPrompt: skillInstruction(intentAgentSkill), ModelProfile: ProductionModelProfile,
			InputSchema: cloneSchema(intentRequestSchema), OutputSchema: cloneSchema(sceneSpecSchema),
			MaxSteps: 6, MaxOutputTokens: 2048, Timeout: 40 * time.Minute, Status: AgentStatusEnabled,
		},
		{
			ID: ScenePlannerAgentID, Name: "Scene Planner Agent",
			Description:  "Builds a dependency-aware asset, environment, and action plan from a SceneSpec.",
			SystemPrompt: skillInstruction(scenePlannerAgentSkill), ModelProfile: ProductionModelProfile,
			InputSchema: cloneSchema(sceneSpecSchema), OutputSchema: cloneSchema(scenePlanSchema),
			MaxSteps: 10, MaxOutputTokens: 3072, Timeout: 40 * time.Minute, Status: AgentStatusEnabled,
		},
		{
			ID: AssetCreatorAgentID, Name: "Asset Creator Agent",
			Description:  "Creates and inspects one isolated Blender asset task, then returns an AssetManifest.",
			SystemPrompt: skillInstruction(assetCreatorAgentSkill), ModelProfile: ProductionModelProfile,
			SkillIDs: []string{
				ToolIDAssetSearch, ToolIDBlenderImportReference, ToolIDBlenderCreateModel,
				ToolIDBlenderCreateMaterial, ToolIDBlenderCreateRig, ToolIDBlenderInspectAsset,
			},
			InputSchema: cloneSchema(assetTaskSchema), OutputSchema: cloneSchema(assetManifestSchema),
			MaxSteps: 28, MaxOutputTokens: 4096, Timeout: 40 * time.Minute, Status: AgentStatusEnabled,
		},
		{
			ID: ShotDesignerAgentID, Name: "Shot Designer Agent",
			Description:  "Combines a SceneSpec and AssetManifests into an executable ShotPlan.",
			SystemPrompt: skillInstruction(shotDesignerAgentSkill), ModelProfile: ProductionModelProfile,
			InputSchema: cloneSchema(shotDesignRequestSchema), OutputSchema: cloneSchema(shotPlanSchema),
			MaxSteps: 10, MaxOutputTokens: 3072, Timeout: 40 * time.Minute, Status: AgentStatusEnabled,
		},
		{
			ID: SceneAssemblyAgentID, Name: "Scene Assembly Agent",
			Description:  "Assembles manifests and a ShotPlan into an editable Blender project.",
			SystemPrompt: skillInstruction(sceneAssemblyAgentSkill), ModelProfile: ProductionModelProfile,
			SkillIDs: []string{
				ToolIDBlenderCreateProject, ToolIDBlenderImportAsset, ToolIDBlenderApplyScenePatch,
				ToolIDBlenderApplyShotPlan, ToolIDBlenderSaveProject, ToolIDBlenderInspectScene,
			},
			InputSchema: cloneSchema(sceneAssemblyRequestSchema), OutputSchema: cloneSchema(sceneAssemblyResultSchema),
			MaxSteps: 24, MaxOutputTokens: 4096, Timeout: 40 * time.Minute, Status: AgentStatusEnabled,
		},
	}
}

// ProductionAgentToolAllowlist returns a catalog agent's fixed executable-tool
// allowlist. The returned slice is independent of the catalog.
func ProductionAgentToolAllowlist(agentID string) ([]string, bool) {
	for _, definition := range ProductionAgentDefinitions() {
		if definition.ID == agentID {
			return append([]string(nil), definition.SkillIDs...), true
		}
	}
	return nil, false
}

func cloneSchema(schema []byte) json.RawMessage {
	return append(json.RawMessage(nil), schema...)
}

func skillInstruction(instruction string) string {
	return strings.TrimSpace(instruction)
}
