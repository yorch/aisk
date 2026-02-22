# aisk — AI Skill Manager

Install, manage, and update AI coding assistant skills across multiple clients.

**Supported clients:** Claude Code, Gemini CLI, Codex CLI, VS Code Copilot, Cursor, Windsurf

Each client receives skills in its **native format** — symlinks for Claude, consolidated markdown for Gemini/Codex/Copilot, `.mdc` files for Cursor, and individual rules for Windsurf.

## Install

```bash
# From source
go install github.com/yorch/aisk/cmd/aisk@latest

# Or build locally
git clone https://github.com/yorch/aisk.git
cd aisk
just build
./bin/aisk --version
```

## Quick Start

```bash
# Scaffold a new skill
aisk create my-new-skill

# Validate it before publishing
aisk lint my-new-skill

# List available skills (run from your skills repo directory)
aisk list

# See which AI clients are detected on your system
aisk clients

# Install a skill to Claude Code
aisk install 5-whys-skill --client claude

# Install to multiple clients (interactive TUI picker)
aisk install 5-whys-skill

# Check what's installed
aisk status

# Skip update checks in status output
aisk status --check-updates=false

# Inspect recent audit events
aisk audit --limit 20

# Update all installations
aisk update

# Remove a skill
aisk uninstall 5-whys-skill --client claude
```

## Commands

### `aisk list [--remote] [--repo <owner/repo>] [--json]`

List available skills from the local repository. Use `--remote` to also fetch from GitHub (requires `--repo` or `AISK_REMOTE_REPO`).

```text
NAME                        VERSION      DIRECTORY               SOURCE
5-Whys Root Cause Analysis  0.1.0        5-whys-skill            local
code-review-excellence      unversioned  code-review-skill       local
First Principles Thinking   0.2.0        first-principles-skill  local
```

### `aisk install [skill] [--client <id>] [--scope global|project] [--include-refs] [--dry-run]`

Install a skill to one or more AI clients.

- **No skill argument**: launches interactive skill browser
- **No --client flag**: launches interactive multi-select client picker
- `--include-refs`: inline reference files (can be large for some skills)
- `--dry-run`: preview changes without writing

When `--scope project` is used, aisk manages a dedicated section in the project `.gitignore`:

- Adds client-specific install artifacts on install (for successful installs only)
- Removes entries on uninstall when that client no longer has project installs in the current repo

### `aisk uninstall <skill> [--client <id>]`

Remove a skill. Without `--client`, removes from all clients where installed.

### `aisk status [--json] [--check-updates=true|false]`

Show installed skills per client in a table view.

- `--check-updates` defaults to `true`
- When enabled, prints an "Updates available" table based on local repository versions

### `aisk update [skill] [--client <id>]`

Re-install skills with the latest version from the source repository.

### `aisk clients [--json]`

Show all detected AI clients with their install paths.

```text
CLIENT           DETECTED  GLOBAL PATH                   PROJECT PATH
Claude Code      *         ~/.claude/skills              .claude/skills
Gemini CLI       *         ~/.gemini/GEMINI.md           GEMINI.md
Codex CLI        *         ~/.codex/instructions.md      AGENTS.md
VS Code Copilot  *         (n/a)                         .github/copilot-instructions.md
Cursor           *         (n/a)                         .cursor/rules
Windsurf         *         ~/.codeium/windsurf/...       .windsurf/rules
```

### `aisk audit [--limit N] [--run-id <id>] [--action <name>] [--status <value>] [--json]`

Inspect audit events from the local audit log.

- `--limit` defaults to `50` (`0` = all)
- `--run-id` filters to a single CLI invocation
- `--action` filters by action key (e.g. `install.adapter.apply`)
- `--status` filters by event status (`started`, `success`, `error`, `skipped`)
- `--json` outputs raw event objects

### `aisk audit prune [--keep-days N] [--keep N] [--dry-run]`

Prune and compact audit logs (including rotated backups).

- `--keep-days` defaults to `30` (`0` disables age filtering)
- `--keep` defaults to `2000` (`0` disables count filtering)
- `--dry-run` previews removals without writing

### `aisk create <name> [--path <dir>]`

Scaffold a new skill directory with:

- `SKILL.md` template with frontmatter + instruction placeholders
- `README.md`
- `reference/`
- `examples/`

`name` must be kebab-case: lowercase letters, digits, and single hyphens.

### `aisk lint [path]`

Validate a skill directory or a single `SKILL.md` file.

- Reports errors and warnings
- Exits with code `1` when errors are present
- Checks frontmatter validity, required fields, body content, version-format warnings, and empty `reference/`/`examples/`

## How It Works

### Adapter System

Each client has a dedicated adapter that transforms skills into the native format:

| Client          | Format                              | Method                            |
| --------------- | ----------------------------------- | --------------------------------- |
| Claude Code     | Directory with SKILL.md             | Symlink (local) or copy (remote)  |
| Gemini CLI      | Markdown section in GEMINI.md       | Append with section markers       |
| Codex CLI       | Markdown section in instructions.md | Append with section markers       |
| VS Code Copilot | Markdown section                    | Append with section markers       |
| Cursor          | `.mdc` file with YAML frontmatter   | Individual rule file              |
| Windsurf        | `.md` file or global rules section  | File (project) or append (global) |

### Section Markers

For append-mode clients (Gemini, Codex, Copilot, Windsurf global), aisk uses HTML comment markers for idempotent installs:

```html
<!-- aisk:start:5-whys-skill -->
...skill content...
<!-- aisk:end:5-whys-skill -->
```

This means re-installing updates the content in-place without duplication.

### Skill Discovery

Skills are discovered by scanning directories for `SKILL.md` files with YAML frontmatter:

```yaml
---
name: my-skill
description: What this skill does
version: 1.0.0
---
```

Set `AISK_SKILLS_PATH` to point to your skills repository, or run `aisk` from the repo directory.

## Configuration

| Environment Variable | Purpose                            | Default                      |
| -------------------- | ---------------------------------- | ---------------------------- |
| `AISK_SKILLS_PATH`   | Local skills repository path       | Current working directory    |
| `AISK_REMOTE_REPO`   | Default GitHub repo for `--remote` | (none)                       |
| `GITHUB_TOKEN`       | GitHub API authentication          | (unauthenticated, 60 req/hr) |
| `AISK_AUDIT_ENABLED` | Enable/disable audit logging       | `true`                       |
| `AISK_AUDIT_LOG_PATH` | Audit log file path (JSONL)       | `~/.aisk/audit.log`          |
| `AISK_AUDIT_MAX_SIZE_MB` | Max audit log size before rotation | `5`                     |
| `AISK_AUDIT_MAX_BACKUPS` | Number of rotated backups (`.1`, `.2`, ...) | `3`         |

Installation tracking is stored in `~/.aisk/manifest.json`.

Audit logs are written as JSON Lines (`.jsonl`-style) with one event per line, including command/action/status and contextual fields (skill, client, scope, target path, details, error).
Sensitive values in audit payloads are sanitized before write (for example token/secret/password fields and inline bearer/key-value secrets).

## Development

```bash
just build     # Build binary to bin/aisk
just test      # Run all tests with race detector
just lint      # Run golangci-lint
just check     # Format, vet, and test
just fmt       # Run gofmt
just vet       # Run go vet
just snapshot  # GoReleaser snapshot build
just install   # Build and copy to /usr/local/bin
just clean     # Remove build artifacts
```

## License

[MIT](LICENSE)
