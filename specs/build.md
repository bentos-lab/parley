# Build Workflow

This repository ships a single `build/build.sh` script that produces both the optimized web UI bundle and the multi-platform Go binaries with the embedded SPA assets. The script is the authoritative way to regenerate `adapter/inbound/rest/dist` and `dist/` in one pass.

## Prerequisites

- `pnpm` (the frontend uses the pnpm workspace defined in `webui/`).
- `Go 1.22+` with CGO disabled for static binaries.
- `git` for version/commit metadata.

## Process

1. **Compile the web UI.**
   - Runs `pnpm install --frozen-lockfile` and `pnpm run build` inside `webui/` to recreate `webui/dist`.
2. **Embed the SPA.**
   - Empties and replaces `adapter/inbound/rest/dist` with the new contents of `webui/dist` so the REST server always ships the current SPA.
3. **Cross-compile the Go CLI/server.**
- Builds `./cmd/parley` for each OS/ARCH in the script, applies `-ldflags` metadata (`version`, `commit`), and archives the executables into `dist/` (`.tar.gz` for Unix platforms, `.zip` for Windows).

## Running the build

Execute the script from the repository root:

```bash
./build/build.sh
```

The script prints progress for each stage and exits non-zero if any prerequisite or build command fails.

## Artifacts

- `adapter/inbound/rest/dist`: must contain the SPA assets that get embedded into the REST server.
- `dist/*.tar.gz` and `dist/*.zip`: platform-specific bundles of the `parley` binary.

See `build/build.sh` for the exact platform list, entrypoint, and metadata flags.

## UI-only build

- `./build/buildui.sh` installs the frontend dependencies, runs `pnpm run build` inside `webui/`, and mirrors the generated `webui/dist` contents into `adapter/inbound/rest/dist` without rebuilding any Go artifacts.
- Run this helper whenever you only need to refresh the embedded SPA or work on frontend assets that are consumed by the REST server.
- Prerequisite: `pnpm` must already be available on your PATH so that the script can install and build the web UI.

## UI-only build

- `./build/buildui.sh` installs the frontend dependencies, runs `pnpm run build` inside `webui/`, and mirrors the generated `webui/dist` contents into `adapter/inbound/rest/dist` without rebuilding any Go artifacts.
- Run this helper whenever you only need to refresh the embedded SPA or work on frontend assets that are consumed by the REST server.
- Prerequisite: `pnpm` must already be available on your PATH so that the script can install and build the web UI.
