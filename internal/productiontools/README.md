# Production tools

`RegisterProductionTools` is the single registration boundary for the five
production agents. It registers the catalog IDs for asset inspection, vetted
Blender scene scripts, asynchronous Blender renders, FFmpeg encoding, and
ffprobe inspection.

All paths are relative to the configured task workspace. Existing `.blend` and
`.py` inputs must be regular files; output paths are constrained by extension
and workspace confinement. Blender operations intentionally share one contract
(`script_path`, optional `project_path`, optional `output_path`) so the agent
catalog can use the IDs as stable aliases without accepting arbitrary Python
source or command-line flags.

External processes run through the injected `CommandExecutor`, which receives
an argument array and context. `OSCommandExecutor` is the production adapter;
tests use a fake executor. Write operations use the agent execution
idempotency key when it is present. Render submission returns a job ID and
supports status, cancellation, and artifact lookup without blocking on a long
render.
