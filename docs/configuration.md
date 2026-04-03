# Configuration

Primary configuration is stored in `~/.bentos/parley.config.toml`. Environment variables (from
`.env` or your shell) override config file values.

## TOML Configuration (with env + defaults)

`[llm]`

- `provider`: LLM provider name. Default: `openai`. Supported values: `openai`.

`[llm.openai]`
- `base_url`: Base URL for the OpenAI-compatible LLM provider. Env: `OPENAI_BASE_URL`. Default:
  `https://api.openai.com/v1`.
- `api_key`: API key for the LLM provider. Env: `OPENAI_API_KEY`. Default: none. Required.
- `model`: Model identifier used for generation. Env: `OPENAI_MODEL`. Default: `gpt-4.1-mini`.

`[tts]`

- `provider`: TTS provider name. Default: `native`. Supported values: `native`, `inworld`.

`[tts.inworld]`
- `api_key`: API key for Inworld TTS. Env: `INWORLD_API_KEY`. Default: none.
- `model`: Model ID for Inworld TTS. Env: `INWORLD_MODEL`. Default: `inworld-tts-1.5-max`.
