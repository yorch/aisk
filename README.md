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
make build
./bin/aisk --version
```

## Quick Start

```bash
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

# Update all installations
aisk update

# Remove a skill
aisk uninstall 5-whys-skill --client claude
```

## Commands

### `aisk list [--remote] [--repo <owner/repo>] [--json]`

List available skills from the local repository. Use `--remote` to also fetch from GitHub (requires `--repo` or `AISK_REMOTE_REPO`).

```
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

### `aisk uninstall <skill> [--client <id>]`

Remove a skill. Without `--client`, removes from all clients where installed.

### `aisk status [--json]`

Show installed skills per client in a table view.

### `aisk update [skill] [--client <id>]`

Re-install skills with the latest version from the source repository.

### `aisk clients [--json]`

Show all detected AI clients with their install paths.

```
CLIENT           DETECTED  GLOBAL PATH                   PROJECT PATH
Claude Code      *         ~/.claude/skills              .claude/skills
Gemini CLI       *         ~/.gemini/GEMINI.md           GEMINI.md
Codex CLI        *         ~/.codex/instructions.md      AGENTS.md
VS Code Copilot  *         (n/a)                         .github/copilot-instructions.md
Cursor           *         (n/a)                         .cursor/rules
Windsurf         *         ~/.codeium/windsurf/...       .windsurf/rules
```

## How It Works

### Adapter System

Each client has a dedicated adapter that transforms skills into the native format:

| Client | Format | Method |
|--------|--------|--------|
| Claude Code | Directory with SKILL.md | Symlink (local) or copy (remote) |
| Gemini CLI | Markdown section in GEMINI.md | Append with section markers |
| Codex CLI | Markdown section in instructions.md | Append with section markers |
| VS Code Copilot | Markdown section | Append with section markers |
| Cursor | `.mdc` file with YAML frontmatter | Individual rule file |
| Windsurf | `.md` file or global rules section | File (project) or append (global) |

### Section Markers

For append-mode clients (Gemini, Codex, Copilot, Windsurf global), aisk uses HTML comment markers for idempotent installs:

```html
<!-- aisk:start:5-Whys Root Cause Analysis -->
...skill content...
<!-- aisk:end:5-Whys Root Cause Analysis -->
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

| Environment Variable | Purpose | Default |
|---------------------|---------|---------|
| `AISK_SKILLS_PATH` | Local skills repository path | Current working directory |
| `AISK_REMOTE_REPO` | Default GitHub repo for `--remote` | (none) |
| `GITHUB_TOKEN` | GitHub API authentication | (unauthenticated, 60 req/hr) |

Installation tracking is stored in `~/.aisk/manifest.json`.

## Development

```bash
make build     # Build binary to bin/aisk
make test      # Run all tests with race detector
make lint      # Run golangci-lint
make check     # Format, vet, and test
make fmt       # Run gofmt
make vet       # Run go vet
make snapshot  # GoReleaser snapshot build
make install   # Build and copy to /usr/local/bin
make clean     # Remove build artifacts
```

## License

[MIT](LICENSE)
