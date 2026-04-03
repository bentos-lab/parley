# Parley

**Parley** makes it easy to run structured debates between AI agents. It supports both CLI and WebUI.

## Install
1. Install Go 1.22 or newer.
2. Build the `parley` binary from the repository root:

```bash
go build ./cmd/parley
```

The result is an executable you can run directly for both CLI and server modes.

## Setup
Use the guided commands to connect providers and configure defaults:

```bash
parley login llm    # authenticate your LLM provider
parley login tts    # authenticate your TTS provider
```

### LLM provider options

`parley login llm` walks through configuring the remote LLM provider (OpenAI, Anthropic, Gemini, or a custom OpenAI-compatible endpoint). Enter the base URL, API key, and model as prompted, and the same values can be supplied via `OPENAI_BASE_URL`, `OPENAI_API_KEY`, and `OPENAI_MODEL` in `.env` or `~/.bentos/parley.config.toml`.

Run `parley config` as needed to adjust other preferences (e.g., models, TTS provider hints). Full configuration reference and default file locations are in [docs/configuration.md](/docs/configuration.md).

## CLI usage
Start a debate by supplying the topic and sizing options:

```bash
parley create "Should cities ban cars downtown?" --num-agents 2 --num-rounds 3
```

Required argument: the debate topic. Optional flags such as `--num-agents` and `--num-rounds` let you control how many participants take part and how many rounds the debate runs.

Use `parley list` (alias `ls`) to display saved debate IDs, then plug the shown ID into `parley resume <id>`, `parley audio <id>`, or `parley delete <id>` when you want to continue, synthesize audio, or clean up the debate.

Run `parley connect whatsapp` to pair a WhatsApp account via QR. When `parley serve` detects the linked session, it listens to `/parley`-prefixed DMs and executes create/list/resume/delete/audio commands through the same debate workflows (responses always begin with `[parley]`). Audio outputs are delivered as WAV documents automatically after a create/resume/audio request.

Run `parley --version` (or `parley -v`) to see the embedded binary version before troubleshooting or reporting issues.

## Web UI
Launch the web interface with:

```bash
parley serve
```

After `serve` begins, open the address printed in your terminal (default http://localhost:8080) to explore debates, step through rounds, and trigger audio generation.

## Desktop launcher
Running `parley` without arguments now starts the HTTP server and WhatsApp listener and displays a small Fyne window with **Open** (launches http://localhost:8080) and **Exit** (stops the services and quits). The same behavior happens when you double-click any binary produced by `build/build.sh` (Linux/macOS `tar.gz`, Windows `.zip`), so the packaged apps behave like native launchers while CLI subcommands (`parley <command>`) remain available.
