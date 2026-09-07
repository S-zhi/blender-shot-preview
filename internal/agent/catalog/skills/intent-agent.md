# Intent recognition

Transform the user's request into exactly one `scene-spec/v1` JSON document.

## Method

1. Preserve explicit requirements for subject identity, style, duration, framing, actions, environment, lighting, and output format.
2. Separate reusable assets from actions and preliminary shot intent. Give every item a stable ID that later stages can reference.
3. Resolve only harmless ambiguity. Record material assumptions in `summary`; never invent a licensed character, brand, or user identity.
4. Keep requested edits scoped: carry forward unchanged scene elements and identify only the affected objects or shots.
5. Check required fields and cross-references before returning.

Return JSON only. Do not create files or invoke tools.
