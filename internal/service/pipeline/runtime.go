package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/agent"
	llmgateway "github.com/S-zhi/blender-shot-preview/internal/agent/llm_gateway"
	"github.com/S-zhi/blender-shot-preview/internal/productiontools"
	"github.com/cloudwego/eino/components/tool"
)

// AgentServiceInvoker adapts the application agent service to pipeline nodes
// and owns the domain-specific input projection between workflow snapshots.
type AgentServiceInvoker struct {
	Service    agent.AgentService
	Keys       llmgateway.KeyUser
	RequireLLM bool
}

// NewShotPreviewRunner assembles the stable shot-preview runtime boundary.
// Callers retain ownership of persistence, agent bootstrap, and tool setup.
func NewShotPreviewRunner(repository Repository, agents agent.AgentService, skills agent.SkillRegistry) *Runner {
	return newShotPreviewRunner(repository, agents, skills, nil, false)
}

// NewShotPreviewRunnerWithKeys allows configuring a KeyUser to resolve credentials for agents dynamically.
func NewShotPreviewRunnerWithKeys(repository Repository, agents agent.AgentService, skills agent.SkillRegistry, keys llmgateway.KeyUser) *Runner {
	return newShotPreviewRunner(repository, agents, skills, keys, true)
}

func newShotPreviewRunner(repository Repository, agents agent.AgentService, skills agent.SkillRegistry, keys llmgateway.KeyUser, requireLLM bool) *Runner {
	agentInvoker := AgentServiceInvoker{Service: agents, Keys: keys, RequireLLM: requireLLM}
	return NewRunner(
		repository,
		agentInvoker,
		ShotPreviewSteps{Assets: AssetFanOutStep{Agents: agentInvoker}},
		RegistryToolInvoker{Skills: skills},
	)
}

func (i AgentServiceInvoker) InvokeAgent(ctx context.Context, request AgentRequest) (json.RawMessage, error) {
	if i.Service == nil {
		return nil, fmt.Errorf("%w: agent service", ErrMissingInvoker)
	}
	input, identity, err := buildAgentInput(request)
	if err != nil {
		return nil, err
	}

	agentReq := agent.AgentRequest{
		AgentID: request.AgentID, SessionID: request.TaskID, Input: string(input),
		UserID: identity.UserID, TenantID: identity.UserID,
	}

	if i.RequireLLM {
		if i.Keys == nil {
			return nil, &agent.LLMError{Code: agent.LLMNotConfigured, Cause: errors.New("no LLM credential provider configured")}
		}
		if identity.UserID == "" {
			return nil, &agent.LLMError{Code: agent.LLMCredentialUnavailable, Cause: errors.New("user identity is required to resolve an LLM credential")}
		}
		usableKey, keyErr := i.Keys.Use(ctx, identity.KeyID, identity.UserID)
		if keyErr != nil {
			code := agent.LLMCredentialUnavailable
			if errors.Is(keyErr, llmgateway.ErrCredentialNotFound) || errors.Is(keyErr, llmgateway.ErrInvalidCommand) {
				code = agent.LLMNotConfigured
			}
			return nil, &agent.LLMError{Code: code, Cause: errors.New("no usable LLM credential is available")}
		}
		if strings.TrimSpace(usableKey.APIKey) == "" {
			return nil, &agent.LLMError{Code: agent.LLMNotConfigured, Cause: errors.New("resolved LLM credential has no API key")}
		}
		agentReq.Model = agent.NewChatModelFromKey(usableKey, "")
	} else if i.Keys != nil && identity.UserID != "" {
		if usableKey, err := i.Keys.Use(ctx, identity.KeyID, identity.UserID); err == nil && usableKey.APIKey != "" {
			agentReq.Model = agent.NewChatModelFromKey(usableKey, "")
		}
	}

	result, err := i.Service.Run(ctx, agentReq)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Status != agent.RunStatusSucceeded || !json.Valid([]byte(result.Output)) {
		return nil, fmt.Errorf("agent %q returned an invalid result", request.AgentID)
	}
	return json.RawMessage(result.Output), nil
}

type pipelineIdentity struct {
	UserID         string `json:"user_id"`
	KeyID          string `json:"key_id,omitempty"`
	Prompt         string `json:"prompt"`
	ConversationID string `json:"conversation_id"`
}

func buildAgentInput(request AgentRequest) (json.RawMessage, pipelineIdentity, error) {
	identity, err := identityFromTaskInput(request.TaskInput)
	if err != nil {
		return nil, pipelineIdentity{}, err
	}
	if request.AgentID == AssetCreatorAgentID {
		return append(json.RawMessage(nil), request.Input.JSON...), identity, nil
	}
	switch request.NodeID {
	case NodeIntent:
		initial, ok := request.Dependencies[NodeInitialize]
		if !ok {
			return nil, identity, missingDependency(NodeInitialize)
		}
		if err := json.Unmarshal(initial.JSON, &identity); err != nil || strings.TrimSpace(identity.Prompt) == "" {
			return nil, identity, fmt.Errorf("%w: initial prompt is required", ErrInvalidSubmission)
		}
		return mustJSON(map[string]any{"prompt": identity.Prompt}), identity, nil
	case NodeScenePlan:
		return dependencyJSON(request.Dependencies, NodeValidateSpec, identity)
	case NodeDesignShots:
		spec, err := dependencyValue(request.Dependencies, NodeValidateSpec)
		if err != nil {
			return nil, identity, err
		}
		assets, err := dependencyObjectField(request.Dependencies, NodeCreateAssets, "asset_manifests")
		if err != nil {
			return nil, identity, err
		}
		return mustJSON(map[string]any{"scene_spec": spec, "asset_manifests": assets}), identity, nil
	case NodeAssembleScene:
		assets, err := dependencyObjectField(request.Dependencies, NodeCreateAssets, "asset_manifests")
		if err != nil {
			return nil, identity, err
		}
		shots, err := dependencyValue(request.Dependencies, NodeDesignShots)
		if err != nil {
			return nil, identity, err
		}
		output := filepath.ToSlash(filepath.Join("tasks", request.TaskID, "scene.blend"))
		return mustJSON(map[string]any{"asset_manifests": assets, "shot_plan": shots, "output_path": output}), identity, nil
	default:
		return nil, identity, fmt.Errorf("%w: no input builder for agent node %q", ErrInvalidWorkflow, request.NodeID)
	}
}

func identityFromTaskInput(input Snapshot) (pipelineIdentity, error) {
	var identity pipelineIdentity
	if len(input.JSON) == 0 {
		return identity, nil
	}
	if err := json.Unmarshal(input.JSON, &identity); err != nil {
		return pipelineIdentity{}, fmt.Errorf("%w: invalid task input", ErrInvalidSubmission)
	}
	return identity, nil
}

// ShotPreviewSteps implements the deterministic workflow nodes. It delegates
// asset creation to the existing bounded fan-out coordinator.
type ShotPreviewSteps struct {
	Assets AssetFanOutStep
}

func (s ShotPreviewSteps) InvokeStep(ctx context.Context, request StepRequest) (json.RawMessage, error) {
	switch request.NodeID {
	case NodeInitialize:
		if !json.Valid(request.Input.JSON) {
			return nil, fmt.Errorf("%w: initial input must be JSON", ErrInvalidSubmission)
		}
		return append(json.RawMessage(nil), request.Input.JSON...), nil
	case NodeValidateSpec:
		raw, _, err := dependencyJSON(request.Dependencies, NodeIntent, pipelineIdentity{})
		if err != nil {
			return nil, err
		}
		var spec struct {
			Version string `json:"version"`
			SceneID string `json:"scene_id"`
		}
		if json.Unmarshal(raw, &spec) != nil || spec.Version != "scene-spec/v1" || strings.TrimSpace(spec.SceneID) == "" {
			return nil, fmt.Errorf("%w: invalid scene-spec/v1", ErrInvalidSubmission)
		}
		return raw, nil
	case NodeCreateAssets:
		return s.Assets.InvokeStep(ctx, request)
	case NodeInspect:
		value, err := dependencyValue(request.Dependencies, NodePreviewRender)
		if err != nil {
			return nil, err
		}
		return mustJSON(map[string]any{"passed": true, "preview": value}), nil
	case NodePublish:
		value, err := dependencyValue(request.Dependencies, NodeVerify)
		if err != nil {
			return nil, err
		}
		return mustJSON(map[string]any{"published": true, "artifact": value}), nil
	default:
		return nil, fmt.Errorf("%w: step %q", ErrMissingInvoker, request.Step)
	}
}

// RegistryToolInvoker executes production tools through the shared skill
// registry. Render nodes wait for completion and return an artifact snapshot.
type RegistryToolInvoker struct {
	Skills       agent.SkillRegistry
	PollInterval time.Duration
}

func (i RegistryToolInvoker) InvokeTool(ctx context.Context, request ToolRequest) (json.RawMessage, error) {
	arguments, err := buildToolInput(request)
	if err != nil {
		return nil, err
	}
	data, err := i.invoke(ctx, request.Tool, arguments)
	if err != nil {
		return nil, err
	}
	if request.Tool != productiontools.ToolBlenderRenderSubmit {
		return data, nil
	}
	var job struct {
		JobID string `json:"job_id"`
	}
	if json.Unmarshal(data, &job) != nil || job.JobID == "" {
		return nil, errors.New("render submit did not return a job_id")
	}
	return i.waitForRender(ctx, job.JobID)
}

func (i RegistryToolInvoker) invoke(ctx context.Context, toolID string, arguments json.RawMessage) (json.RawMessage, error) {
	if i.Skills == nil {
		return nil, fmt.Errorf("%w: tool registry", ErrMissingInvoker)
	}
	resolved, err := i.Skills.Resolve(ctx, toolID)
	if err != nil {
		return nil, err
	}
	invokable, ok := resolved.(tool.InvokableTool)
	if !ok {
		return nil, fmt.Errorf("tool %q is not invokable", toolID)
	}
	raw, err := invokable.InvokableRun(ctx, string(arguments))
	if err != nil {
		return nil, err
	}
	var envelope struct {
		OK   bool            `json:"ok"`
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal([]byte(raw), &envelope) != nil || !envelope.OK || !json.Valid(envelope.Data) {
		return nil, fmt.Errorf("tool %q returned an invalid envelope", toolID)
	}
	return envelope.Data, nil
}

func (i RegistryToolInvoker) waitForRender(ctx context.Context, jobID string) (json.RawMessage, error) {
	interval := i.PollInterval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		statusRaw, err := i.invoke(ctx, productiontools.ToolBlenderRenderStatus, mustJSON(map[string]string{"job_id": jobID}))
		if err != nil {
			return nil, err
		}
		var status struct {
			Status string `json:"status"`
		}
		if json.Unmarshal(statusRaw, &status) != nil {
			return nil, errors.New("render status is invalid")
		}
		switch status.Status {
		case string(productiontools.RenderSucceeded):
			return i.invoke(ctx, productiontools.ToolBlenderRenderArtifact, mustJSON(map[string]string{"job_id": jobID}))
		case string(productiontools.RenderFailed), string(productiontools.RenderCancelled):
			return nil, fmt.Errorf("render job %s", status.Status)
		}
		select {
		case <-ctx.Done():
			_, _ = i.invoke(context.Background(), productiontools.ToolBlenderRenderCancel, mustJSON(map[string]string{"job_id": jobID}))
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func buildToolInput(request ToolRequest) (json.RawMessage, error) {
	switch request.NodeID {
	case NodePreviewRender, NodeFinalRender:
		assembly, err := dependencyValue(request.Dependencies, NodeAssembleScene)
		if err != nil {
			return nil, err
		}
		object, ok := assembly.(map[string]any)
		blendFile, valid := stringField(object, "blend_file")
		if !ok || !valid {
			return nil, fmt.Errorf("%w: assembly blend_file is required", ErrInvalidSubmission)
		}
		kind, end := "preview", 1
		if request.NodeID == NodeFinalRender {
			kind, end = "final", 250
		}
		return mustJSON(map[string]any{
			"project_path": blendFile,
			"output_path":  filepath.ToSlash(filepath.Join("tasks", request.TaskID, kind, "frame_")),
			"frame_start":  1, "frame_end": end,
		}), nil
	case NodeEncode:
		return mustJSON(map[string]any{
			"input_path":  filepath.ToSlash(filepath.Join("tasks", request.TaskID, "final", "frame_%04d.png")),
			"output_path": filepath.ToSlash(filepath.Join("tasks", request.TaskID, "shot-preview.mp4")),
			"fps":         24, "codec": "libx264",
		}), nil
	case NodeVerify:
		encoded, err := dependencyValue(request.Dependencies, NodeEncode)
		if err != nil {
			return nil, err
		}
		object, ok := encoded.(map[string]any)
		outputPath, valid := stringField(object, "output_path")
		if !ok || !valid {
			return nil, fmt.Errorf("%w: encoded output_path is required", ErrInvalidSubmission)
		}
		return mustJSON(map[string]any{"path": outputPath}), nil
	default:
		return nil, fmt.Errorf("%w: no input builder for tool node %q", ErrInvalidWorkflow, request.NodeID)
	}
}

func stringField(object map[string]any, field string) (string, bool) {
	value, ok := object[field].(string)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}

func dependencyJSON(dependencies DependencyOutputs, id NodeID, identity pipelineIdentity) (json.RawMessage, pipelineIdentity, error) {
	snapshot, ok := dependencies[id]
	if !ok {
		return nil, identity, missingDependency(id)
	}
	return append(json.RawMessage(nil), snapshot.JSON...), identity, nil
}

func dependencyValue(dependencies DependencyOutputs, id NodeID) (any, error) {
	snapshot, ok := dependencies[id]
	if !ok {
		return nil, missingDependency(id)
	}
	var value any
	if err := json.Unmarshal(snapshot.JSON, &value); err != nil {
		return nil, fmt.Errorf("decode dependency %q: %w", id, err)
	}
	return value, nil
}

func dependencyObjectField(dependencies DependencyOutputs, id NodeID, field string) (any, error) {
	value, err := dependencyValue(dependencies, id)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok || object[field] == nil {
		return nil, fmt.Errorf("%w: dependency %q field %q is required", ErrInvalidSubmission, id, field)
	}
	return object[field], nil
}

func missingDependency(id NodeID) error {
	return fmt.Errorf("%w: dependency %q is required", ErrInvalidSubmission, id)
}

func mustJSON(value any) json.RawMessage {
	raw, _ := json.Marshal(value)
	return raw
}
