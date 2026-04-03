# Changelog

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
