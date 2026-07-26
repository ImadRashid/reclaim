# reclaim — Rule expansion plan (v0.2 and beyond)

> A working document for `internal/rules/catalog.yaml` growth. Catalogues every
> developer cache, build artifact, and language-toolchain directory worth covering,
> organised into reclaim's existing rule format. Each entry is annotated with
> safety tier and detection mode so the catalog can grow without engine changes
> until §4 features are implemented.
>
> **Scope.** Only paths where deletion is genuinely safe (caches, build artifacts,
> regenerable data). User-data-adjacent locations (browser profiles, app
> preferences, documents, mail, photos) are excluded by design.

---

## 1. Coverage status

Legend: ✅ = covered, 🟡 = partial / behind a flag, ❌ = not yet covered

| Category / path | reclaim v0.1.1 |
|---|---|
| **Build / language caches** | |
| `~/.gradle/caches` | ✅ |
| `~/.gradle/wrapper` | ✅ |
| `~/.gradle/daemon` | ✅ |
| `~/.gradle/native` | ❌ |
| `~/.m2/repository` | ✅ |
| `~/.ivy2/cache` | ❌ |
| `~/.sbt/boot` | ❌ |
| `~/Library/Caches/go-build` | ✅ |
| `~/go/pkg/mod` (Go module cache) | ❌ |
| `~/.cargo/registry` | ✅ |
| `~/.cargo/git` | ✅ |
| `~/.rustup/toolchains` (stale) | 🟡 (whole `.rustup`) |
| **Package managers (JS)** | |
| `~/.npm/_cacache` | ✅ |
| `~/.npm/_logs` | ❌ |
| `~/Library/Caches/Yarn` | 🟡 |
| `~/.yarn/berry/cache` (Yarn 2+) | ❌ |
| `~/Library/pnpm/store` | ✅ |
| `~/.bun/install/cache` | ❌ |
| `~/.deno/deps` | ❌ |
| `~/.deno/gen` | ❌ |
| `~/.cache/turbo` (Turborepo) | ❌ |
| **Package managers (Python)** | |
| `~/Library/Caches/pip` | ✅ |
| `~/.cache/pip` (Linux) | 🟡 |
| `~/Library/Caches/uv` | ✅ |
| `~/Library/Caches/pypoetry` | ✅ |
| `~/.cache/pipenv` | ❌ |
| `~/.conda/pkgs` | ❌ |
| `~/anaconda3/pkgs`, `~/miniconda3/pkgs` | ❌ |
| `~/.pyenv/versions` (stale) | ❌ |
| **Package managers (Ruby)** | |
| `~/.gem/cache` | ❌ |
| `~/.rbenv/versions` (stale) | ❌ |
| `~/.rvm/gems` | ❌ |
| **Package managers (other)** | |
| `~/Library/Caches/CocoaPods` | ✅ |
| `~/.pub-cache` (Dart/Flutter) | ✅ |
| `~/Library/Caches/Homebrew` | ✅ |
| `~/.composer/cache` (PHP) | ❌ |
| `~/.nuget/packages` | ❌ |
| `~/.cache/Carthage` | ❌ |
| **Xcode** | |
| `~/Library/Developer/Xcode/DerivedData` | ✅ |
| `~/Library/Developer/Xcode/Archives` | 🟡 (dangerous) |
| `~/Library/Developer/Xcode/iOS DeviceSupport` | ✅ |
| `~/Library/Developer/Xcode/watchOS DeviceSupport` | ✅ |
| `~/Library/Developer/Xcode/tvOS DeviceSupport` | ✅ |
| `~/Library/Developer/Xcode/UserData/Previews` | ❌ |
| `~/Library/Developer/Xcode/iOS Device Logs` | ❌ |
| `~/Library/Developer/CoreSimulator/Caches` | ❌ |
| `~/Library/Developer/CoreSimulator/Devices` (unused) | 🟡 |
| `~/Library/Developer/CoreSimulator/Volumes/Cryptex` (unused) | ❌ |
| Stale per-runtime simulator runtimes | ❌ |
| **Flutter / Dart** | |
| `~/.dartServer` | ❌ |
| `~/fvm/versions` (stale) | ❌ |
| Per-project `.dart_tool` | ✅ |
| Per-project `build/` (Flutter) | 🟡 |
| **Android** | |
| `~/.android/avd` (unused emulators) | ❌ |
| `~/Library/Android/sdk/system-images` (stale) | ❌ |
| `~/Library/Android/sdk/platforms` (stale) | ❌ |
| `~/.android/build-cache` | ❌ |
| Per-project `android/.gradle` | ✅ |
| Per-project `android/build` | 🟡 |
| **IDEs / Editors** | |
| VS Code `Cache`, `CachedData`, `Code Cache`, `GPUCache` | 🟡 |
| VS Code `logs/` | ❌ |
| Cursor caches (`~/Library/Application Support/Cursor/Cache*`) | ❌ |
| Windsurf caches | ❌ |
| Zed caches (`~/Library/Caches/dev.zed.Zed`) | ❌ |
| JetBrains `~/Library/Caches/JetBrains/*` | ✅ |
| JetBrains `~/Library/Logs/JetBrains/*` | ❌ |
| **Containers / VMs** | |
| Docker Desktop VM disk | ✅ |
| `docker system prune` (opt-in) | ❌ |
| `~/.docker/buildx/cache` | ❌ |
| Colima (`~/.colima`) | ❌ |
| Lima (`~/.lima`) | ❌ |
| **CI / build tools** | |
| `~/.cache/bazel` | ❌ |
| `~/.cache/terraform` (plugin cache) | ❌ |
| `~/.terraform.d/plugin-cache` | ❌ |
| **Browsers (dev only — caches, never profiles)** | |
| Chrome cache subset | 🟡 (Chrome only) |
| Brave cache subset | ❌ |
| Safari `~/Library/Caches/com.apple.Safari` | ❌ |
| Firefox profile caches | ❌ |
| Edge cache subset | ❌ |
| Arc | ✅ |
| **AI / ML** | |
| `~/.cache/huggingface` | ✅ |
| `~/.ollama/models` | ✅ |
| `~/.cache/torch` | ✅ |
| `~/Library/Caches/Claude` | ❌ |
| **Testing tools** | |
| `~/.cache/ms-playwright` / `~/Library/Caches/ms-playwright` | ❌ |
| `~/.cache/puppeteer` | ❌ |
| `~/.cache/Cypress` | ❌ |
| `~/.prisma` | ❌ |
| **Game engines** | |
| Unity caches | ❌ |
| Unity Hub caches | ❌ |
| Godot caches | ❌ |
| Unreal `DerivedDataCache` | ❌ |
| **Version managers** | |
| `~/.nvm/.cache` (stale node tarballs) | ❌ |
| `~/.local/share/mise` (stale) | ❌ |
| `~/.asdf/installs` (stale) | ❌ |
| **Cloud CLIs** | |
| `~/.aws/cli/cache` | ❌ |
| `~/.azure/logs` | ❌ |
| `~/.config/gcloud/logs` | ❌ |
| **System (low-effort, big-win)** | |
| `~/.Trash` (with consent) | ❌ |
| `~/Library/Caches/com.apple.dt.Xcode` | 🟡 |
| Old `*.ipsw` firmware | ❌ |
| `~/Library/Logs/CrashReporter` | ❌ |
| **Project-walker artifacts** *(activity-aware via `git log`)* | |
| `node_modules` | 🟡 |
| `.next`, `.nuxt`, `.svelte-kit` | 🟡 (Next only) |
| `.open-next` | ✅ |
| `dist`, `build`, `out` | 🟡 |
| Rust `target/` | ✅ |
| Maven/Gradle `target/`, `build/` (per project) | 🟡 |
| Python `__pycache__`, `.pytest_cache`, `.mypy_cache`, `.ruff_cache` | 🟡 |
| Python `venv/`, `.venv/`, `env/` | ❌ |
| `Pods/` (CocoaPods) | ✅ |
| `.dart_tool` | ✅ |
| Unity `Library/`, `Temp/`, `Logs/` | ❌ |
| Unreal `Binaries/`, `Intermediate/`, `DerivedDataCache/`, `Saved/` | ❌ |
| Godot `.godot/` | ❌ |
| .NET `bin/`, `obj/` | ❌ |
| Elixir `_build/`, `deps/` | ❌ |
| Haskell `.stack-work/`, `dist-newstyle/` | ❌ |
| Zig `zig-cache/`, `zig-out/` | ❌ |
| Terraform `.terraform/` | ❌ |

---

## 2. Where reclaim is differentiated

The catalog above shows we have plenty of room to grow rule coverage, but the
project's design choices already give it a defensible position:

1. **Rules live in YAML, not Go.** Adding support for a new tool is a config
   edit anyone can do — no Go knowledge required, no rebuild needed for users
   running custom rule files.
2. **Staleness uses `git log`, not file mtime.** Editor activity, indexers,
   and Spotlight all bump mtime on files inside abandoned projects. Last commit
   date is a much better signal of "this project is actually dead."
3. **Every apply writes a JSONL audit log.** Months later you can answer
   "what did reclaim delete on October 14?" by `cat ~/.reclaim/logs/2026-10-14.jsonl`.
4. **Three-tier safety model exposed in the TUI.** `safe` / `confirm` /
   `dangerous` is part of the data model, not a runtime heuristic.
5. **Single static binary.** No runtime to install for end users.

The README's framing should reflect these specifics rather than imply this is
a greenfield space.

---

## 3. Proposed catalog additions (drop-in for `internal/rules/catalog.yaml`)

These are organized into the same categories already in reclaim. Each entry follows
reclaim's existing rule format. Where a rule needs path-pattern scanning (per-project),
the project-walker subsystem from §4 is required first.

### 3.1 New global cache rules (no engine changes required)

```yaml
# === Build caches: additions ===

- id: gradle-native-cache
  category: build-caches
  paths:
    - ~/.gradle/native
  safety: safe
  regenerates: true
  description: Gradle native compilation cache
  detect: dir-exists

- id: ivy2-cache
  category: build-caches
  paths:
    - ~/.ivy2/cache
  safety: safe
  regenerates: true
  description: Apache Ivy / SBT dependency cache
  detect: dir-exists

- id: sbt-boot
  category: build-caches
  paths:
    - ~/.sbt/boot
  safety: safe
  regenerates: true
  description: SBT (Scala) bootstrap launchers
  detect: dir-exists

- id: go-module-cache
  category: build-caches
  paths:
    - ~/go/pkg/mod
  safety: confirm
  regenerates: true
  description: |
    Go module cache. Large but re-downloaded on next `go build`.
    Confirm because re-download can be slow for many modules.
  detect: dir-exists

- id: bazel-cache
  category: build-caches
  paths:
    - ~/.cache/bazel
  safety: safe
  regenerates: true
  description: Bazel build cache
  detect: dir-exists

- id: turbo-cache
  category: build-caches
  paths:
    - ~/.cache/turbo
  safety: safe
  regenerates: true
  description: Turborepo remote cache directory
  detect: dir-exists

# === Package manager caches: JS additions ===

- id: npm-logs
  category: package-caches
  paths:
    - ~/.npm/_logs
  safety: safe
  regenerates: false
  description: npm error/install logs
  detect: dir-exists

- id: yarn-berry-cache
  category: package-caches
  paths:
    - ~/.yarn/berry/cache
  safety: safe
  regenerates: true
  description: Yarn 2+ (Berry) global cache
  detect: dir-exists

- id: bun-install-cache
  category: package-caches
  paths:
    - ~/.bun/install/cache
  safety: safe
  regenerates: true
  description: Bun package install cache
  detect: dir-exists

- id: deno-cache
  category: package-caches
  paths:
    - ~/.deno/deps
    - ~/.deno/gen
  safety: safe
  regenerates: true
  description: Deno module and codegen cache
  detect: dir-exists

# === Package manager caches: Python additions ===

- id: pipenv-cache
  category: package-caches
  paths:
    - ~/.cache/pipenv
  safety: safe
  regenerates: true
  description: Pipenv dependency cache
  detect: dir-exists

- id: conda-pkgs
  category: package-caches
  paths:
    - ~/.conda/pkgs
    - ~/anaconda3/pkgs
    - ~/miniconda3/pkgs
  safety: confirm
  regenerates: true
  description: Conda package tarball cache (re-download required after delete)
  detect: dir-exists

# === Package manager caches: Ruby additions ===

- id: gem-cache
  category: package-caches
  paths:
    - ~/.gem/cache
  safety: safe
  regenerates: true
  description: RubyGems download cache
  detect: dir-exists

# === Package manager caches: PHP / .NET / Carthage ===

- id: composer-cache
  category: package-caches
  paths:
    - ~/.composer/cache
  safety: safe
  regenerates: true
  description: Composer (PHP) package cache
  detect: dir-exists

- id: nuget-packages
  category: package-caches
  paths:
    - ~/.nuget/packages
  safety: confirm
  regenerates: true
  description: NuGet (.NET) global package cache
  detect: dir-exists

- id: carthage-cache
  category: package-caches
  paths:
    - ~/Library/Caches/org.carthage.CarthageKit
  safety: safe
  regenerates: true
  description: Carthage download cache
  detect: dir-exists

# === Xcode additions ===

- id: xcode-previews
  category: ide-caches
  paths:
    - ~/Library/Developer/Xcode/UserData/Previews
  safety: safe
  regenerates: true
  description: Xcode SwiftUI Preview cache
  detect: dir-exists

- id: xcode-ios-device-logs
  category: ide-caches
  paths:
    - ~/Library/Developer/Xcode/iOS Device Logs
  safety: safe
  regenerates: true
  description: Logs pulled from connected iOS devices
  detect: dir-exists

- id: coresimulator-caches
  category: ide-caches
  paths:
    - ~/Library/Developer/CoreSimulator/Caches
  safety: safe
  regenerates: true
  description: CoreSimulator temporary caches
  detect: dir-exists

# === IDE / editor additions ===

- id: cursor-caches
  category: ide-caches
  paths:
    - ~/Library/Application Support/Cursor/Cache
    - ~/Library/Application Support/Cursor/CachedData
    - ~/Library/Application Support/Cursor/Code Cache
    - ~/Library/Application Support/Cursor/GPUCache
    - ~/Library/Application Support/Cursor/logs
  safety: safe
  regenerates: true
  description: Cursor IDE caches and logs (NOT extensions or settings)
  detect: dir-exists

- id: windsurf-caches
  category: ide-caches
  paths:
    - ~/Library/Application Support/Windsurf/Cache
    - ~/Library/Application Support/Windsurf/CachedData
    - ~/Library/Application Support/Windsurf/Code Cache
    - ~/Library/Application Support/Windsurf/GPUCache
  safety: safe
  regenerates: true
  description: Windsurf IDE caches
  detect: dir-exists

- id: zed-cache
  category: ide-caches
  paths:
    - ~/Library/Caches/dev.zed.Zed
  safety: safe
  regenerates: true
  description: Zed editor cache
  detect: dir-exists

- id: vscode-logs
  category: ide-caches
  paths:
    - ~/Library/Application Support/Code/logs
  safety: safe
  regenerates: false
  description: VS Code logs
  detect: dir-exists

- id: jetbrains-logs
  category: ide-caches
  paths:
    - ~/Library/Logs/JetBrains
  safety: safe
  regenerates: false
  description: JetBrains IDE logs
  detect: dir-exists

# === Containers / VMs ===

- id: docker-buildx-cache
  category: containers
  paths:
    - ~/.docker/buildx/cache
  safety: safe
  regenerates: true
  description: Docker BuildKit cache (filesystem-side; daemon prune is separate)
  detect: dir-exists

- id: colima-vm
  category: containers
  paths:
    - ~/.colima
  safety: confirm
  regenerates: true
  description: |
    Colima Lima VM disk images. Stopping Colima first is recommended.
    Each VM can be 5–20 GB.
  detect: dir-exists

- id: lima-vm
  category: containers
  paths:
    - ~/.lima
  safety: confirm
  regenerates: true
  description: Lima VM disk images
  detect: dir-exists

# === Testing tools ===

- id: playwright-browsers
  category: testing-caches
  paths:
    - ~/Library/Caches/ms-playwright
    - ~/.cache/ms-playwright
  safety: safe
  regenerates: true
  description: Playwright downloaded browser binaries
  detect: dir-exists

- id: puppeteer-cache
  category: testing-caches
  paths:
    - ~/.cache/puppeteer
  safety: safe
  regenerates: true
  description: Puppeteer downloaded Chromium
  detect: dir-exists

- id: cypress-cache
  category: testing-caches
  paths:
    - ~/Library/Caches/Cypress
  safety: safe
  regenerates: true
  description: Cypress downloaded test runner binaries
  detect: dir-exists

- id: prisma-cache
  category: testing-caches
  paths:
    - ~/.cache/prisma
  safety: safe
  regenerates: true
  description: Prisma engine binary cache
  detect: dir-exists

# === Android additions ===

- id: android-avd-unused
  category: ide-caches
  paths:
    - ~/.android/avd
  safety: confirm
  regenerates: true
  description: |
    Android Virtual Devices. Confirm because each AVD = 5–10 GB and
    deleting wipes app state inside the emulator.
  detect: dir-exists

- id: android-build-cache
  category: build-caches
  paths:
    - ~/.android/build-cache
  safety: safe
  regenerates: true
  description: Android build-cache directory
  detect: dir-exists

- id: android-system-images-stale
  category: ide-caches
  paths:
    - ~/Library/Android/sdk/system-images
  safety: confirm
  regenerates: true
  description: |
    Android SDK system images for emulators.
    Confirm — re-downloading API levels takes time.
  detect: dir-exists

# === Flutter additions ===

- id: dart-server-cache
  category: build-caches
  paths:
    - ~/.dartServer
  safety: safe
  regenerates: true
  description: Dart Analysis Server cache
  detect: dir-exists

- id: fvm-versions-stale
  category: ide-caches
  paths:
    - ~/fvm/versions
  safety: confirm
  regenerates: true
  description: |
    fvm-managed Flutter SDK versions. Confirm because re-downloading
    a Flutter SDK = 1–2 GB and several minutes per version.
  detect: dir-exists

# === Version manager stale entries ===

- id: nvm-cache
  category: package-caches
  paths:
    - ~/.nvm/.cache
  safety: safe
  regenerates: true
  description: nvm downloaded node tarball cache
  detect: dir-exists

# === Cloud CLI caches ===

- id: aws-cli-cache
  category: package-caches
  paths:
    - ~/.aws/cli/cache
  safety: safe
  regenerates: true
  description: AWS CLI SSO/role cache
  detect: dir-exists

- id: terraform-plugin-cache
  category: package-caches
  paths:
    - ~/.terraform.d/plugin-cache
  safety: safe
  regenerates: true
  description: Terraform shared provider plugin cache
  detect: dir-exists

# === System-level developer wins ===

- id: ipsw-firmware
  category: ide-caches
  paths:
    - ~/Library/iTunes/iPhone Software Updates
    - ~/Library/iTunes/iPad Software Updates
  safety: confirm
  regenerates: true
  description: Old iOS/iPadOS firmware downloaded by iTunes/Finder
  detect: dir-exists
```

### 3.2 New project-walker rules (require §4.1 walker)

These rules need path-pattern scanning under user-configured project roots
(`~/Projects`, `~/dev`, etc.). They use `detect: scan` + `scan_for`. Each entry
*must* be gated behind reclaim's `git log`-based staleness check before listing.

```yaml
# === Project artifacts (activity-aware) ===

- id: project-node-modules
  category: project-caches
  scan_for: node_modules
  safety: safe
  regenerates: true
  description: node_modules in stale projects (last commit older than threshold)
  detect: scan
  staleness: git-log
  warn_if_running:
    - Spotify
    - Discord
    - Slack
    - VSCode

- id: project-rust-target
  category: project-caches
  scan_for: target
  marker_file: Cargo.toml
  safety: safe
  regenerates: true
  description: Rust target/ directory in projects with Cargo.toml
  detect: scan
  staleness: git-log

- id: project-python-venv
  category: project-caches
  scan_for: ["venv", ".venv", "env"]
  marker_file: ["requirements.txt", "pyproject.toml", "setup.py"]
  safety: confirm
  regenerates: true
  description: Python virtual environments. Confirm — recreating takes time.
  detect: scan
  staleness: git-log

- id: project-pycache
  category: project-caches
  scan_for: ["__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache"]
  safety: safe
  regenerates: true
  description: Python tooling caches
  detect: scan

- id: project-dotnet-bin-obj
  category: project-caches
  scan_for: ["bin", "obj"]
  marker_file: ["*.csproj", "*.fsproj", "*.sln"]
  safety: safe
  regenerates: true
  description: .NET build outputs
  detect: scan
  staleness: git-log

- id: project-unity
  category: project-caches
  scan_for: ["Library", "Temp", "Logs", "obj"]
  marker_file: "ProjectSettings/ProjectVersion.txt"
  safety: confirm
  regenerates: true
  description: Unity project caches. Confirm — Library can be 5+ GB to rebuild.
  detect: scan
  staleness: git-log

- id: project-unreal
  category: project-caches
  scan_for: ["Binaries", "Intermediate", "DerivedDataCache", "Saved"]
  marker_file: "*.uproject"
  safety: confirm
  regenerates: true
  description: Unreal Engine project caches
  detect: scan
  staleness: git-log

- id: project-godot
  category: project-caches
  scan_for: ".godot"
  marker_file: "project.godot"
  safety: safe
  regenerates: true
  description: Godot 4.x project import cache
  detect: scan
  staleness: git-log

- id: project-elixir
  category: project-caches
  scan_for: ["_build", "deps"]
  marker_file: "mix.exs"
  safety: safe
  regenerates: true
  description: Elixir/Mix build and dependency dirs
  detect: scan
  staleness: git-log

- id: project-haskell
  category: project-caches
  scan_for: [".stack-work", "dist-newstyle"]
  marker_file: ["*.cabal", "stack.yaml", "package.yaml"]
  safety: safe
  regenerates: true
  description: Stack/Cabal build outputs
  detect: scan
  staleness: git-log

- id: project-zig
  category: project-caches
  scan_for: ["zig-cache", "zig-out"]
  marker_file: "build.zig"
  safety: safe
  regenerates: true
  description: Zig build outputs
  detect: scan
  staleness: git-log

- id: project-terraform
  category: project-caches
  scan_for: ".terraform"
  marker_file: ["*.tf"]
  safety: safe
  regenerates: true
  description: Terraform per-workspace plugins/modules cache
  detect: scan
  staleness: git-log

- id: project-frontend-build
  category: project-caches
  scan_for: [".next", ".nuxt", ".svelte-kit", ".astro", ".cache", "dist", "build", "out"]
  marker_file: "package.json"
  safety: safe
  regenerates: true
  description: Common JS framework build outputs
  detect: scan
  staleness: git-log
```

---

## 4. Required engine changes for v0.2

The catalog above implies engine work beyond just adding YAML entries. Listed
in priority order:

### 4.1 Project walker subsystem (blocker for §3.2)

- New scanner that walks user-configured roots (`~/Projects`, `~/GitHub`, `~/dev`,
  configurable via `~/.reclaim/config.yaml`).
- For each scanned dir, detect project type by `marker_file`.
- For each candidate, run `git log -1 --format=%cI` to get last commit date;
  fall back to mtime if not a git repo.
- Apply `--stale 30d`/`60d`/`90d` filter.
- App-embedded `node_modules` warning: skip / warn for paths inside
  `/Applications/*.app/`.

### 4.2 Process-running detection

- Currently reclaim has "won't suggest cleaning Docker if it's running". Extend
  to all rules with `warn_if_running: [...]` field.
- Implementation = `pgrep -f` on macOS / Linux, `lsof` to confirm path is held.

### 4.3 Shell-action rules

- New `action: shell` + `action_command:` field for rules that *don't* delete
  paths but run a tool's own cleanup (e.g. `docker system prune`, `brew cleanup -s`).
- Captures stdout for the audit log; prefix output in TUI.

### 4.4 Conditional detection

- New `detect: command` + `detect_command:` to gate rules on a tool being
  available/running (e.g. only show docker-prune if `docker info` succeeds).

### 4.5 Safety: lock-state checks for simulator volumes

- For CoreSimulator volumes, must read each subdir and skip ones marked
  `IN_USE`. Otherwise wiping risks active simulators.

### 4.6 Disk delta in summary

- Already on roadmap. Compute `df -k /` before and after apply, surface
  "Freed X GB | Free disk: Y GB → Z GB" in the summary line.

---

## 5. Quick-win priority order for v0.2

If shipping incrementally, sequence by impact-per-engineering-hour:

**Sprint 1 (catalog-only — 1 PR):**
- All §3.1 rules that already match reclaim's existing engine (`detect: dir-exists`)
- ~35 new rules, no engine changes
- Biggest single user wins: `bun-install-cache`, `playwright-browsers`,
  `cursor-caches`, `coresimulator-caches`, `gem-cache`, `nuget-packages`,
  `carthage-cache`, `fvm-versions-stale`

**Sprint 2 (engine + catalog):**
- Add `detect: command` + `action: shell` rule support
- Add `warn_if_running` field with process detection
- Catalog: `docker-buildx-cache`, `docker-system-prune`, plus warn-if-running on
  the existing Docker rule

**Sprint 3 (project walker — biggest delta):**
- Implement walker with `marker_file` + `git log` staleness
- Ship §3.2 rules
- Add `--stale 30d` flag

**Sprint 4 (polish):**
- Disk delta in summary
- App-embedded `node_modules` warning
- Lock-state check for simulator volumes
- README rewrite per §2

---

## 6. Anti-patterns to avoid

1. **Don't clean browser profiles, cookies, or history.** Cache subdirectories
   only. Even though Chrome's `Default/Cache` is safe, `Default/Cookies` and
   `Default/Login Data` are not.
2. **Don't auto-delete CoreSimulator Devices wholesale.** Skip IN_USE entries.
   Wiping all simulators wipes the data inside test apps.
3. **Don't pipe `rm -rf` directly.** All deletions must go through
   `engine.Apply` (which calls `CheckPathSafe` first). Audit any new code
   path that touches the filesystem.
4. **Don't claim space savings you can't verify.** Report actual delta from
   `df`, not a sum of estimated sizes (estimates lie when paths overlap or
   contain symlinks).
5. **Don't surface AVDs / `~/.pyenv/versions` / Conda envs as `safe`.** They
   look like caches but contain user-installed state. `confirm` tier minimum.

---

## 7. Out of scope (deliberate non-goals)

- **App uninstaller.** Different problem domain — needs bundle ID resolution,
  launchd plist cleanup, login items. Stay focused.
- **Disk visualizer.** Tangential; tools like `ncdu` already exist.
- **System optimization** (rebuilding launchd, spotlight reindex, etc.).
  Risk/reward is bad and the audience is different.
- **Duplicate file finder.** Different problem.
- **Native GUI.** TUI is enough; native GUI is a 10x effort multiplier for
  marginal gain in our target audience (CLI-comfortable developers).
