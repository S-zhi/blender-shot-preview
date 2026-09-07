package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAssetFanOutStepInvokesCreatorConcurrentlyAndPreservesOrder(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var once sync.Once
	step := AssetFanOutStep{
		MaxParallel: 2,
		Agents: AgentInvokerFunc(func(_ context.Context, request AgentRequest) (json.RawMessage, error) {
			started <- struct{}{}
			<-release
			return json.RawMessage(fmt.Sprintf(`{"node":%q}`, request.NodeID)), nil
		}),
	}
	plan := snapshot(t, map[string]any{"asset_tasks": []map[string]string{{"task_id": "one"}, {"task_id": "two"}}})
	done := make(chan json.RawMessage, 1)
	go func() {
		output, err := step.InvokeStep(context.Background(), StepRequest{
			TaskID: "task", NodeID: NodeCreateAssets, Step: string(NodeCreateAssets),
			Attempt: 1, Input: snapshot(t, map[string]string{"request": "assets"}),
			Dependencies: DependencyOutputs{NodeScenePlan: plan},
		})
		if err != nil {
			once.Do(func() { t.Errorf("InvokeStep(): %v", err) })
		}
		done <- output
	}()
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("asset creators did not run concurrently")
		}
	}
	close(release)
	var output struct {
		AssetManifests []map[string]string `json:"asset_manifests"`
	}
	if err := json.Unmarshal(<-done, &output); err != nil {
		t.Fatal(err)
	}
	if len(output.AssetManifests) != 2 || output.AssetManifests[0]["node"] != "CreateAssets-1" || output.AssetManifests[1]["node"] != "CreateAssets-2" {
		t.Fatalf("manifests = %#v", output.AssetManifests)
	}
}

func TestAssetFanOutStepCancelsRemainingWorkAfterFailure(t *testing.T) {
	started := make(chan NodeID, 2)
	releaseFailure := make(chan struct{})
	cancelled := make(chan struct{}, 1)
	step := AssetFanOutStep{
		MaxParallel: 2,
		Agents: AgentInvokerFunc(func(ctx context.Context, request AgentRequest) (json.RawMessage, error) {
			started <- request.NodeID
			if request.NodeID == NodeID("CreateAssets-1") {
				<-releaseFailure
				return nil, fmt.Errorf("asset creation failed")
			}
			<-ctx.Done()
			cancelled <- struct{}{}
			return nil, ctx.Err()
		}),
	}
	plan := snapshot(t, map[string]any{"asset_tasks": []map[string]string{{"task_id": "one"}, {"task_id": "two"}}})
	done := make(chan error, 1)
	go func() {
		_, err := step.InvokeStep(context.Background(), StepRequest{
			TaskID: "task", NodeID: NodeCreateAssets, Step: string(NodeCreateAssets),
			Attempt: 1, Input: snapshot(t, map[string]string{"request": "assets"}),
			Dependencies: DependencyOutputs{NodeScenePlan: plan},
		})
		done <- err
	}()
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("both workers were not started")
		}
	}
	close(releaseFailure)
	if err := <-done; err == nil || err.Error() != "asset creation failed" {
		t.Fatalf("InvokeStep() error = %v", err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("peer invocation did not observe cancellation")
	}
}
