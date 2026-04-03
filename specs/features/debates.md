# Debates

A debate stores information about a single debate session.

## Data Model

- `Name`: Display name of the debate.
- `NormalizedName`: Lowercase alphanumeric name (`a-z`, `0-9`) derived from `Name`.
- `Topic`: Topic of the debate.
- `Agents`: Participating agents.
- `Agent` fields: `id` (identifier), `name` (display name), `stance` (position), `voice_name` (selected TTS voice).
- If an agent `id` is missing, it is auto-generated in order as `agent-1`, `agent-2`, and so on.
- `Rounds`: Ordered list of debate rounds.
- `Round` fields: `agent_id` (empty means the user, not an agent), `message` (round content), `weakness`, `new_point`, `rebuttal`, `summary`.
- `TTSProvider`: Preferred TTS provider for the debate.

## Features

### Create

Create a new debate with the following information:

- `name`: If empty, generate it from the `topic`. The name generator must be generic.
- `agents` and `stances`: If not provided, generate them from the `topic`.
- `num_agents`: Optional auto-generation of agents based on a requested count. The agent generator must be generic.

### Save

- Save the debate as a JSON file.
- Always store files under `HOME/.bentos/parley/`.

### Load

- Load a debate from a previously saved JSON file.
- Always load files from `HOME/.bentos/parley/`.

### Edit

Allow editing of debate data, even while the debate is in progress:

- `topic`
- `agents` details
- `rounds` messages

### Generate a Round

- Choose an `agent_id` explicitly or leave it empty.

If no agent is specified, select one using the following process:

1. Count how many times each agent has joined the debate so far, excluding the most recent speaking agent.
2. Randomly select an agent. Agents with fewer appearances should have a higher selection probability.

Generate the next round based on history, the selected agent's stance, and the debate topic.

### Generate Audio

- Use a TTS provider (generic) to generate voice for the rounds.
- Concatenate the rounds with approximately 3 seconds of padding between them.
- Save the output as a WAV file.
