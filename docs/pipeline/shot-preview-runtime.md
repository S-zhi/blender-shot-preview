# Shot preview pipeline runtime

The `internal/service/pipeline` package owns the durable shot-preview DAG and
the adapters that turn persisted node snapshots into agent and production-tool
calls. RPC handlers should enter the pipeline through
`service.ShotPreviewService`; bootstrap code constructs its runner through the
shared runtime boundary:

```go
runner := pipeline.NewShotPreviewRunner(repository, agentService, skillRegistry)
shotPreviewService := service.NewShotPreviewService(runner)
```

The constructor deliberately requires its production dependencies instead of
creating defaults. The bootstrap layer remains responsible for:

- a durable `pipeline.Repository` implementation;
- an `agent.AgentService` backed by registered production agent definitions and
  a real `agent.ModelResolver`;
- an `agent.SkillRegistry` populated with
  `productiontools.RegisterProductionTools`;
- calling `runner.Recover(ctx)` during startup after dependencies are ready.

`pipeline.MemoryRepository` is suitable for tests and local experiments only.
The server must not report the pipeline as available until all three production
dependencies have been initialized successfully.

## Workflow contract

`pipeline.ShotPreviewWorkflow` creates the stable `shot-preview/v1` graph:

```text
Initialize -> Intent -> ValidateSpec -> ScenePlan -> CreateAssets
                                      |              |
                                      +----------> DesignShots
                                                     |
                                              AssembleScene
                                                |       |
                                         PreviewRender |
                                                |       |
                                             Inspect    |
                                                +--> FinalRender
                                                        |
                                               Encode -> Verify -> Publish
```

`CreateAssets` is a bounded fan-out step. It runs one `asset-creator-agent` per
planned asset concurrently, preserves planner order when joining results, and
passes the original task identity into every child invocation.

Every node receives an immutable task-level input snapshot in addition to its
own node input and completed dependency outputs. The shot-preview service stores
the following task input:

```json
{
  "user_id": "user-1",
  "prompt": "Create a forest establishing shot",
  "conversation_id": "conversation-1",
  "request_id": "request-1"
}
```

Runtime adapters project only the fields required by each agent or tool. Agent
outputs and tool envelopes must contain valid JSON. Render submission is
followed by status polling and artifact retrieval; cancelling the pipeline also
attempts to cancel an active render job.

## Parallel development boundary

The runtime package and the RPC/API layer can evolve independently when they
keep the following boundary stable:

- API code depends on `service.ShotPreviewService`, never on individual DAG
  nodes or tools.
- Runtime code depends on `pipeline.Repository`, `agent.AgentService`, and
  `agent.SkillRegistry`, never on a concrete RPC transport.
- Bootstrap code is the only place that assembles concrete repositories, model
  providers, agent definitions, and production tools.

Run `go test ./...` and `go vet ./...` after changing this boundary. Runtime
adapter behavior is covered by `internal/service/pipeline/runtime_test.go`.
