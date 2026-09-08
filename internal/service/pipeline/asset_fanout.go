package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

const AssetCreatorAgentID = "asset-creator-agent"

// AssetFanOutStep turns the planner's asset_tasks into concurrent invocations
// of one reusable asset-creator agent definition. Results retain planner order
// so downstream snapshots are deterministic.
type AssetFanOutStep struct {
	Agents      AgentInvoker
	MaxParallel int
}

func (s AssetFanOutStep) InvokeStep(ctx context.Context, request StepRequest) (json.RawMessage, error) {
	if request.Step != string(NodeCreateAssets) {
		return nil, fmt.Errorf("asset fan-out cannot execute step %q", request.Step)
	}
	if s.Agents == nil {
		return nil, fmt.Errorf("%w: %s", ErrMissingInvoker, AssetCreatorAgentID)
	}
	planner, ok := request.Dependencies[NodeScenePlan]
	if !ok {
		return nil, fmt.Errorf("%w: %s output is required", ErrInvalidSubmission, NodeScenePlan)
	}
	var plan struct {
		AssetTasks []json.RawMessage `json:"asset_tasks"`
	}
	if err := json.Unmarshal(planner.JSON, &plan); err != nil || len(plan.AssetTasks) == 0 {
		return nil, fmt.Errorf("%w: scene plan must contain asset_tasks", ErrInvalidSubmission)
	}
	parallel := s.MaxParallel
	if parallel <= 0 || parallel > len(plan.AssetTasks) {
		parallel = len(plan.AssetTasks)
	}
	manifests := make([]json.RawMessage, len(plan.AssetTasks))
	work := make(chan int)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var workers sync.WaitGroup
	for range parallel {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range work {
				input, err := newSnapshot(plan.AssetTasks[index], request.Input.CapturedAt)
				if err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}
				output, err := s.Agents.InvokeAgent(ctx, AgentRequest{
					TaskID: request.TaskID, NodeID: NodeID(fmt.Sprintf("%s-%d", NodeCreateAssets, index+1)),
					AgentID: AssetCreatorAgentID, Attempt: request.Attempt,
					TaskInput: cloneSnapshot(request.TaskInput), Input: input,
				})
				if err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}
				manifests[index] = append(json.RawMessage(nil), output...)
			}
		}()
	}
dispatch:
	for index := range plan.AssetTasks {
		select {
		case work <- index:
		case <-ctx.Done():
			break dispatch
		}
	}
	close(work)
	workers.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return json.Marshal(struct {
		AssetManifests []json.RawMessage `json:"asset_manifests"`
	}{AssetManifests: manifests})
}
