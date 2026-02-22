# aisk Prioritized Roadmap

This roadmap captures additional high-value features to improve reliability, automation, governance, and operator ergonomics.

## Now

Milestone: `Core workflows and safety`

- Install plan preview (`aisk plan`) before apply — `BRN-16`
- Non-interactive automation mode (`--yes`) with strict behavior — `BRN-17`
- Manifest drift detection and repair (`aisk doctor`) — `BRN-18`
- Capabilities matrix command (`aisk capabilities`) for clients/scopes/formats — `BRN-19`
- Repository-local defaults (`.aisk/config.json`) for project workflows — `BRN-20`
- Remote trust/pinning controls for reproducible installs — `BRN-21`
- Rollback support (`aisk rollback <run-id|timestamp>`) — `BRN-22`

## Next

Milestone: `Team operations and policy`

- Backup/restore for manifest + install state — `BRN-23`
- Team policy engine (allowed clients/scopes/remotes, required checks) — `BRN-24`
- Dependency graph tooling (`aisk deps`) with cycle/conflict visibility — `BRN-25`

## Later

Milestone: `Advanced trust and scale`

- Skill signature verification for remote installs — `BRN-26`
- Bulk operations by tag/label (install/update sets of skills) — `BRN-27`
- Template-based scaffolding (`aisk create --template`) — `BRN-28`
- Rich linting/autofix suggestions (references/style/policy) — `BRN-29`
- Local operational metrics (success/failure/update cadence) — `BRN-30`
- Cross-client golden/integration test matrix — `BRN-31`

## Existing Tracked Work

Already tracked in Linear and not duplicated here:

- Distribution: Homebrew tap (`BRN-5`)
- Shell completions (`BRN-6`, completed)
- Multi-repo support (`BRN-7`, `BRN-8`)
- Skill dependencies (`BRN-9`, `BRN-10`)
- Audit hardening (`BRN-11` through `BRN-15`)
