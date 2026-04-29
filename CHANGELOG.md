# Changelog

All notable changes are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
