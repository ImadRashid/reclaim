# Changelog

All notable changes are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Added

- **Interactive scan screen**: scanning from the welcome menu or `--pick`
  now shows an animated spinner, which rule is being checked (`12/77`), a
  live tally of items and size found, elapsed time, and can be cancelled
  with `q`. `--plain` and `--apply` keep the simple text progress.

- **Once-daily update check** (interactive modes only): a background request
  to the GitHub releases API, cached for 24h in
  `~/.reclaim/update-check.json`, prints a one-line upgrade notice on exit
  when a newer version exists — suggesting `brew upgrade reclaim` for brew
  installs and the releases page otherwise. Nothing is downloaded or
  installed automatically. Opt out with `--no-update-check` or
  `RECLAIM_NO_UPDATE_CHECK=1`; `--plain` and `--apply` never check. This is
  the only network call reclaim makes.

### Changed

- README refreshed for v0.2.0: download one-liners point at the current
  release assets, the coverage table lists the new categories and rules, and
  the roadmap section targets v0.3.
- CONTRIBUTING documents the new rule fields (`detect_command`, `action`,
  `action_command`, `warn_if_running`); the vendor-cleaner example now uses
  the `action: shell` form.
- `scripts/release.sh` now bumps the versioned asset names in the README
  automatically as part of the release commit.
- ROADMAP updated: v0.2.0 marked shipped with what landed vs. slipped.

## v0.2.0 — 2026-07-26

### Added

- **39 new cleanup rules** across the catalog: JS tooling (Bun, Deno, Yarn
  Berry, npm logs, Turborepo, nvm), Python (pipenv, conda), Ruby gems,
  Composer, NuGet, Carthage, Go module cache, JVM (Gradle native, Ivy, sbt),
  Bazel, AWS CLI and Terraform plugin caches, Android (build cache, unused
  AVDs, stale system images), Xcode (Previews, device logs, CoreSimulator
  caches, IPSW firmware), fvm stale Flutter SDKs, editors (Cursor, Windsurf,
  Zed, VS Code logs, JetBrains logs), Docker (`system prune`, buildx cache),
  Colima/Lima VMs, Cypress, Prisma, and the Dart analysis server cache.
- **New "Testing tool caches" category** (Playwright, Cypress).
- **New rule capabilities**: `detect_command` gates a rule on a probe command
  exiting 0 (e.g. only show `docker system prune` when the daemon responds),
  `action: shell` runs a vendor-supplied cleanup command instead of deleting
  paths, and `warn_if_running` skips deletion while any listed app is open.
- **`o` key in the picker TUI** opens the highlighted folder in Finder
  (`xdg-open` on Linux) so you can inspect contents before deleting.
- **Scroll indicator** below the list showing how many rows are hidden
  off-screen (e.g. `… ↑ 5 more above · ↓ 19 more below`).
- **One-command releases**: `scripts/release.sh vX.Y.Z` handles the CHANGELOG
  rename, commit, tag, and push; the release workflow now also updates the
  Homebrew tap formula automatically (new `homebrew` job in `release.yml`).

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
