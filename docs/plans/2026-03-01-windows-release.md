# Windows Release CI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a `build-windows` job to the release workflow that builds `snipt.exe` and uploads it as `snipt-windows-amd64.zip`.

**Architecture:** Native `windows-latest` GitHub Actions runner with MSYS2/MinGW-w64 providing `gcc` for CGO. Mirrors the existing macOS/Linux job structure. The `release` job already handles picking up artifacts, so only the `needs:` list and the new job need changing.

**Tech Stack:** GitHub Actions, `msys2/setup-msys2@v2`, Go CGO, Wails v3 (`go-webview2`)

---

### Task 1: Add `build-windows` job and wire it into `release`

**Files:**
- Modify: `.github/workflows/release.yml`

This is a pure YAML edit — no code, no tests. Verification is visual inspection of the diff.

**Step 1: Add the `build-windows` job**

In `.github/workflows/release.yml`, add the following job after `build-linux` and before `release`:

```yaml
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: src/frontend/package-lock.json

      - name: Install frontend dependencies
        run: cd src/frontend && npm ci

      - name: Build frontend
        run: cd src/frontend && npm run build

      - name: Set up MSYS2
        uses: msys2/setup-msys2@v2
        with:
          msystem: MINGW64
          update: true
          install: mingw-w64-x86_64-gcc

      - name: Build Windows binary
        env:
          CGO_ENABLED: '1'
          GOARCH: amd64
          CC: C:\msys64\mingw64\bin\gcc.exe
        shell: bash
        run: |
          VERSION="${GITHUB_REF_NAME#v}"
          go build -ldflags "-s -w -X main.version=$VERSION" \
            -o snipt.exe ./src/cmd/snipt

      - name: Create zip
        shell: bash
        run: zip snipt-windows-amd64.zip snipt.exe

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: windows-amd64
          path: snipt-windows-amd64.zip
```

**Step 2: Update `release` job's `needs` list**

Find this line in the `release` job:

```yaml
    needs: [build-macos, build-linux]
```

Change it to:

```yaml
    needs: [build-macos, build-linux, build-windows]
```

**Step 3: Verify the diff looks correct**

```bash
git diff .github/workflows/release.yml
```

Expected: two hunks — one adding the `build-windows` job, one updating `needs:`.

**Step 4: Commit**

```bash
git add .github/workflows/release.yml docs/plans/2026-03-01-windows-release-design.md docs/plans/2026-03-01-windows-release.md
git commit -m "feat(ci): add Windows amd64 release build"
```

---

## Notes

- `msys2/setup-msys2@v2` installs to `C:\msys64` by default. The `CC` env var points directly at MinGW gcc so Go's CGO picks it up without shell PATH gymnastics.
- `shell: bash` on Windows runners uses Git Bash, which handles the `${GITHUB_REF_NAME#v}` substring syntax fine.
- The `zip` command is available in Git Bash on Windows runners.
- The `release` job's flatten step (`find dist/ -type f \( -name '*.zip' -o -name '*.tar.gz' \)`) already matches `.zip`, so `snipt-windows-amd64.zip` is picked up with no further changes.
