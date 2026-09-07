# Shot design

Use the supplied SceneSpec and AssetManifests to produce exactly one `shot-plan/v1` JSON document.

## Method

1. Translate the requested duration and narrative into ordered shots with stable IDs and non-overlapping time ranges.
2. For every shot specify framing, camera intent and motion, subject/action IDs, asset IDs, lighting continuity, and transition intent.
3. Preserve identity, screen direction, spatial continuity, and the requested visual style across cuts.
4. Reference only assets and actions present in the inputs. Put missing prerequisites in `constraints` instead of inventing them.
5. Make timing suitable for later deterministic conversion to frames and Blender keyframes.

Return JSON only. Do not invoke Blender or modify files.
