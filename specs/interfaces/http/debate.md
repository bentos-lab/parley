# HTTP API

Refer to the features guide in `specs/features/debates.md`.

## `POST /api/debates`

Description: Creates a new debate.

Request:

- `name`: Debate name.
- `topic`: Debate topic.
- `agents` or `num_agents`: Mutually exclusive inputs.
- `tts_provider`: Optional TTS provider override for this debate (default: config `tts.provider`, which is `native` when unset).
- `agent_voices`: Optional map of `agent_id` to `voice_name` (validated against the provider).

Response:

- `name`: Debate name.
- `id`: Persisted debate identifier.

Procedure:

- Initialize a new `Debate` object with the input above. Auto-generate agents if `num_agents` is greater than 0.
- If any agent `id` values are missing, assign `agent-<n>` identifiers in order.
- If `agent_voices` is provided, verify all voices exist in the selected provider.
- Persist `tts_provider` and any provided `voice_name` values in the debate.
- Save the debate to a file.
- The `id` must follow this format: `{normalized-name}.{yyyy-mm-dd-hh-mm-ss}`.
- The `normalized-name` contains only lowercase letters (`a-z`) and digits (`0-9`).

## `GET /api/debates`

Description: Lists all debates on the user's machine.

Request:

- Empty.

Response:

- Array of items containing `id` (debate identifier), `name` (debate name), and `topic` (debate topic).
- Items are ordered by `id` timestamp descending (newest first).

Procedure:

- Load all `.json` files from `HOME/.bentos/parley/`.
- Discard invalid files.
- Return the list of topics, names, and ids.

## `GET /api/debates/{id}`

Description: Fetches a saved debate by id.

Request:

- Empty.

Response:

- Full debate payload, including `summary` and rounds with `agent_id`, `message`, `weakness`, `new_point`, `rebuttal`, and `summary`.

Procedure:

- Load the debate from the specified id by appending `.json` to form the filename.
- If the id already includes `.json`, treat it as literal input (resulting in `.json.json` on disk).
- Return the debate JSON.

## `PUT /api/debates/{id}`

Description: Updates a debate.

Request:

- Full debate content, including agents and rounds.

Response:

- Empty.

Procedure:

- Overwrite the file for the id by appending `.json` to form the filename.

## `GET /api/debates/{id}/summary`

Description: Fetches the debate summary for a saved debate by id.

Request:

- Query `new=true` to force regenerating the summary.

Response:

- `agents`: Array of arrays of summary points, where index `i` corresponds to the debate agent at index `i` in the saved debate payload.
- `final_conclusion`: Final debate conclusion.

Procedure:

- Load the debate from the id by appending `.json`.
- If the debate has no rounds, return an error.
- If the debate already has a summary and `new` is not set, return the stored summary.
- Otherwise, generate a new summary with the LLM, persist it, and return it.

## `DELETE /api/debates/{id}`

Description: Deletes a debate.

Request:

- Empty.

Response:

- Empty.

Procedure:

- Delete the corresponding file for the id by appending `.json`.

## `POST /api/debates/{id}/rounds`

Description: Creates a new round.

Request:

- `agent_id`: Agent identifier (optional).
- `content`: Round content (optional).

Response:

- `agent_id`: Agent identifier.
- `content`: Round content.
- `weakness`: Weakness in the opponent's point.
- `new_point`: New point with evidence.
- `rebuttal`: Rebuttal that quotes the opponent.
- `summary`: 1–2 sentence plain-text summary of the spoken reply.

Procedure:

- Load the debate from the id by appending `.json`.
- If `content` is empty, call the automatic round generator. Otherwise, append the content to the rounds.
- Save the file.
- Return the generated content.

## `GET /api/debates/{id}/rounds/{index}`

Description: Fetches a single debate round by index.

Request:

- `index`: 0-based round index.

Response:

- `agent_id`: Agent identifier.
- `content`: Round content.
- `weakness`: Weakness in the opponent's point.
- `new_point`: New point with evidence.
- `rebuttal`: Rebuttal that quotes the opponent.
- `summary`: 1–2 sentence plain-text summary of the spoken reply.

Procedure:

- Load the debate from the id by appending `.json`.
- Validate `index` is within bounds.
- Return the selected round.

## `GET /api/debates/{id}/rounds/sse`

Description: Streams new rounds via Server-Sent Events (SSE).

Request:

- `n`: Maximum number of rounds.

Response:

- Each event includes `agent_id`, `content`, `weakness`, `new_point`, `rebuttal`, and `summary`.

Procedure:

- Continue while the request context is active, and stop after `n` rounds.
- Load the debate from the id by appending `.json`.
- Call the automatic round generator.
- Save the file.
- Emit an event with the generated content.

## `GET /api/debates/{id}/rounds/{index}/audio`

Description: Fetches the audio file for a debate round.

Request:

- `index`: 0-based round index.

Response:

- Binary content of the round audio file.

Procedure:

- Load the debate using the id by appending `.json`.
- Resolve the debate `tts_provider` (fall back to the configured default when the debate has no provider).
- Compute the round signature and expected `signature.wav` path.
- If the file is missing, synthesize and save it.
- Return the file as binary.

## `GET /api/debates/{id}/audio`

Description: Fetches the audio file for a debate.

Request:

- Empty.

Response:

- Binary content of the audio file.

Procedure:

- Load the debate using the id by appending `.json`.
- Resolve the debate `tts_provider` (fall back to the configured default when the debate has no provider).
- Compute the debate signature (hash of ordered round signatures) and expected `signature.wav` path.
- If the file is missing, synthesize:
  - Reuse existing round audio when present.
  - Otherwise, synthesize and save each round audio as `signature.wav`.
- Return the file as binary.
