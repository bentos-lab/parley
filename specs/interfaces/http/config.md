# HTTP API – Configuration

## `GET /api/config`

Description: Returns the runtime configuration that the HTTP server is currently using for LLM and TTS providers. The response reflects defaults, values loaded from `HOME/.bentos/parley.config.toml`, and any overriding environment variables.

- Response (JSON):
-
- `llm.provider`: Currently active LLM provider (`openai`).
- `llm.openai.base_url`: The OpenAI-compatible base URL.
- `llm.openai.api_key`: The API key currently in use (empty if not configured).
- `llm.openai.model`: The active model.
- `tts.provider`: Either `native` or `inworld`.
- `tts.inworld.api_key`: The configured Inworld API key (empty if not set).
- `tts.inworld.model`: The configured Inworld model.

## `PUT /api/config`

Description: Updates one or both configuration sections (`llm` and `tts`) in a single request. Only the keys supplied in the payload are written; missing objects or fields are left untouched.

Request (JSON):

- `llm` (optional)
  - `provider`: Optional override (supported values `openai`).
  - `openai.base_url`: Optional new API URL.
  - `openai.api_key`: Optional API key (stored verbatim).
  - `openai.model`: Optional model override.
- `tts` (optional)
  - `provider`: Optional provider override. Valid values are `native` and `inworld`.
    - Switching to `inworld` requires providing `tts.inworld.api_key` even if the key already exists on disk.
    - Switching to `native` triggers a readiness check for the native TTS tool; if the tool is missing and an install command is configured, the server runs it before persisting the config. A failure in this step returns a 500 and leaves the file untouched.
  - `inworld.api_key`: Optional API key for the Inworld provider.
  - `inworld.model`: Optional model for the Inworld provider.

Errors:

- 400 for invalid payloads, unsupported providers, or when no configurable fields were supplied.
- 500 when the server cannot read/write the config file or when a native tool installation fails.

Response: Same JSON shape as `GET /api/config`, representing the runtime configuration after the update succeeds.
