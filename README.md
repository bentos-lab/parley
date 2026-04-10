# Parley

**Parley** brings structured AI debates within reach, blending CLI power with a polished WebUI for every kind of workflow.

## Install

### Stable

#### Linux/macOS

```bash
curl -fsSL https://raw.githubusercontent.com/bentos-lab/parley/master/install.sh | bash
```

#### Windows (powershell)

```bash
iwr https://raw.githubusercontent.com/bentos-lab/parley/master/install.ps1 -useb | iex
```

### Latest

Prerequisites:
- Go `1.26`

```bash
go install https://github.com/bentos-lab/parley/cmd/parley
```

## Setup

Authenticate your LLM/TTS provider first:

```bash
parley login llm    # authenticate your LLM provider
parley login tts    # authenticate your TTS provider
```

Run `parley config` as needed to adjust other preferences. Configuration reference is in [docs/configuration.md](/docs/configuration.md).

For example:
```bash
parley config llm.openai.base_url=http://localhost:8000/v1
```

## Whatsapp

Optional: connect your WhatsApp account:
```bash
parley connect whatsapp
```

## Desktop app

After installing, double-click the executable to start the server and open the WebUI if you prefer not to use the terminal.  

Common binary locations:
- Linux/macOS: `/usr/local/bin/parley`
- Windows: `C:\Users\<your_user>\bin\parley.exe`


## CLI usage

Start a debate:

```bash
parley create "Should cities ban cars downtown?" --num-agents 2 --num-rounds 3
```

Show a saved debate:

```bash
parley get <debate_id>
```

Run `parley --help` for more details.

## Web UI

Launch the web interface with:

```bash
parley serve
```

Open `http://localhost:8080` in your browser to use the Web UI.
