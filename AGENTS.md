# Project Guidelines for AI Agents

This file provides guidance to AI Agents when working with code in this repository.

## What This Is

`aisk` is a Go CLI tool that installs, manages, and updates AI coding assistant skills across 6 clients: Claude Code, Gemini CLI, Codex CLI, VS Code Copilot, Cursor, and Windsurf. Each client receives skills in its native format via a dedicated adapter.

## Build & Test

```bash
just build          # → bin/aisk
just test           # go test ./... -count=1 -race
just lint           # golangci-lint run ./...
just check          # fmt + vet + test
just snapshot       # goreleaser snapshot build
```

Run a single test:
```bash
go test ./internal/skill/ -run TestParseFrontmatter -count=1 -v
```

The binary is built from `cmd/aisk/main.go`. Version is in `internal/config/config.go` (`AppVersion`).

## Architecture

The data flow is: **CLI command → skill discovery → client detection → adapter transforms → manifest tracks**.

### Package Responsibilities

- **`cmd/aisk`** — Entrypoint, delegates to `cli.Execute()`
- **`internal/cli`** — Cobra commands: `install`, `uninstall`, `list`, `update`, `status`, `clients`, `create`, `lint`. Orchestrates the full install pipeline (discover skills → detect clients → pick adapter → install → update manifest).
- **`internal/skill`** — Skill model and discovery. `ScanLocal()` finds `SKILL.md` files in subdirectories and parses YAML frontmatter. `FetchRemoteList()`/`FetchRemoteSkill()` fetch from GitHub API. `ReadFullContent()` optionally inlines reference files.
- **`internal/client`** — Client registry and detection. `DetectAll()` checks for config dirs (`~/.claude`, `~/.gemini`, etc.) and binaries in PATH. Each client knows its global/project install paths.
- **`internal/adapter`** — The `Adapter` interface (`Install`, `Uninstall`, `Describe`) with per-client implementations:
  - `ClaudeAdapter` — symlinks local skills, copies remote ones to `~/.claude/skills/`
  - `MarkdownAdapter` — shared by Gemini/Codex/Copilot; appends markdown sections with `<!-- aisk:start/end -->` markers for idempotent updates
  - `CursorAdapter` — writes `.mdc` files with Cursor YAML frontmatter
  - `WindsurfAdapter` — individual `.md` files (project) or appended sections (global)
- **`internal/manifest`** — JSON file at `~/.aisk/manifest.json` tracking all installations (skill, client, scope, timestamps, paths). Uses file-based locking with stale-lock recovery.
- **`internal/config`** — Path resolution (`~/.aisk/`, cache dir, skills repo from `AISK_SKILLS_PATH` or cwd) and project-root detection for project-scope behaviors.
- **`internal/audit`** — Structured JSONL audit logger for command/action events (`~/.aisk/audit.log` by default).
- **`internal/gitignore`** — Managed `.gitignore` section helpers for project-scope installs/uninstalls.
- **`internal/tui`** — Bubble Tea interactive components (skill picker, client multi-select, progress display)

### Key Design Patterns

- **Adapter pattern**: `adapter.ForClient(id)` returns the right adapter. All adapters implement the same 3-method interface. The `MarkdownAdapter` is reused across Gemini, Codex, and Copilot with only a name difference.
- **Section markers**: Append-mode adapters use HTML comment markers (`<!-- aisk:start:name -->`) for idempotent installs — re-installing replaces in-place.
- **Skill frontmatter**: Every skill has a `SKILL.md` with YAML frontmatter (`name`, `description`, `version`, `allowed-tools`). `ParseFrontmatter()` splits it from the markdown body.
- **Scope duality**: Install supports `--scope global|project`. Clients declare which scopes they support via `SupportsGlobal`/`SupportsProject` bools.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `AISK_SKILLS_PATH` | Local skills repo path (default: cwd) |
| `AISK_REMOTE_REPO` | Default GitHub repo for `--remote` |
| `GITHUB_TOKEN` | GitHub API auth for remote skill fetching |

## Dependencies

Go 1.25+, Cobra (CLI), Bubble Tea + Lipgloss (TUI), yaml.v3 (frontmatter parsing). Releases via GoReleaser.
