# Changelog

## [0.1.5] - 2026-04-04
### Added
- Introduced a WhatsApp integration with REST endpoints for status, pairing, and session cleanup, including an SSE-based QR pairing stream and matching HTTP API documentation.
- Expanded the Web UI with settings and WhatsApp integration screens, plus provider catalog metadata to drive configuration UX.
- Added audio seek support and a mini player experience for debate rounds, including caching and playback helpers.
- Included peer configuration files in the repository.

### Changed
- Refined the CLI and REST WhatsApp connect flow to align with the new pairing endpoints and session management.
- Refreshed core Web UI layouts and routes across debates, audio, and navigation surfaces.

### Fixed
- Run the WhatsApp listener asynchronously to keep REST handlers responsive.
- Correct the return value order in the WhatsApp connect routine.
- Remove redundant error logging during WhatsApp listener startup.

## [0.1.4] - 2026-04-04
### Changed
- Prevent the `parley update self` command from re-downloading the current release by printing the current and latest versions and returning early when they match.
- Clarify the Windows updater script by renaming the PID variable used to watch the parent process.

## [0.1.3] - 2026-04-04
### Changed
- Store WhatsApp sessions and history in `~/.bentos/parley/connect/whatsapp/` as JSON files, document the new paths, and drop the SQLite-backed database to simplify dependencies.
- Improve Windows workflows: fall back to buffered CLI prompts, base64-encode TTS text before handing it to PowerShell, keep the updater temp directory around until the Windows script finishes, and remove the PID-based guard so the desktop launcher simply starts the server and opens the browser.

## [0.1.2] - 2026-04-03
### Changed
- Quiet the dotenv loader and prevent the non-fatal load error from spamming startup logs.
- Add `-H=windowsgui` to the Windows build so the desktop release runs without an extra console window.
- Correct the Windows install instructions to reference the `parley` repository instead of `peer`.

## [0.1.1] - 2026-04-03
### Changed
- Replaced the embedded WebView desktop launcher with a PID-based toggle workflow that starts the headless server, opens the browser after a short delay, and stops the running process when the user double-clicks the binary again.
- Switched to the vendored `github.com/godeps/opus` fork that bundles static libogg/libopus/libopusfile binaries, eliminating the `webview_go` dependency and simplifying the desktop distribution.

## [0.1.0] - 2026-04-03
### Added
- Initial implementation of the `parley` CLI for launching structured AI debates, configuring LLM/TTS providers, and managing debate workflows.
- Documentation covering installation, configuration, WhatsApp connectivity, and the `parley serve` Web UI entry point.
- Desktop launcher that opens the Web UI, making the experience feel native on Linux, macOS, and Windows.
- Instructions for optional WhatsApp integration and setup guidance for LLM/TTS providers.
