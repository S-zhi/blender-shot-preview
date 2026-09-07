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
		stageNode(NodeIntent, []NodeID{NodeInitialize}, nil),
		stageNode(NodeValidateSpec, []NodeID{NodeIntent}, nil),
		stageNode(NodeScenePlan, []NodeID{NodeValidateSpec}, nil),
		stageNode(NodeCreateAssets, []NodeID{NodeScenePlan}, nil),
		stageNode(NodeDesignShots, []NodeID{NodeScenePlan}, nil),
		stageNode(NodeAssembleScene, []NodeID{NodeCreateAssets, NodeDesignShots}, nil),
		stageNode(NodePreviewRender, []NodeID{NodeAssembleScene}, nil),
		stageNode(NodeInspect, []NodeID{NodePreviewRender}, nil),
		stageNode(NodeFinalRender, []NodeID{NodeInspect}, nil),
		stageNode(NodeEncode, []NodeID{NodeFinalRender}, nil),
		stageNode(NodeVerify, []NodeID{NodeEncode}, nil),
		stageNode(NodePublish, []NodeID{NodeVerify}, nil),
	}}
}

func stageNode(id NodeID, dependencies []NodeID, input json.RawMessage) NodeSpec {
	return NodeSpec{
		ID: id, DependsOn: dependencies, Input: input,
		Invocation: Invocation{Kind: InvocationStep, Target: string(id)},
	}
}
