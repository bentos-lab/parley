# WhatsApp Inbound

- **connect**: Run `parley connect whatsapp` to print a QR code and pair your WhatsApp account. The pairing data stays under `~/.bentos/parley/connect/whatsapp.db`. If a session already exists, the command prompts for confirmation before removing the old store and creating a fresh pairing.
- **serve**: When `parley serve` detects the stored session, it mounts a listener that watches direct messages to the paired account.
- **Commands**: Users must prefix their DM with `/parley`; `[parley]` marks assistant responses and is not reprocessed. The last 10 command exchanges (user + assistant) are fed into an LLM which classifies the intent (`create`, `resume`, `audio`, `delete`, `list`, `unknown`). Audio messages are ignored for history.
- **History persistence**: The listener saves only `/parley` command messages and `[parley]` replies to the JSON cache at `~/.bentos/parley/connect/whatsapp.history.json`, reloads it on startup, and keeps the most recent 10 entries per chat so that history survives restarts.
- **Behavior**:
  - Only commands that the linked account posts to its own “Saved Messages” chat (where the chat and sender are identical) are processed; other DMs are ignored.
  - `create`: Follows the CLI workflow (name, agents, rounds) without printing to stdout. Automatically synthesizes audio, encodes it as an OGG/Opus voice note via the `github.com/hraban/opus` encoder, and sends that voice note with a confirmation message.
  - `resume`: Appends the requested rounds, regenerates the debate audio, and sends the updated voice note.
  - `audio`: Generates audio for the given debate ID and delivers the converted voice note (or a WAV document with a warning if conversion fails).
  - `list`: Returns stored debate IDs.
  - `delete`: Removes the specified debate.
  - `unknown`: Replies that the command could not be determined.
- **Responses** always begin with `[parley]` and include an audio notification message before the file transfer.
  - Every response that references a debate explicitly mentions the active debate ID so you can correlate replies with the correct conversation.
  - When voice-note conversion fails the listener falls back to sending the WAV document while mentioning the conversion issue.
