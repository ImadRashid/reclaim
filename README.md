# reclaim

**A developer-aware Mac cleaner CLI.** Scans build caches, package manager
caches, IDE artifacts, and stale project dependencies, then lets you free disk
space without touching anything important.

```
$ reclaim
🧹 reclaim — interactive cleanup
Selected: 21.4 GB of 47.2 GB

▼ Build caches — 25 GB [3/3 items, 25 GB selected]
    [✓] ✓ ~/.gradle/caches                                  19.2 GB
    [✓] ✓ ~/.gradle/wrapper                                   4.2 GB
    [✓] ✓ ~/.gradle/daemon                                    1.3 GB

▼ Package manager caches — 7.7 GB [4/8 items, 4 GB selected]
    [✓] ✓ ~/.npm/_cacache                                    3.3 GB
    [ ] ✓ ~/.pub-cache                                       3.1 GB
    [✓] ✓ ~/Library/Caches/CocoaPods                         381 MB
    [✓] ✓ ~/Library/pnpm/store                               358 MB

▼ IDE & editor caches — 8.3 GB [1/3 items, 8.0 GB selected]
    [✓] ✓ ~/Library/Developer/Xcode/DerivedData              8.0 GB
    [ ] ✓ ~/Library/Application Support/Code/...             405 MB
    [ ] ⚠ ~/Library/Developer/Xcode/Archives                     0 B

↑/↓ move  space toggle  ←/→ collapse  a all  n none  enter apply  q quit
```

---

## Why another cleaner?

Most disk cleaners are general-purpose. `reclaim` is for developers who
accumulate build caches, `node_modules`, SDK installs, and IDE artifacts from
many languages and frameworks. It knows the difference between *safe to delete*
(regenerates automatically) and *be careful* (your profile data, that 6-month-old
project's `target/`).

Highlights:

- **Knows your tools** — Gradle, Maven, Go, Cargo, npm, yarn, pnpm, pip, uv, poetry, pub, CocoaPods, Homebrew, Docker, Xcode, VS Code, JetBrains, browsers, AI/ML caches.
- **Activity-aware** — reads `git log` to detect stale projects, so you can clean abandoned `node_modules` without nuking active ones.
- **Safe by default** — dry-run mode, three-tier safety (`safe`/`confirm`/`dangerous`), process detection (won't suggest cleaning Docker if it's running).
- **Auditable** — every apply run logs to `~/.reclaim/logs/YYYY-MM-DD.jsonl`.
- **Single binary** — no runtime dependencies. Install via Homebrew or grab a release.

---

## Install

### Homebrew (recommended)

```sh
brew install imadrashid/tap/reclaim
```

### From a release archive

```sh
# Apple Silicon
curl -L https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.0_darwin_arm64.tar.gz | tar -xz
sudo mv reclaim_v0.1.0_darwin_arm64/reclaim /usr/local/bin/

# Intel Mac
curl -L https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.0_darwin_amd64.tar.gz | tar -xz
```

### From source

```sh
git clone https://github.com/ImadRashid/reclaim
cd reclaim
go build -o reclaim .
./reclaim --help
```

Requires Go 1.21+.

---

## Usage

```sh
reclaim                              # Interactive TUI: scan, pick, apply
reclaim --plain                      # Print plain-text report (no TUI)
reclaim --apply                      # Non-interactive apply (safe + confirm items)
reclaim --apply -c build-caches      # Limit to one category
reclaim --version
reclaim --help
```

### TUI keys

| Key | Action |
|-----|--------|
| `↑` / `↓` / `j` / `k` | Navigate |
| `space` | Toggle item (or select/deselect a whole category) |
| `→` / `l`, `←` / `h` | Expand / collapse category |
| `a` | Select all non-dangerous items |
| `n` | Deselect everything |
| `enter` | Apply selected (with confirmation) |
| `q` / `esc` | Quit without changes |

### Safety levels

| Marker | Meaning |
|:------:|---------|
| ✓ | **safe** — regenerates automatically; deletion has no impact |
| ⚠ | **confirm** — should be reviewed before deleting |
| ✗ | **dangerous** — skipped by default; handle manually |

`--apply` (non-interactive) will delete `safe` and `confirm` items only.
The TUI lets you select anything except `dangerous`.

### Run history

Every apply run appends a JSONL record to `~/.reclaim/logs/YYYY-MM-DD.jsonl`,
recording rule id, path, freed bytes, errors, and skipped items. Useful for
auditing what changed and when.

---

## What it knows about

| Category | What gets cleaned |
|----------|-------------------|
| **Build caches** | `~/.gradle/{caches,wrapper,daemon}`, `~/.m2/repository`, `~/Library/Caches/go-build`, `~/.cargo/{registry,git}`, `~/.rustup` |
| **Package managers** | npm, yarn, pnpm, pip, uv, poetry, pub, CocoaPods, Homebrew, Hugging Face, Ollama, PyTorch |
| **Project caches** | `node_modules`, `.next`, `.open-next`, `.dart_tool`, `Pods`, `android/.gradle`, `__pycache__`, virtualenvs, Rust `target/` |
| **IDE / Xcode** | DerivedData, iOS/watchOS/tvOS DeviceSupport, simulator runtimes, VS Code & JetBrains caches |
| **Containers** | Docker Desktop VM disk |
| **Browsers** | Chrome, Arc caches (won't log you out) |

The full catalog lives in [`internal/rules/catalog.yaml`](internal/rules/catalog.yaml). Adding a
rule is a YAML edit — no code changes required.

---

## Adding your own rules

Edit `internal/rules/catalog.yaml` and add an entry like:

```yaml
- id: my-custom-cache
  category: package-caches
  paths:
    - ~/Library/Caches/my-tool
  safety: safe
  regenerates: true
  description: My tool's download cache
  detect: dir-exists
```

For path-pattern rules (e.g. find every `.foo-cache` under a project root):

```yaml
- id: my-project-cache
  category: project-caches
  scan_for: .foo-cache
  scan_root: ~/Desktop/Projects
  scan_max_depth: 5
  safety: safe
  description: Foo build cache
  detect: scan
```

Rebuild with `go build` and run.

---

## Roadmap

- [x] Interactive TUI with checkboxes and category groups
- [x] Activity-aware project staleness detection
- [x] Run history / audit log
- [x] Single static binary distribution
- [ ] Plugin / external rule files (`~/.reclaim/rules.d/*.yaml`)
- [ ] Free space delta in summary (before/after `df -h`)
- [ ] launchd integration for scheduled cleanups
- [ ] Linux-specific rule pack (XDG cache dirs)
- [ ] `reclaim undo` (restore from `~/.Trash` if items were trashed not deleted)

---

## License

MIT — see [LICENSE](LICENSE).
