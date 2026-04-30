# Changelog

All notable changes are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## v0.1.1 — 2026-04-30

### Added

- **Welcome menu** as the default entry point. Running `reclaim` with no flags
  now opens a menu first; nothing is scanned until the user picks an action.
  Options: Quick scan & pick, Quick clean (safe defaults), Browse by category,
  Last cleanup log, About / help, Quit.
- **`--pick` flag** to skip the welcome menu and go straight to the picker TUI
  (the previous default behavior).
- **About screen** showing version, links, safety levels, TUI keys, and how to
  skip the menu via flags.
- **Last cleanup log viewer** that summarizes the most recent run from
  `~/.reclaim/logs/` (deleted/skipped/failed counts, total freed, recent paths).

### Changed

- README now documents both the one-liner Homebrew install
  (`brew install ImadRashid/tap/reclaim`) and the tap-then-install short form
  (`brew tap ImadRashid/tap && brew install reclaim`). Also explains why plain
  `brew install reclaim` doesn't work yet (homebrew-core acceptance criteria).
- README includes detailed direct-download instructions for all four release
  archives (darwin/linux × arm64/amd64), with a note about macOS Gatekeeper
  quarantine and how to bypass it.
- Help text and module docs updated to describe the welcome menu.

## v0.1.0 — 2026-04-30

Initial release.

### Added

- Interactive Bubble Tea TUI as the default mode (checkboxes, expandable categories, size totals).
- Plain-text scan mode (`--plain`) and non-interactive apply (`--apply`).
- 38 cleanup rules across 8 categories — Node, Python, Rust, Go, Flutter, Java/Gradle, Maven, Docker, Xcode, browsers, and more.
- Activity-aware project scanning: detects last `git log` for each project and labels staleness (active / recent / idle / stale / abandoned).
- Per-rule safety levels — `safe`, `confirm`, `dangerous`. Dangerous items are skipped by default.
- Process-aware deletion: skips items if a named app (e.g. Docker) is running.
- Run history written to `~/.reclaim/logs/YYYY-MM-DD.jsonl` after every apply.
- Single static binary distribution via GitHub Releases for darwin/linux × arm64/amd64.
- Homebrew formula template at `packaging/homebrew/reclaim.rb`.
