# Contributing to reclaim

Thanks for considering a contribution! `reclaim` is a tool that **deletes
files**, so the bar for changes is higher than usual. This guide explains how
to propose changes safely.

## Table of contents

- [Ways to contribute](#ways-to-contribute)
- [Project layout](#project-layout)
- [Local development](#local-development)
- [Adding a cleanup rule (no code required)](#adding-a-cleanup-rule-no-code-required)
- [Code changes](#code-changes)
- [Testing requirements](#testing-requirements)
- [Pull request checklist](#pull-request-checklist)
- [Reporting bugs](#reporting-bugs)
- [Reporting security issues](#reporting-security-issues)
- [Style guide](#style-guide)
- [Releases](#releases)

---

## Ways to contribute

The most useful contributions, in roughly increasing order of effort:

1. **Report a bug** with a minimal repro and your `--version`.
2. **Suggest a new cleanup target** — open an issue with the path, what writes
   to it, what regenerates it, and proof it's safe to delete.
3. **Submit a YAML rule** for a tool we don't cover (this is the easiest
   code-free contribution — see below).
4. **Improve docs** — typos, missing keybindings, clearer explanations.
5. **Fix a bug** with a test that reproduces it.
6. **Add a feature** from the [Roadmap](README.md#roadmap) — please open an
   issue to discuss first so we don't duplicate work.

We **don't** accept:

- Rules that delete user data outside conventional cache locations (browser
  profiles, mail archives, photos, Downloads).
- Rules that auto-run shell commands a user hasn't seen (e.g. `rm -rf` via
  `cmd:`). Shell rules are reserved for vendor-supplied cleaners (`brew
  cleanup`, `npm cache clean`).
- Telemetry or any kind of "phone home."

---

## Project layout

```
reclaim/
├── main.go                      # CLI entry point + argument parsing
├── internal/
│   ├── rules/                   # YAML loader + types (catalog embedded at build)
│   │   ├── catalog.yaml         # ★ The cleanup rule catalog ★
│   │   ├── load.go
│   │   └── types.go
│   ├── scanner/                 # Filesystem walking + git staleness
│   │   ├── scanner.go           # Path resolution, dir size, pattern matching
│   │   ├── git.go               # Last-commit detection
│   │   └── *_test.go
│   ├── engine/                  # Apply / delete / safety
│   │   ├── engine.go            # Apply() entrypoint
│   │   ├── safety.go            # CheckPathSafe — refuses unsafe paths
│   │   ├── log.go               # ~/.reclaim/logs/ JSONL writer
│   │   └── *_test.go
│   └── ui/                      # Bubble Tea TUI
│       └── tui.go
├── packaging/
│   └── homebrew/reclaim.rb      # Homebrew formula template
├── .github/workflows/
│   ├── ci.yml                   # vet + build + test
│   └── release.yml              # tag → cross-compiled tarballs + GitHub Release
└── go.mod / go.sum
```

---

## Local development

Prereqs: **Go 1.21+** and a Mac (Linux works for most things, but the catalog
targets macOS paths).

```sh
git clone https://github.com/ImadRashid/reclaim
cd reclaim
go build -o reclaim .         # produces ./reclaim
./reclaim --version
./reclaim --plain             # see what it found, without deleting
./reclaim                     # interactive TUI

# Run tests
go test ./...
go test -race ./...           # catch concurrency issues

# Run vet (CI does this too)
go vet ./...

# Build for another platform
GOOS=linux GOARCH=arm64 go build -o reclaim-linux-arm64 .
```

---

## Adding a cleanup rule (no code required)

This is the most common contribution and **does not require Go knowledge**.

Open `internal/rules/catalog.yaml` and add an entry. Two shapes:

### 1. Fixed-path rule

For caches that always live at the same path:

```yaml
- id: my-tool-cache
  category: package-caches              # see "categories:" at top of file
  paths:
    - ~/Library/Caches/my-tool          # ~ is expanded; multiple paths allowed
  safety: safe                          # safe | confirm | dangerous
  regenerates: true                     # true means it auto-rebuilds; tell users
  description: My tool's download cache. Re-downloads on next use.
  detect: dir-exists                    # only consider it if the dir exists
```

### 2. Pattern rule

For per-project caches discovered by walking the filesystem:

```yaml
- id: my-project-cache
  category: project-caches
  scan_for: .my-cache                   # directory name to look for
  scan_root: ~/Desktop/Projects         # where to start walking
  scan_max_depth: 5                     # don't recurse forever
  path_contains: src                    # OPTIONAL: substring match on full path
  requires_sibling: package.json        # OPTIONAL: only match if a sibling file exists
  safety: safe
  regenerates: true
  description: Per-project My-tool build cache
  detect: scan
```

### 3. Vendor-cleaner rule

For tools that have their own cleanup command (rare; needs review):

```yaml
- id: ios-simulators-unavailable
  category: ide-caches
  cmd: xcrun simctl delete unavailable
  safety: safe
  description: Remove iOS simulators left behind from old Xcode versions.
  detect: command
```

### Rule fields reference

| Field | Type | Required | Notes |
|-------|------|:---:|---|
| `id` | string | ✓ | Unique within catalog. kebab-case. |
| `category` | string | ✓ | Must match one of the `categories:` at top of file. |
| `paths` | []string | for `dir-exists` | Absolute paths or `~/...` |
| `scan_for` | string | for `scan` | Directory basename, or `a\|b` for alternatives |
| `scan_root` | string | for `scan` | Root of the walk |
| `scan_max_depth` | int |  | Default 5 |
| `path_contains` | string |  | Restrict matches to paths containing this substring |
| `requires_sibling` | string |  | Match only if this file/dir exists in the same parent |
| `safety` | enum | ✓ | `safe` \| `confirm` \| `dangerous` |
| `regenerates` | bool |  | True if the path is recreated by normal use |
| `cmd` | string | for `command` | Shell command, only for vendor cleaners |
| `description` | string | ✓ | One sentence; first 60 chars are shown in plain mode |
| `detect` | enum | ✓ | `dir-exists` \| `scan` \| `command` |
| `activity_check` | bool |  | If true, scanner annotates hits with git staleness |
| `process_check` | string |  | Skip the rule's hits if this process is running |

### Safety guidelines for new rules

- **`safe`** is for paths that regenerate automatically with **zero user-visible
  loss**. Build caches, package downloads, browser asset caches.
- **`confirm`** is for things a thoughtful user might want to inspect first.
  Examples: `node_modules` (regenerates but takes minutes), Trash, virtualenvs.
- **`dangerous`** is anything that loses real data: profiles, archives,
  history, mail, anything in `Documents`/`Desktop`/`Downloads`.

When in doubt, use a more conservative tier. We'd rather miss a few MB than
delete something a user wanted.

---

## Code changes

### Branching

- Branch off `main`. Use a descriptive name: `add-deno-cache`, `fix-tui-scroll`.
- Keep PRs focused. One feature/fix per PR. Big PRs are hard to review.

### What goes where

- **New rule** → `internal/rules/catalog.yaml` only. No Go changes needed.
- **New rule field** → `internal/rules/types.go` first, then wire it into
  `scanner` and/or `engine` as appropriate.
- **TUI tweak** → `internal/ui/tui.go`.
- **New CLI flag** → `main.go` (`parseArgs` + `printHelp`).
- **New safety check** → `internal/engine/safety.go` + a test.

### Don't

- Don't add network calls. `reclaim` runs fully offline.
- Don't add telemetry, crash reporting, or analytics.
- Don't shell out to `rm`. Use `os.RemoveAll` (it's safer and gives us errors).
- Don't bypass `CheckPathSafe`. If your code path needs to delete, it goes
  through `engine.Apply`.
- Don't introduce CGO. We want a single static binary.

---

## Testing requirements

- Every PR must pass `go vet ./...` and `go test ./...`.
- New rules don't need Go tests, but **paste a sample of `./reclaim --plain`
  output in the PR description** showing your rule fires correctly on a real
  machine.
- New code in `engine` (anything that decides to delete) **must** have a unit
  test, including a negative test that proves it refuses an unsafe input.
- Use `t.TempDir()` for any test that creates files. Never write to a hardcoded
  path.

---

## Pull request checklist

Before submitting, verify each:

- [ ] `go vet ./...` is clean
- [ ] `go test ./...` passes
- [ ] `go test -race ./...` passes (if you touched anything concurrent)
- [ ] If you added a rule, you tested it with `./reclaim --plain` and pasted
      the output in the PR description
- [ ] If you added user-facing behavior, you updated `README.md` and/or
      `CHANGELOG.md`
- [ ] If you changed safety logic, you added a test that proves it rejects
      unsafe input
- [ ] No new external dependencies, or you've justified them in the PR
- [ ] Commit messages explain *why*, not just *what*

---

## Reporting bugs

Open an issue with:

1. `reclaim --version`
2. macOS version (`sw_vers`)
3. Exact command you ran
4. What you expected
5. What happened (paste the output, including any error)
6. If the bug involves the TUI, a screen recording or screenshot helps

If `reclaim` deleted something it shouldn't have, **also include**:

- The relevant log line(s) from `~/.reclaim/logs/YYYY-MM-DD.jsonl`
- Whether the affected path was in the catalog or scanned dynamically

We treat unwanted deletion as a critical bug.

---

## Reporting security issues

Don't open a public issue for security problems — see [SECURITY.md](SECURITY.md).

---

## Style guide

### Go

- Follow standard Go conventions. We don't run `gofumpt`, but `gofmt` is
  required (CI enforces via `go vet`).
- Package names are short and lowercase. No `util` or `helpers` packages.
- Comments are for *why*, not *what*. The code says what.
- Errors flow up. The only `os.Exit` is `die()` in `main.go`.
- No `panic()` in normal code paths. Use it for "this can't happen" only.

### YAML

- Two-space indentation.
- Group rules by category, in the same order as `categories:` at the top.
- IDs are kebab-case.
- Descriptions are one sentence, present tense ("Cleans X" not "Will clean X").

---

## Releases

Releases are cut by maintainers via tags:

```sh
git tag v0.2.0 -m "v0.2.0"
git push origin v0.2.0
```

The `release.yml` workflow then:

1. Cross-compiles `darwin/arm64`, `darwin/amd64`, `linux/arm64`, `linux/amd64`
2. Creates `.tar.gz` archives with `README.md` + `LICENSE`
3. Generates `.sha256` files
4. Publishes a GitHub Release with auto-generated notes

After the release, the maintainer updates `packaging/homebrew/reclaim.rb` with
the new version + SHAs and pushes to the `homebrew-tap` repo.

Contributors don't need to do any of this — just merge to `main`.

---

Thanks again for contributing. If anything in this guide is unclear, that's a
bug in the docs — please open an issue or PR to fix it.
