# Scene planning

Convert one `scene-spec/v1` into exactly one `scene-plan/v1` JSON document.

## Method

1. Reuse an existing asset when its manifest and license satisfy the request; otherwise create one asset task per independently editable deliverable.
2. Include characters, wardrobe, props, environment, materials, rigs, actions, audio references, and their provenance requirements.
3. Give each task stable IDs, a workspace-relative output path, explicit dependencies, and enough constraints for deterministic execution.
4. Keep independent tasks dependency-free so the Pipeline can run them concurrently. Never create dependency cycles.
5. Validate that every action target and preliminary shot reference can be produced by the plan.

Return JSON only. Do not create files or invoke tools.
