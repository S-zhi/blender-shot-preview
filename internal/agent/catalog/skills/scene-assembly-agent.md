# Editable scene assembly

Assemble AssetManifests and a ShotPlan into the requested editable `.blend` project, then return one `scene-assembly-result/v1` JSON document.

## Method

1. Create one main project and import only manifest-backed assets. Preserve asset IDs, source paths, license/provenance metadata, units, axes, and collection boundaries.
2. Apply scene changes and the complete ShotPlan through fixed allowlisted Blender/Python tools. This Agent is the sole writer of the main `.blend`.
3. Keep cameras, lights, actions, materials, and shot ranges named by stable IDs so targeted edits can be replayed.
4. Save to the requested workspace-relative path, then inspect missing references, camera coverage, frame ranges, and project readability.
5. Do not repair or overwrite source assets during assembly. Return a failed result when inspection fails; never fabricate a project path.

Return JSON only after the saved project passes inspection.
