# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o mcrgt.exe .        # Build for Windows
go run . <command> [flags]     # Run directly
go build .                     # Build for current platform
```

No test suite exists yet. No lint/formatter is configured.

## Architecture

This is a Go CLI tool (`module mcr_gradletools`) that fixes MCreator's Gradle download failures. It exploits the fact that MCreator stores `.lck`/`.part` files in `~/.mcreator/gradle/wrapper/dists/` after a failed download, from which version/edition info can be extracted to then download from Chinese mirrors instead.

**Entry point:** `main.go` — defines the CLI app using `urfave/cli/v2` with five subcommands:

| Command | Purpose |
|---|---|
| `gradle` | Core workflow: scan MCreator's Gradle dir for `.lck`/`.part` files → delete them → download the correct Gradle zip from a mirror → copy it into place |
| `download` | Download a specific Gradle version+edition (bin/all) from mirrors into the local cache (`~/.mcrgradletool/cache/`) |
| `check-mirrors` | Check all three Chinese Gradle mirrors (Tencent, Huawei, Tsinghua) for availability |
| `clear-cache` | List or delete cached Gradle zip files |
| `version` | Print version info |

**Library (`lib/`):**
- `downloadGradle.go` — Mirror list (Tencent/Huawei/Tsinghua × bin/all = 6 URLs), download logic with progress bars, cache directory management. `DownloadGradle(version, edition)` is the main entry point that checks mirrors in order.
- `mcreator_gradle.go` — `ProcessMCreatorGradle(gradlePath)` is the core: scans for `.lck`/`.part` files, extracts `gradle-X.Y.Z-{bin,all}.zip` pattern via regex, deletes temp files, then calls `DownloadGradle` + copies into MCreator's `dists/` subdirectory.
- `file.go` — Generic file utilities: `FileFind` (regex-based recursive search), `FileCopy`, `FileDelete`. Note: `FileCopy` and `FileDelete` return an error if the target doesn't exist rather than treating it as success.

**Key external dependencies:**
- `github.com/urfave/cli/v2` — CLI framework
- `github.com/schollz/progressbar/v3` — Download/copy progress bars (output to stderr)

## Important Details

- The default Gradle path is `~/.mcreator/gradle/wrapper/dists/` (hardcoded in `main.go` via `os/user.Current()`).
- Cache goes to `~/.mcrgradletool/cache/` (not the working directory).
- Version matching regex: `gradle-(\d+\.\d+\.\d+)-(bin|all)\.zip` — only three-component versions (e.g., `8.14.2`) are recognized.
- The CI workflow (`.github/workflows/sync.yml`) mirrors from Gitee to GitHub via scheduled cron + manual dispatch. It uses `PAT_TOKEN` secret for push.
- Go version is 1.25.1 (unusually new — the project tracks the latest Go release).
- Windows is the primary target platform (only Windows installers are built, via `InnoSetupScript.iss`).
