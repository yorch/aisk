# aisk Architecture

## Overview

`aisk` is a Go CLI/TUI tool that manages AI coding assistant skills across 6 clients. The architecture follows a clean layered design: CLI commands orchestrate skill discovery, client detection, format adaptation, and installation tracking.

```text
┌─────────────────────────────────────────────────────┐
│                   cmd/aisk/main.go                  │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                  internal/cli                        │
│   root · list · install · uninstall · status         │
│   update · clients · create · lint · audit           │
└──┬────┬────┬────┬────┬──────────────────────────────┘
   │    │    │    │    │
   ▼    ▼    ▼    ▼    ▼
 skill client adapter manifest tui
```

## Package Dependency Graph

```text
cmd/aisk/main.go
    └→ cli.Execute()

internal/cli
    ├→ config     (ResolvePaths, EnsureDirs, FindProjectRoot)
    ├→ skill      (ScanLocal, FetchRemoteList, ParseFrontmatter, Scaffold, LintSkillDir, CheckUpdates)
    ├→ client     (NewRegistry, DetectAll, ParseClientID)
    ├→ adapter    (ForClient, InstallOpts)
    ├→ manifest   (Load, Save, Lock, Add/Remove/Find/FindByScope)
    ├→ audit      (New logger, structured command/action events)
    ├→ gitignore  (EnsureEntries, RemoveEntries)
    └→ tui        (RunSkillSelect, RunClientSelect, PrintProgress, PrintStatusTable, PrintUpdateTable)

internal/adapter
    ├→ skill      (Skill type, ReadFullContent)
    └→ client     (ClientID constants)

internal/tui
    ├→ client     (Client type, AllClientIDs)
    ├→ skill      (Skill type)
    └→ manifest   (Manifest type — for BuildStatusEntries)

internal/manifest   (no internal deps)
internal/client     (no internal deps)
internal/skill      (no internal deps)
internal/config     (no internal deps)
internal/audit      (no internal deps)
internal/gitignore  (no internal deps)
```

**Key constraint**: `adapter` never imports `manifest`. The CLI loads the manifest after adapter operations complete, keeping adaptation and tracking cleanly separated.

## Packages

### `internal/config`

Resolves application paths from the user's home directory and environment.

| Export               | Type   | Purpose                                                  |
| -------------------- | ------ | -------------------------------------------------------- |
| `AppName`            | const  | `"aisk"`                                                 |
| `AppVersion`         | const  | CLI version string                                       |
| `Paths`              | struct | Home, AiskDir, CacheDir, ManifestDB, SkillsRepo          |
| `ResolvePaths()`     | func   | Resolves paths; `AISK_SKILLS_PATH` overrides SkillsRepo  |
| `Paths.EnsureDirs()` | method | Creates `~/.aisk/` and `~/.aisk/cache/`                  |
| `FindProjectRoot()`  | func   | Walks up from cwd to find root markers (`.git`, `go.mod`) |

### `internal/skill`

Skill model, YAML frontmatter parsing, local filesystem scanning, and GitHub remote fetching.

**Types:**

```go
type SkillSource int   // SourceLocal | SourceRemote

type Frontmatter struct {
    Name         string   `yaml:"name"`
    Description  string   `yaml:"description"`
    Version      string   `yaml:"version"`
    AllowedTools []string `yaml:"allowed-tools"`
}

type Skill struct {
    Frontmatter                        // embedded
    DirName        string              // directory name, e.g. "5-whys-skill"
    Path           string              // absolute path (local) or cache path (remote)
    Source         SkillSource
    MarkdownBody   string              // SKILL.md content after frontmatter
    ReferenceFiles []string            // relative paths
    ExampleFiles   []string
    AssetFiles     []string
}
```

**Functions:**

| Function                                                    | Purpose                                             |
| ----------------------------------------------------------- | --------------------------------------------------- |
| `ParseFrontmatter(content) → (Frontmatter, body, error)`    | Split `---` delimited YAML from markdown body       |
| `Skill.DisplayVersion() → string`                           | Returns version or `"unversioned"`                  |
| `ScanLocal(repoPath) → ([]*Skill, error)`                   | Scans subdirectories for SKILL.md files             |
| `ReadFullContent(skill, includeRefs) → (string, error)`     | Assembles body + optionally inlined reference files |
| `FetchRemoteList(owner, repo) → ([]*Skill, error)`          | Lists skills from a GitHub repo via API             |
| `FetchRemoteSkill(owner, repo, cacheDir) → (*Skill, error)` | Downloads full skill to local cache                 |
| `ParseRepoURL(url) → (owner, repo, ok)`                     | Parses `github.com/owner/repo` format               |
| `Scaffold(parentDir, name) → (string, error)`               | Creates skill skeleton (`SKILL.md`, `README.md`, dirs) |
| `LintSkillMD(content) → *LintReport`                        | Validates frontmatter/body and returns findings     |
| `LintSkillDir(path) → (*LintReport, error)`                 | Validates a full skill directory                    |
| `CheckUpdates(installed, available) → []UpdateInfo`         | Computes version mismatches for status/update hints |

**Local discovery logic:**

1. Read directory entries in `repoPath`
2. Skip hidden dirs and `node_modules`
3. For each subdirectory containing `SKILL.md`: parse frontmatter, discover `reference/` or `references/`, `examples/`, `assets/`

### `internal/client`

AI client detection and registry.

**Types:**

```go
type ClientID string  // "claude" | "gemini" | "codex" | "copilot" | "cursor" | "windsurf"

type Client struct {
    ID              ClientID
    Name            string    // "Claude Code"
    Detected        bool
    GlobalPath      string    // resolved global install path
    ProjectPath     string    // relative project install path
    SupportsGlobal  bool
    SupportsProject bool
}

type Registry struct { clients map[ClientID]*Client }
```

**Detection strategy** — two signals per client: config directory existence AND binary in PATH.

| Client          | Config Dir             | Binary     | Global Path                                    | Project Path                      |
| --------------- | ---------------------- | ---------- | ---------------------------------------------- | --------------------------------- |
| Claude Code     | `~/.claude/`           | `claude`   | `~/.claude/skills/`                            | `.claude/skills/`                 |
| Gemini CLI      | `~/.gemini/`           | `gemini`   | `~/.gemini/GEMINI.md`                          | `GEMINI.md`                       |
| Codex CLI       | `~/.codex/`            | `codex`    | `~/.codex/instructions.md`                     | `AGENTS.md`                       |
| VS Code Copilot | `~/.vscode/`           | `code`     | (none)                                         | `.github/copilot-instructions.md` |
| Cursor          | `~/.cursor/`           | `cursor`   | (none)                                         | `.cursor/rules/`                  |
| Windsurf        | `~/.codeium/windsurf/` | `windsurf` | `~/.codeium/windsurf/memories/global_rules.md` | `.windsurf/rules/`                |

**Scope support:**

- **Both global + project**: Claude, Gemini, Codex, Windsurf
- **Project only**: Copilot, Cursor

### `internal/adapter`

Format transformation layer. Each client receives skills in its native format.

**Interface:**

```go
type Adapter interface {
    Install(skill *Skill, targetPath string, opts InstallOpts) error
    Uninstall(skill *Skill, targetPath string) error
    Describe(skill *Skill, targetPath string, opts InstallOpts) string  // dry-run preview
}

type InstallOpts struct {
    Scope       string  // "global" or "project"
    IncludeRefs bool    // inline reference files
    DryRun      bool
}
```

**Factory**: `ForClient(id ClientID) → (Adapter, error)`

**Adapter implementations:**

| Adapter           | Clients                | Method                                                         | Idempotent                   |
| ----------------- | ---------------------- | -------------------------------------------------------------- | ---------------------------- |
| `ClaudeAdapter`   | Claude Code            | Symlink (local) or recursive copy (remote) to `skills/{name}/` | Remove + recreate            |
| `MarkdownAdapter` | Gemini, Codex, Copilot | Append consolidated markdown section to target file            | Yes — section markers        |
| `CursorAdapter`   | Cursor                 | Write `.mdc` file with Cursor YAML frontmatter                 | Overwrite                    |
| `WindsurfAdapter` | Windsurf               | Individual `.md` (project) or section-appended (global)        | Partial — markers for global |

**Section markers** (used by MarkdownAdapter and WindsurfAdapter global mode):

```html
<!-- aisk:start:skill-name -->
...consolidated skill content...
<!-- aisk:end:skill-name -->
```

Re-installing replaces content between markers without duplication. Uninstalling removes the entire section.

**Markdown consolidation algorithm:**

1. `# <skill-name>` header
2. Description as blockquote
3. SKILL.md markdown body
4. If `--include-refs`: inline reference files under `## Reference: <name>` headers
5. Wrap in section markers

**Cursor `.mdc` format:**

```yaml
---
description: <truncated first line of skill description>
globs:
alwaysApply: false
---
<SKILL.md markdown body>
```

### `internal/manifest`

Installation tracking persisted at `~/.aisk/manifest.json`.

**Types:**

```go
type Installation struct {
    SkillName    string    `json:"skill_name"`
    SkillVersion string    `json:"skill_version"`
    ClientID     string    `json:"client_id"`
    Scope        string    `json:"scope"`
    InstalledAt  time.Time `json:"installed_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    InstallPath  string    `json:"install_path"`
}

type Manifest struct {
    Installations []Installation `json:"installations"`
}
```

**Operations:**

| Method                         | Purpose                                                      |
| ------------------------------ | ------------------------------------------------------------ |
| `Load(path)`                   | Read JSON or return empty manifest                           |
| `Save()`                       | Write to JSON                                                |
| `Add(inst)`                    | Upsert — replaces existing entry for same skill+client+scope |
| `Remove(skill, client, scope)` | Delete single entry                                          |
| `RemoveAll(skill)`             | Delete all entries for a skill                               |
| `Find(skill, client)`          | Filter by skill name, optionally by client                   |
| `FindByClient(client)`         | All installations for a client                               |
| `FindByScope(scope)`           | All installations for a scope (`global`/`project`)           |
| `AllSkillNames()`              | Deduplicated list of installed skill names                   |

New project-scope installs store absolute install paths in manifest entries so cleanup logic can identify the current repository accurately.

### `internal/gitignore`

Manages a dedicated `# aisk managed` block in `.gitignore` for project-scope installs.

- `EnsureEntries(path, patterns)` adds missing patterns idempotently
- `RemoveEntries(path, patterns)` removes patterns and deletes empty managed block
- `GitignorePatternsForClient(clientID, installPath)` maps clients to expected project artifacts

### `internal/audit`

Structured audit logging for command and sub-action events.

- JSONL output at `~/.aisk/audit.log` by default (override: `AISK_AUDIT_LOG_PATH`)
- Enable/disable via `AISK_AUDIT_ENABLED` (`true` by default)
- Non-fatal best-effort writes (audit failures never fail commands)
- Built-in size-based rotation with numbered backups (`audit.log.1`, `.2`, ...), configurable via `AISK_AUDIT_MAX_SIZE_MB` and `AISK_AUDIT_MAX_BACKUPS`
- Sanitization/redaction runs before write to reduce risk of logging secrets from details/error fields

**Locking** (`Lock` type):

- File-based: creates `manifest.json.lock`
- `Acquire(timeout)`: retries every 100ms, recovers stale locks (>30s old)
- `Release()`: removes lock file

### `internal/tui`

Interactive Bubble Tea components with Lip Gloss styling.

**Components:**

| Component           | Model      | Purpose                          | Key bindings                                                                      |
| ------------------- | ---------- | -------------------------------- | --------------------------------------------------------------------------------- |
| `ClientSelectModel` | Bubble Tea | Multi-select client picker       | `↑↓` navigate, `space` toggle, `a` all, `n` none, `enter` confirm, `q`/`esc` quit |
| `SkillSelectModel`  | Bubble Tea | Filterable skill browser         | `↑↓` navigate, type to filter, `backspace` clear, `enter` select, `esc` quit      |
| `ProgressModel`     | Bubble Tea | Install/update progress with bar | Static output via `PrintProgress()`                                               |
| `StatusTable`       | tabwriter  | Cross-client status grid         | Non-interactive — `PrintStatusTable()`                                            |
| `UpdateTable`       | tabwriter  | Available update summary         | Non-interactive — `PrintUpdateTable()`                                            |

**Styling** (Lip Gloss):

- Color palette: Purple (titles), Cyan (selected), Green (success/checked), Yellow (active), Red (error), Gray (help text)
- Status indicators: `*` done, `o` active, `-` pending, `!` error

### `internal/cli`

Cobra command definitions. The CLI layer orchestrates all other packages.

**Commands:**

| Command     | Args      | Key Flags                                            | Interactive                                                |
| ----------- | --------- | ---------------------------------------------------- | ---------------------------------------------------------- |
| `list`      | (none)    | `--remote`, `--repo`, `--json`                       | No                                                         |
| `install`   | `[skill]` | `--client`, `--scope`, `--include-refs`, `--dry-run` | Yes — skill picker + client multi-select when args omitted |
| `uninstall` | `<skill>` | `--client`                                           | No                                                         |
| `status`    | (none)    | `--json`, `--check-updates`                          | No                                                         |
| `update`    | `[skill]` | `--client`                                           | No                                                         |
| `clients`   | (none)    | `--json`                                             | No                                                         |
| `create`    | `<name>`  | `--path`                                             | No                                                         |
| `lint`      | `[path]`  | (none)                                               | No                                                         |
| `audit`     | (none)    | `--limit`, `--run-id`, `--action`, `--status`, `--json`; subcommand: `prune` | No                                   |

**Install flow:**

```
Resolve skill (arg or TUI picker)
  → Detect all clients
    → Resolve target clients (--client flag or TUI multi-select)
      → Acquire manifest lock
        → For each client:
           → Resolve target path (global/project)
           → Get adapter (ForClient factory)
           → adapter.Install() or adapter.Describe() for dry-run
           → manifest.Add()
        → Save manifest
        → Print progress summary
```

## File Structure

```
aisk/
├── cmd/aisk/main.go                     # Entry point (15 lines)
├── internal/
│   ├── cli/                             # Cobra commands (~640 lines)
│   │   ├── root.go                      #   Root command, subcommand registration
│   │   ├── list.go                      #   aisk list
│   │   ├── install.go                   #   aisk install (TUI integration)
│   │   ├── uninstall.go                 #   aisk uninstall
│   │   ├── status.go                    #   aisk status
│   │   ├── update.go                    #   aisk update
│   │   ├── clients.go                   #   aisk clients
│   │   ├── create.go                    #   aisk create
│   │   ├── lint.go                      #   aisk lint
│   │   └── auditcmd.go                  #   aisk audit
│   ├── skill/                           # Skill model & discovery (~550 lines)
│   │   ├── skill.go                     #   Skill struct, frontmatter parsing
│   │   ├── local.go                     #   Local filesystem scanner
│   │   ├── remote.go                    #   GitHub API fetcher
│   │   ├── content.go                   #   Content reader (body + refs)
│   │   ├── scaffold.go                  #   Skill scaffolding
│   │   ├── validate.go                  #   Skill linting and name validation
│   │   └── updates.go                   #   Installed vs available version checks
│   ├── client/                          # AI client detection (~190 lines)
│   │   ├── client.go                    #   Client model + registry
│   │   └── detect.go                    #   Per-client detection logic
│   ├── adapter/                         # Format transformation (~410 lines)
│   │   ├── adapter.go                   #   Interface + factory
│   │   ├── claude.go                    #   Symlink/copy directory
│   │   ├── markdown.go                  #   Consolidated markdown (3 clients)
│   │   ├── cursor.go                    #   .mdc with YAML frontmatter
│   │   └── windsurf.go                  #   File (project) / append (global)
│   ├── manifest/                        # Installation tracking (~230 lines)
│   │   ├── manifest.go                  #   Read/write ~/.aisk/manifest.json
│   │   └── lockfile.go                  #   Concurrent access protection
│   ├── tui/                             # Bubble Tea components (~510 lines)
│   │   ├── styles.go                    #   Shared Lip Gloss styles
│   │   ├── clientselect.go              #   Multi-select client picker
│   │   ├── skillselect.go              #   Skill browser with filtering
│   │   ├── progress.go                  #   Install/update progress view
│   │   ├── statustable.go              #   Status table view
│   │   └── updatetable.go              #   Updates table view
│   ├── gitignore/
│   │   └── gitignore.go                #   Managed .gitignore section helpers
│   ├── audit/
│   │   └── audit.go                    #   Structured JSONL action logging
│   └── config/
│       ├── config.go                    #   Paths, defaults, env vars
│       └── projectroot.go               #   Project root detection
├── internal/**/*_test.go                # Tests (~900 lines across 8 files)
├── go.mod / go.sum
├── justfile
├── .goreleaser.yaml
└── README.md
```

## External Dependencies

| Package                              | Version | Purpose                  |
| ------------------------------------ | ------- | ------------------------ |
| `github.com/spf13/cobra`             | v1.10.2 | CLI command framework    |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | TUI framework            |
| `github.com/charmbracelet/bubbles`   | v0.21.1 | Pre-built TUI widgets    |
| `github.com/charmbracelet/lipgloss`  | v1.1.0  | Terminal styling         |
| `gopkg.in/yaml.v3`                   | v3.0.1  | YAML frontmatter parsing |

No `go-github` dependency — the remote fetcher uses `net/http` with the GitHub REST API directly, keeping the dependency tree minimal.

## Environment Variables

| Variable           | Purpose                            | Default                   |
| ------------------ | ---------------------------------- | ------------------------- |
| `AISK_SKILLS_PATH` | Local skills repository path       | Current working directory |
| `AISK_REMOTE_REPO` | Default GitHub repo for `--remote` | (none)                    |
| `GITHUB_TOKEN`     | GitHub API auth (60 → 5000 req/hr) | Unauthenticated           |

## Data Flow

### Skill Discovery

```text
Local:  AISK_SKILLS_PATH → ScanLocal() → []*Skill
Remote: GitHub API → FetchRemoteList() → []*Skill (metadata only)
                   → FetchRemoteSkill() → *Skill (full download to cache)
```

### Installation

```text
Skill + Client + Scope
  → resolveTargetPath(client, scope)
  → adapter.ForClient(clientID)
  → adapter.Install(skill, targetPath, opts)
  → manifest.Add(installation)
  → (project scope) update managed .gitignore section
  → manifest.Save()
```

### Section Marker Lifecycle

```text
First install:   create file with <!-- aisk:start:X -->...<!-- aisk:end:X -->
Re-install:      find markers → replace content between them
Uninstall:       find markers → remove section + surrounding whitespace
```
