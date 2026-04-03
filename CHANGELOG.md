# Changelog

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
