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

### Homebrew

One-liner (taps and installs in one go):

```sh
brew install ImadRashid/tap/reclaim
```

Or tap once, then install by short name forever:

```sh
brew tap ImadRashid/tap
brew install reclaim
```

To upgrade later: `brew upgrade reclaim`. To uninstall: `brew uninstall reclaim`.

> Why not plain `brew install reclaim`? That's reserved for tools in the
> official `homebrew-core` registry, which has acceptance criteria like a
> minimum number of stars. `reclaim` will apply once it has real adoption.

### Direct download (prebuilt binary, no build step)

Each release ships a single ready-to-run binary — no compile, no runtime
needed. Pick the archive for your machine and copy the `reclaim` file onto
your `$PATH`. One-line installer:

```sh
# Apple Silicon (M1/M2/M3/M4)
curl -fsSL https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.1_darwin_arm64.tar.gz \
  | tar -xz \
  && sudo install -m 0755 reclaim_v0.1.1_darwin_arm64/reclaim /usr/local/bin/reclaim \
  && rm -rf reclaim_v0.1.1_darwin_arm64

# Intel Mac
curl -fsSL https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.1_darwin_amd64.tar.gz \
  | tar -xz \
  && sudo install -m 0755 reclaim_v0.1.1_darwin_amd64/reclaim /usr/local/bin/reclaim \
  && rm -rf reclaim_v0.1.1_darwin_amd64

# Linux ARM64
curl -fsSL https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.1_linux_arm64.tar.gz \
  | tar -xz \
  && sudo install -m 0755 reclaim_v0.1.1_linux_arm64/reclaim /usr/local/bin/reclaim \
  && rm -rf reclaim_v0.1.1_linux_arm64

# Linux AMD64
curl -fsSL https://github.com/ImadRashid/reclaim/releases/latest/download/reclaim_v0.1.1_linux_amd64.tar.gz \
  | tar -xz \
  && sudo install -m 0755 reclaim_v0.1.1_linux_amd64/reclaim /usr/local/bin/reclaim \
  && rm -rf reclaim_v0.1.1_linux_amd64
```

Then run `reclaim` from anywhere.

Prefer to grab the file by hand? Open the
[releases page](https://github.com/ImadRashid/reclaim/releases),
download the archive for your platform, double-click to extract, and drag the
`reclaim` binary into a PATH directory like `/usr/local/bin/`.

> **Gatekeeper note (macOS):** the first time you run a directly-downloaded
> binary, macOS may say *"reclaim cannot be opened because the developer
> cannot be verified."* This is normal for unsigned binaries. Two ways to
> bypass it:
>
> 1. Right-click `reclaim` in Finder → **Open** (one-time approval), or
> 2. Run `xattr -d com.apple.quarantine /usr/local/bin/reclaim` in a terminal.
>
> Homebrew installs don't trigger this because brew strips the quarantine
> attribute automatically.

#### Verify the download (optional)

Each release also publishes `.sha256` sidecar files. To verify integrity:

```sh
shasum -a 256 reclaim_v0.1.1_darwin_arm64.tar.gz
# Compare against the value in reclaim_v0.1.1_darwin_arm64.tar.gz.sha256
```

### Build from source (only if you want to)

You don't need to build from source to use `reclaim` — the prebuilt binary
above is the same thing you'd produce. This path is for contributors and for
users who'd rather compile their own.

```sh
git clone https://github.com/ImadRashid/reclaim
cd reclaim
go build -o reclaim .
./reclaim --help
```

Requires **Go 1.21+**. The resulting binary is a single static file with no
runtime dependencies.

---

## Usage

```sh
reclaim                              # Welcome menu (default entry point)
reclaim --pick                       # Skip welcome, go straight to picker TUI
reclaim --plain                      # Print plain-text report (no TUI, no I/O until you confirm)
reclaim --apply                      # Non-interactive apply (safe + confirm items)
reclaim --apply -c build-caches      # Limit to one category
reclaim --version
reclaim --help
```

### The welcome menu

Running `reclaim` with no flags now opens a menu first — nothing is scanned
until you pick an action.

```
🧹 reclaim — developer-aware Mac cleaner

  Find and free disk space taken by build caches, package
  managers, IDE artifacts, and stale project dependencies.

  Nothing is deleted without your explicit confirmation.

  What would you like to do?

  ▶ Quick scan & pick           Scan and choose what to clean
    Quick clean (safe defaults) Scan and clean only ✓ safe items
    Browse by category          Pick a single category to scan
    Last cleanup log            See what was deleted last time
    About / help                Keys, safety model, docs
    Quit
```

Power users can skip the menu with any flag (`--pick`, `--plain`, `--apply`).

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

The full roadmap, gap analysis by stack (Flutter, RN, Node, Python, etc.),
and tentative release plan lives in [`docs/ROADMAP.md`](docs/ROADMAP.md).

Headline items for the next release (v0.2):

- fvm SDK versions and per-version Xcode simulator runtime cleanup (the
  two biggest single wins on a Flutter/iOS dev's machine)
- Bun, Turborepo, Cypress, Metro / Expo
- conda environments, pyenv versions, NuGet packages
- Free disk space delta in the summary

After that: per-project view in the TUI, `--trash` (soft delete), preset
profiles, `reclaim doctor`, external rule files. See the roadmap for details.

---

## License

MIT — see [LICENSE](LICENSE).
