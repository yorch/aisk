# aisk Planning

## Problem Statement

AI coding assistants (Claude Code, Gemini CLI, Codex CLI, VS Code Copilot, Cursor, Windsurf) each have their own format for loading custom skills/rules/instructions. Managing skills across multiple clients requires manual copying, format conversion, and tracking — error-prone and tedious.

## Goals

1. **Single command** to install a skill to any combination of AI clients
2. **Native format** for each client — not a lowest-common-denominator approach
3. **Idempotent** operations — re-installing updates in place, never duplicates
4. **Interactive when convenient, scriptable when needed** — TUI pickers when args are omitted, flags for automation
5. **Track installations** — know what's installed where, detect version mismatches

## Design Decisions

### Go + Cobra + Bubble Tea

**Chosen**: Go with Cobra (CLI) and Bubble Tea (TUI)

**Rationale**:
- Single static binary — no runtime dependencies, easy distribution
- Cobra gives us subcommands, flags, and help generation for free
- Bubble Tea enables rich terminal UIs (multi-select, filtering, progress bars)
- Lip Gloss provides clean terminal styling without ANSI escape code management
- Cross-platform (darwin/linux/windows) via GoReleaser

**Alternatives considered**:
- Python + Click: would require Python runtime; TUI options (textual, rich) add weight
- Rust + clap: stronger type safety but slower development velocity for a tool this size
- Node + ink: would require Node.js runtime

### Adapter Pattern over Template System

**Chosen**: Per-client adapter structs implementing a common interface

**Rationale**:
- Each client's format is different enough that a template system would still need per-client logic
- Adapters encapsulate format-specific knowledge (symlinks vs append vs .mdc)
- Easy to add new clients — implement the 3-method interface
- `MarkdownAdapter` reused for 3 clients proves the abstraction works

**Alternative considered**: Single function with a switch statement — rejected because uninstall and describe logic also differ per client.

### Symlinks for Claude Code

**Chosen**: Symlink local skills, copy remote skills

**Rationale**:
- Symlinks keep skills in sync with the source repo automatically
- No duplication — changes in the skills repo are immediately visible
- Remote skills can't be symlinked (no local path), so they get copied

### Section Markers for Append-Mode Clients

**Chosen**: HTML comment markers `<!-- aisk:start:name -->...<!-- aisk:end:name -->`

**Rationale**:
- Invisible in rendered markdown
- Enables idempotent updates — find markers, replace content between them
- Enables clean uninstall — remove the section without corrupting surrounding content
- Human-readable for debugging

**Alternative considered**: Separate files per skill for all clients — rejected because Gemini/Codex/Copilot expect a single file.

### Manifest Tracking

**Chosen**: JSON file at `~/.aisk/manifest.json` with file-based locking

**Rationale**:
- JSON is human-readable and debuggable
- Simple enough — no need for SQLite or other databases
- File-based lock prevents concurrent aisk processes from corrupting the manifest
- Stale lock recovery (30s timeout) handles crashed processes

### Skill Discovery via SKILL.md

**Chosen**: Scan for `SKILL.md` files with YAML frontmatter in subdirectories

**Rationale**:
- Matches the existing claude-skills repo convention
- YAML frontmatter is standard in static site generators, familiar to developers
- Progressive disclosure — frontmatter is small, reference files are loaded on demand
- `reference/` vs `references/` both supported to handle existing inconsistency

### Interactive TUI as Default, Flags as Override

**Chosen**: Launch TUI pickers when skill/client args are omitted

**Rationale**:
- `aisk install` with no args → interactive skill browser → client multi-select → install
- `aisk install 5-whys-skill --client claude` → direct install, no TUI
- Best of both worlds: discoverable for new users, scriptable for automation
- Follows the pattern of tools like `gh`, `fzf`, `gum`

## Client Format Matrix

| Client | Install Format | Uninstall Method | Scope Support |
|--------|---------------|------------------|---------------|
| Claude Code | Symlinked directory in `~/.claude/skills/` | Remove symlink/dir | Global + Project |
| Gemini CLI | Markdown section appended to `GEMINI.md` | Remove section by markers | Global + Project |
| Codex CLI | Markdown section appended to `instructions.md` | Remove section by markers | Global + Project |
| VS Code Copilot | Markdown section appended to `copilot-instructions.md` | Remove section by markers | Project only |
| Cursor | `.mdc` file in `.cursor/rules/` | Delete file | Project only |
| Windsurf | `.md` file (project) or section append (global) | Delete file or remove section | Global + Project |

## Phased Delivery

### Phase 1: Foundation (MVP) — Complete

- [x] Go module setup with Cobra
- [x] Skill struct, frontmatter parser (`ParseFrontmatter`)
- [x] Local filesystem scanner (`ScanLocal`)
- [x] Config (paths, env vars)
- [x] Manifest read/write with file locking
- [x] Claude Code client detection + adapter (symlink/copy)
- [x] CLI commands: `list`, `install`, `status` (basic text output)
- [x] Unit tests for parsing, scanner, Claude adapter, manifest

### Phase 2: All Client Adapters — Complete

- [x] All 6 client detectors (Claude, Gemini, Codex, Copilot, Cursor, Windsurf)
- [x] `MarkdownAdapter` — consolidated markdown for Gemini, Codex, Copilot
- [x] `CursorAdapter` — `.mdc` files with YAML frontmatter
- [x] `WindsurfAdapter` — file (project) + section append (global)
- [x] `uninstall`, `update`, `clients` commands
- [x] `--dry-run`, `--include-refs`, `--scope` flags
- [x] Integration tests for each adapter

### Phase 3: TUI Layer — Complete

- [x] Bubble Tea client multi-select picker
- [x] Bubble Tea skill browser with filtering
- [x] Status table view
- [x] Progress display with bar
- [x] TUI integration in `install` command (auto-launches when args omitted)

### Phase 4: Remote Skills & Polish — Complete

- [x] GitHub API skill listing (`FetchRemoteList`)
- [x] GitHub API skill download (`FetchRemoteSkill`)
- [x] `--remote` and `--repo` flags on `list`
- [x] `--json` output for `list`, `status`, `clients`
- [x] GoReleaser config (darwin/linux/windows × amd64/arm64)
- [x] Makefile (build, test, lint, check, snapshot, install, clean)
- [x] README with usage documentation

### Phase 5: Future (post-MVP)

- [ ] `aisk create <name>` — scaffold a new skill directory
- [ ] `aisk lint <path>` — validate SKILL.md format and structure
- [ ] Project-scoped installs with automatic `.gitignore` management
- [ ] Auto-update notifications when installed skills have newer versions
- [ ] Homebrew tap publishing
- [ ] Shell completions (bash, zsh, fish)

## Risk Analysis

| Risk | Mitigation |
|------|-----------|
| Client format changes | Adapter pattern isolates changes to one file per client |
| GitHub API rate limits | `GITHUB_TOKEN` support; unauthenticated still allows 60 req/hr |
| Large reference files | `--include-refs` is opt-in; default installs are lean (SKILL.md body only) |
| Concurrent aisk processes | File-based manifest locking with stale lock recovery |
| New AI clients emerge | Adding a client = new detector function + adapter struct + register in factory |
| Symlink not supported (Windows) | `ClaudeAdapter` falls back to copy for remote skills; local symlinks work on Windows 10+ with developer mode |
