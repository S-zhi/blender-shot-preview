package pipeline

import "encoding/json"

const (
	NodeInitialize    NodeID = "Initialize"
	NodeIntent        NodeID = "Intent"
	NodeValidateSpec  NodeID = "ValidateSpec"
	NodeScenePlan     NodeID = "ScenePlan"
	NodeCreateAssets  NodeID = "CreateAssets"
	NodeDesignShots   NodeID = "DesignShots"
	NodeAssembleScene NodeID = "AssembleScene"
	NodePreviewRender NodeID = "PreviewRender"
	NodeInspect       NodeID = "Inspect"
	NodeFinalRender   NodeID = "FinalRender"
	NodeEncode        NodeID = "Encode"
	NodeVerify        NodeID = "Verify"
	NodePublish       NodeID = "Publish"
)

// ShotPreviewWorkflow is the stable MVP graph for a production request. Asset
// discovery can replace CreateAssets with FanOutNodes before Submit without
// changing the Runner's dependency semantics.
func ShotPreviewWorkflow(initialInput json.RawMessage) Workflow {
	return Workflow{Nodes: []NodeSpec{
		stageNode(NodeInitialize, nil, initialInput),
		agentNode(NodeIntent, []NodeID{NodeInitialize}, "intent-agent"),
		stageNode(NodeValidateSpec, []NodeID{NodeIntent}, nil),
		agentNode(NodeScenePlan, []NodeID{NodeValidateSpec}, "scene-planner-agent"),
		// CreateAssets is a deterministic fan-out coordinator. It invokes one
		// asset-creator-agent run per planned asset and joins the manifests.
		stageNode(NodeCreateAssets, []NodeID{NodeScenePlan}, nil),
		agentNode(NodeDesignShots, []NodeID{NodeValidateSpec, NodeCreateAssets}, "shot-designer-agent"),
		agentNode(NodeAssembleScene, []NodeID{NodeCreateAssets, NodeDesignShots}, "scene-assembly-agent"),
		toolNode(NodePreviewRender, []NodeID{NodeAssembleScene}, "blender.render.submit"),
		stageNode(NodeInspect, []NodeID{NodePreviewRender}, nil),
		toolNode(NodeFinalRender, []NodeID{NodeAssembleScene, NodeInspect}, "blender.render.submit"),
		toolNode(NodeEncode, []NodeID{NodeFinalRender}, "ffmpeg.encode"),
		toolNode(NodeVerify, []NodeID{NodeEncode}, "ffprobe.inspect"),
		stageNode(NodePublish, []NodeID{NodeVerify}, nil),
	}}
}

func stageNode(id NodeID, dependencies []NodeID, input json.RawMessage) NodeSpec {
	return NodeSpec{
		ID: id, DependsOn: dependencies, Input: input,
		Invocation: Invocation{Kind: InvocationStep, Target: string(id)},
	}
}

func agentNode(id NodeID, dependencies []NodeID, agentID string) NodeSpec {
	return NodeSpec{
		ID: id, DependsOn: dependencies,
		Invocation: Invocation{Kind: InvocationAgent, Target: agentID},
	}
}

func toolNode(id NodeID, dependencies []NodeID, toolID string) NodeSpec {
	return NodeSpec{
		ID: id, DependsOn: dependencies,
		Invocation: Invocation{Kind: InvocationTool, Target: toolID},
	}
}
