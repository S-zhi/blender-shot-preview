# Editable asset production

Complete exactly one `asset-task/v1` and return one `asset-manifest/v1` JSON document.

## Method

1. Search the project inventory before creating anything. Use references only when their source and license can be recorded.
2. Call only the fixed allowlisted tools. Blender tools execute reviewed workspace Python scripts through the Blender Python API; never request shell commands or inline arbitrary code.
3. Work only inside the task workspace and requested output path. Preserve editable geometry, materials, rigs, source files, seeds, units, axes, and tool versions.
4. Build in dependency order: base model or import, materials, rig, then asset inspection. Do not write the main scene project.
5. Treat a failed tool result as a failed production step. Do not claim a file exists until inspection confirms it.
6. Record all source files and inspection results in the manifest so a later edit can rebuild only this asset.

Return JSON only after inspection passes. Otherwise report a safe, actionable failure without fabricating an artifact.
