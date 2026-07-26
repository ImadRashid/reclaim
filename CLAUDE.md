# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`reclaim` is a developer-aware Mac cleaner CLI (Go 1.21+, single static binary, no CGO). It scans build/package/IDE caches and stale project artifacts, then deletes them with user confirmation. **Because this tool deletes files, safety logic is the highest-stakes part of the codebase.**

## Commands

```sh
go build -o reclaim .          # build
go vet ./...                   # CI runs this; gofmt required
go test ./...                  # all tests
go test -race ./...            # required if touching anything concurrent
go test ./internal/engine -run TestCheckPathSafe   # single test
./reclaim --plain              # scan report without TUI (safe way to test rules)
```

CI (`.github/workflows/ci.yml`) runs vet + build + test. Releases are cut with `scripts/release.sh vX.Y.Z` (renames `## Unreleased` in CHANGELOG, commits, tags, pushes); `release.yml` then cross-compiles darwin/linux × arm64/amd64 tarballs, publishes the GitHub Release, and pushes the filled-in formula template (`packaging/homebrew/reclaim.rb`) to `ImadRashid/homebrew-tap` using the `TAP_GITHUB_TOKEN` secret. The version string in `main.go` is overridden at build time via `-ldflags "-X main.version=..."`.

## Architecture

Data flow: **catalog → scan → select → apply → log**

1. `internal/rules/` — `catalog.yaml` is the rule catalog, **embedded into the binary via `go:embed`** (a YAML change requires a rebuild to take effect). `types.go` defines `Rule`/`Category`; `load.go` parses.
2. `internal/scanner/` — turns rules into `Hit`s (path + size). Three detect modes, dispatched in `scanRule`:
   - `dir-exists`: stat fixed paths (with `~` expansion)
   - `scan`: walk `scan_root` for directories named `scan_for` (supports `a|b` alternation, `path_contains`, `requires_sibling`, depth cap); does not descend into matched dirs
   - `command`: emits a synthetic pathless/zero-size hit, optionally gated on `detect_command` exiting 0
   - `git.go` adds git-staleness annotation (`activity_check: true` rules get `StalenessDays`).
3. `internal/ui/` — Bubble Tea TUIs: `welcome.go` (default entry menu), `tui.go` (picker), `info.go` (about/last-log). The TUI returns selections; deletion happens back in `main.go` → `engine`.
4. `internal/engine/` — the only code that deletes. `Apply()` handles process-check skips (`process_check`, `warn_if_running`), shell-action rules (`action: shell` / legacy `cmd:`), and path deletion via `os.RemoveAll`. `log.go` appends JSONL to `~/.reclaim/logs/YYYY-MM-DD.jsonl`.
5. `main.go` — flag parsing (hand-rolled, no flag package) and mode dispatch: no flags → welcome menu; `--pick` → picker TUI; `--plain` → text report; `--apply [-c category]` → non-interactive.

### Safety model (do not weaken)

- Three tiers on every rule: `safe` (regenerates, zero loss) / `confirm` (review first) / `dangerous` (never auto-deleted; TUI won't even select it). `--apply` deletes safe + confirm; "Quick clean" deletes safe only.
- `engine/safety.go` `CheckPathSafe` is the last line of defense: rejects system paths, anything outside `$HOME` (plus temp dirs for tests), the home dir itself, and direct children of home not on a hard-coded allowlist. **Every path deletion must go through `engine.Apply` / `CheckPathSafe` — never bypass it.**
- New code in `engine` (anything that decides to delete) must have a unit test **including a negative test proving it refuses unsafe input**. Use `t.TempDir()`; never write to hardcoded paths.

## Where changes go

- **New cleanup rule** → `internal/rules/catalog.yaml` only (no Go changes). Test with `./reclaim --plain`; PRs for new rules must paste that output.
- **New rule field** → `internal/rules/types.go` first, then wire into `scanner` and/or `engine`.
- **New CLI flag** → `main.go` (`parseArgs` + `printHelp`).
- **Safety check** → `internal/engine/safety.go` + test.

## Hard constraints (from CONTRIBUTING.md)

- No network calls, telemetry, or crash reporting — reclaim runs fully offline.
- Don't shell out to `rm`; use `os.RemoveAll`.
- No CGO; no new external dependencies without justification.
- Errors flow up; the only `os.Exit` is `die()` in `main.go`. No `panic()` in normal paths.
- Shell-command rules (`cmd:` / `action: shell`) are reserved for vendor-supplied cleaners (`brew cleanup`, `docker system prune`); rules must never delete user data outside conventional cache locations.
- YAML style: two-space indent, kebab-case ids, rules grouped by category in the same order as the `categories:` block, one-sentence present-tense descriptions (first 60 chars shown in plain mode).
- When unsure about a rule's safety tier, pick the more conservative one.

## Docs to keep in sync

User-facing behavior changes → update `README.md` and `CHANGELOG.md`. Roadmap and per-stack gap analysis live in `docs/ROADMAP.md`.
