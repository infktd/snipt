# Windows Release CI Design

**Date:** 2026-03-01
**Status:** Approved

## Problem

The release workflow builds for macOS (amd64/arm64) and Linux (amd64) but not Windows. Wails v3 supports Windows via `go-webview2`, which is already in `go.mod`.

## Decision

Add a `build-windows` job to `.github/workflows/release.yml` using a native `windows-latest` runner with MSYS2/MinGW-w64 for CGO support.

Cross-compilation was ruled out: the project already dropped Linux arm64 due to CGO cross-compilation issues, making native Windows the safer choice.

## Design

### New job: `build-windows`

- **Runner:** `windows-latest`
- **Architecture:** amd64 only
- **C compiler:** MSYS2 + MinGW-w64 via `msys2/setup-msys2@v2`
- **Build:** `go build` with `CGO_ENABLED=1 GOARCH=amd64`, output `snipt.exe`
- **Artifact:** `snipt-windows-amd64.zip` containing `snipt.exe`
- **Upload:** artifact name `windows-amd64`

### Release job changes

- Add `build-windows` to `needs: [build-macos, build-linux, build-windows]`
- No other changes — the existing flatten/checksum/release steps pick up the new zip automatically

## Artifact naming

Consistent with existing pattern: `snipt-{os}-{arch}.{ext}`

| Platform      | File                        |
|---------------|-----------------------------|
| macOS amd64   | `snipt-macos-amd64.zip`     |
| macOS arm64   | `snipt-macos-arm64.zip`     |
| Linux amd64   | `snipt-linux-amd64.tar.gz`  |
| Windows amd64 | `snipt-windows-amd64.zip`   |
