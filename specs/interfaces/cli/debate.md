# Command Line Interface

Refer to the features guide in `specs/features/debates.md`.

## Create a Debate

Usage:

```sh
<cli> create [topic] [--num-agents int default 2] [--num-rounds int default 10] [--tts-provider string default from config] [--llm-provider string] [--llm-model string]
```

Behavior:

- `topic`: Debate topic used to generate the debate name and agents.
- `--num-agents`: Number of agents to auto-generate.
- `--num-rounds`: Maximum number of rounds to generate.
- `--tts-provider`: Optional TTS provider override for this debate (default: config `tts.provider`, which is `native` when unset).
- `--llm-provider`: Optional LLM provider override for this debate (default: config `llm.provider`).
- `--llm-model`: Optional LLM model override for this debate (default: provider-specific config model).
- Create a new debate with the given topic. The name and agents are generated automatically.
- Immediately print a basic header with app/model info and the topic.
- Generate the debate name and print it as muted text.
- Generate the debate agents and print them in the same format as the existing agents table.
- Print each generated round as a card with two bullets: `Voice` (formatted message) and `Summary` (plain text).
- After rounds finish, generate the debate summary and print it, including the conclusion.
- Save the debate to a file.
- Run a loop for up to the specified number of rounds.
- For each round, generate it automatically and save the debate.
- After the loop finishes, print the saved filename plus the debate ID to highlight the identifier that other commands accept.

## Resume a Debate

Usage:

```sh
<cli> resume [id] [--num-rounds int default 10] [--llm-provider string] [--llm-model string]
```

Behavior:

- `id`: Debate identifier produced by `create` or listed via `list` (omit the `.json` suffix).
- `--num-rounds`: Maximum number of rounds to generate.
- `--llm-provider`: Optional LLM provider override for this resume run.
- `--llm-model`: Optional LLM model override for this resume run.
- Load the debate by converting the ID to the stored filename.
- Print a debate header that includes the debate name, topic, identifier, and agent details.
- Print each generated round as a card with two bullets: `Voice` (formatted message) and `Summary` (plain text).
- After rounds finish, generate the debate summary and print it, including the conclusion.
- Run a loop for up to the specified number of rounds.
- For each round, generate it automatically and save the debate.

## Generate Audio

Usage:

```sh
<cli> audio [id] [--tts-provider string]
```

Behavior:

- `id`: Debate identifier to generate audio for (omit the `.json` suffix).
- `--tts-provider`: Optional TTS provider override for this audio run.
- Load the debate by converting the ID to the stored filename.
- If `--tts-provider` is omitted, use the stored debate `tts_provider`.
- When the debate has no stored `tts_provider`, fall back to the configured default so audio can still be generated.
- Generate the audio file from the debate.
- Set the audio path in the debate file.
- Save the file.
- Print the audio path.

## List Debates

Usage:

```sh
<cli> list [--format pretty|json]
```

Behavior:

- Show only the debate identifiers stored on disk (no filenames).
- The `pretty` format prints newline-separated IDs while `json` emits a structured payload keyed by `type`.
- Alias: `ls`.

## Delete a Debate

Usage:

```sh
<cli> delete [id]
```

Behavior:

- `id`: Debate identifier to delete (omit the `.json` suffix).
- Remove the corresponding saved file and return success or an error if the debate is missing.

## Desktop launcher
Running `parley` without arguments starts the HTTP server and WhatsApp listener while showing a small Fyne window with **Open** (launches `http://localhost:8080`) and **Exit** (stops the services and exits). The same behavior occurs when double-clicking the binaries produced by `build/build.sh` (Linux/macOS `tar.gz`, Windows `.zip`), giving a native launcher while retaining CLI subcommands such as `parley create`.
