# aisk Task Breakdown

## Summary

| Phase                        | Tasks  | Status                 |
| ---------------------------- | ------ | ---------------------- |
| Phase 1: Foundation          | 12     | Complete               |
| Phase 2: All Client Adapters | 10     | Complete               |
| Phase 3: TUI Layer           | 6      | Complete               |
| Phase 4: Remote & Polish     | 9      | Complete               |
| Phase 5: Future              | 6      | Backlog                |
| **Total**                    | **43** | **37 done, 6 backlog** |

---

## Phase 1: Foundation (MVP)

### 1.1 Project scaffold

- [x] Create `aisk/` directory structure (`cmd/`, `internal/` with 7 packages)
- [x] Initialize Go module (`github.com/yorch/aisk`)
- [x] Write `cmd/aisk/main.go` entry point

### 1.2 Config package

- [x] Implement `Paths` struct with Home, AiskDir, CacheDir, ManifestDB, SkillsRepo
- [x] `ResolvePaths()` — resolve from home dir + `AISK_SKILLS_PATH` env var
- [x] `EnsureDirs()` — create `~/.aisk/` and `~/.aisk/cache/`

### 1.3 Skill model and parsing

- [x] Define `Frontmatter` struct (name, description, version, allowed-tools)
- [x] Define `Skill` struct (frontmatter + DirName, Path, Source, MarkdownBody, file lists)
- [x] Implement `ParseFrontmatter()` — split `---` delimited YAML from markdown body
- [x] Handle edge cases: missing version, multi-line description, missing delimiters

### 1.4 Local skill scanner

- [x] Implement `ScanLocal()` — discover SKILL.md in subdirectories
- [x] Skip hidden dirs and `node_modules`
- [x] Discover reference files in both `reference/` and `references/` (handle inconsistency)
- [x] Discover `examples/` and `assets/` subdirectories

### 1.5 Manifest system

- [x] Define `Installation` struct (skill, version, client, scope, timestamps, path)
- [x] Implement `Load()` / `Save()` — JSON file at `~/.aisk/manifest.json`
- [x] Implement `Add()` with upsert semantics (replace existing skill+client+scope)
- [x] Implement `Remove()`, `RemoveAll()`, `Find()`, `FindByClient()`, `AllSkillNames()`
- [x] File-based locking with stale-lock recovery (30s timeout)

### 1.6 Client detection (Claude)

- [x] Define `Client` struct and `Registry`
- [x] Implement `DetectAll()` framework
- [x] Claude Code detector: check `~/.claude/` dir + `claude` binary in PATH

### 1.7 Claude adapter

- [x] Implement `Adapter` interface: `Install`, `Uninstall`, `Describe`
- [x] Symlink for local skills
- [x] Recursive copy for remote skills
- [x] Adapter factory: `ForClient(clientID)`

### 1.8 CLI commands (list, install, status)

- [x] Cobra root command with version flag
- [x] `list` command with tabwriter output
- [x] `install` command with `--client`, `--scope`, `--dry-run` flags
- [x] `status` command showing installed skills

### 1.9 Unit tests

- [x] Frontmatter parsing: valid, missing version, multi-line, missing delimiters
- [x] Local scanner: discovery, reference/references dirs, hidden dir skip
- [x] Claude adapter: symlink, copy, uninstall
- [x] Manifest: add, replace, save/load cycle, remove, AllSkillNames
- [x] Client detection: with/without config dirs, ParseClientID

**Verification**: `go build ./cmd/aisk && go test ./... -race` — all passing

---

## Phase 2: All Client Adapters

### 2.1 Client detectors

- [x] Gemini CLI: `~/.gemini/` + `gemini` binary
- [x] Codex CLI: `~/.codex/` + `codex` binary
- [x] VS Code Copilot: `~/.vscode/` + `code` binary
- [x] Cursor: `~/.cursor/` + `cursor` binary
- [x] Windsurf: `~/.codeium/windsurf/` + `windsurf` binary

### 2.2 Markdown consolidated adapter

- [x] Build markdown section: header + description blockquote + body
- [x] Section markers: `<!-- aisk:start:name -->` / `<!-- aisk:end:name -->`
- [x] Append to existing file, create if missing
- [x] Replace existing section (idempotent re-install)
- [x] Remove section (uninstall)
- [x] Optional reference file inlining (`--include-refs`)
- [x] Register for Gemini, Codex, Copilot in adapter factory

### 2.3 Cursor adapter

- [x] Generate `.mdc` format with Cursor YAML frontmatter
- [x] Truncate description for frontmatter (first line, max 200 chars)
- [x] Write to target rules directory
- [x] Delete file on uninstall

### 2.4 Windsurf adapter

- [x] Project-level: write individual `.md` file
- [x] Global-level: append to `global_rules.md` using section markers
- [x] Uninstall: delete file (project) or remove section (global)

### 2.5 Remaining CLI commands

- [x] `uninstall` command: remove from specific client or all
- [x] `update` command: re-install with latest version, update manifest timestamps
- [x] `clients` command: show all 6 clients with detection status and paths

### 2.6 Additional flags

- [x] `--include-refs` on install
- [x] `--dry-run` with `Describe()` output
- [x] `--json` on list, status, clients

### 2.7 Integration tests

- [x] Markdown adapter: new file, idempotent replace, append to existing, uninstall
- [x] Cursor adapter: install creates .mdc, uninstall removes, non-existent is no-op
- [x] Windsurf adapter: project file, global section markers

**Verification**: All 6 clients detected, dry-run works for each adapter format

---

## Phase 3: TUI Layer

### 3.1 Shared styles

- [x] Define Lip Gloss color palette and named styles
- [x] Status indicator symbols (`*` done, `o` active, `-` pending)

### 3.2 Client multi-select

- [x] Bubble Tea model with cursor, toggle, select-all/none
- [x] Pre-select detected clients by default
- [x] Show client name + install path per row
- [x] `RunClientSelect()` convenience function

### 3.3 Skill browser

- [x] Bubble Tea model with cursor and real-time filtering
- [x] Type to filter by name or directory name (substring match)
- [x] Show version next to each skill
- [x] `RunSkillSelect()` convenience function

### 3.4 Progress display

- [x] `ProgressItem` with label, detail, and status (pending/active/done/error)
- [x] ASCII progress bar `[====------] 2/5`
- [x] `PrintProgress()` for static output after install

### 3.5 Status table

- [x] `BuildStatusEntries()` groups installations by skill, maps clients to versions
- [x] `PrintStatusTable()` renders cross-client grid

### 3.6 CLI integration

- [x] `install` command: launch skill picker when no skill arg
- [x] `install` command: launch client multi-select when no `--client` flag
- [x] `install` command: show progress summary after multi-client install
- [x] `status` command: use TUI status table

**Verification**: Interactive install flow works end-to-end with TUI pickers

---

## Phase 4: Remote Skills & Polish

### 4.1 GitHub remote fetching

- [x] `FetchRemoteList()` — list skills from a GitHub repo via contents API
- [x] `FetchRemoteSkill()` — download full skill directory to `~/.aisk/cache/`
- [x] `ParseRepoURL()` — parse `github.com/owner/repo` format
- [x] `GITHUB_TOKEN` support for authenticated requests

### 4.2 Remote integration in CLI

- [x] `list --remote --repo owner/repo` fetches from GitHub
- [x] `AISK_REMOTE_REPO` env var as default repo

### 4.3 JSON output

- [x] `list --json` — array of skill objects
- [x] `status --json` — array of installation objects
- [x] `clients --json` — array of client objects

### 4.4 Build configuration

- [x] Makefile: build, test, lint, check, fmt, vet, snapshot, install, clean
- [x] GoReleaser: darwin/linux/windows × amd64/arm64, tar.gz/zip, checksums
- [x] Homebrew tap formula (GoReleaser config)

### 4.5 Documentation

- [x] README.md with install, quick start, command reference, adapter system, configuration

**Verification**: `aisk list --json`, `aisk clients --json`, `make build && make test`

---

## Phase 5: Future (Backlog)

### 5.1 Skill authoring

- [ ] `aisk create <name>` — scaffold a new skill directory with SKILL.md template
- [ ] `aisk lint <path>` — validate SKILL.md frontmatter, check for required fields

### 5.2 Project-scoped installs

- [ ] Automatic `.gitignore` management when installing to project scope
- [ ] Detect project root (look for `.git/`, `package.json`, `go.mod`)

### 5.3 Auto-update

- [ ] Compare manifest version vs available version on `aisk status`
- [ ] Notification when installed skills have newer versions

### 5.4 Distribution

- [ ] Publish Homebrew tap
- [ ] Shell completions: bash, zsh, fish via Cobra's built-in generator

### 5.5 Multi-repo support

- [ ] Configure multiple remote repos in `~/.aisk/config.json`
- [ ] `aisk list --remote` aggregates from all configured repos

### 5.6 Skill dependencies

- [ ] Support `depends-on` field in frontmatter
- [ ] Install dependencies automatically

---

## Test Coverage

| Package    | Test File          | Tests        | Focus                                                                                            |
| ---------- | ------------------ | ------------ | ------------------------------------------------------------------------------------------------ |
| `skill`    | `skill_test.go`    | 6            | Frontmatter parsing: valid, missing version, multi-line, missing delimiters, DisplayVersion      |
| `skill`    | `local_test.go`    | 3            | Scanner: discovery, references plural, hidden dir skip                                           |
| `skill`    | `remote_test.go`   | 1            | ParseRepoURL: various formats                                                                    |
| `adapter`  | `claude_test.go`   | 3            | Symlink, copy, uninstall                                                                         |
| `adapter`  | `markdown_test.go` | 4            | New file, idempotent replace, append to existing, uninstall                                      |
| `adapter`  | `cursor_test.go`   | 3            | Install, uninstall, uninstall non-existent                                                       |
| `adapter`  | `windsurf_test.go` | 2            | Project install, global install with markers                                                     |
| `manifest` | `manifest_test.go` | 7            | Add/find, replace existing, save/load cycle, remove, removeAll, allSkillNames, load non-existent |
| `client`   | `detect_test.go`   | 4            | Detect with config dirs, detect without, ParseClientID, registry.Detected                        |
| **Total**  | **9 files**        | **33 tests** | All pass with `-race`                                                                            |
