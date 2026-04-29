# Roadmap and gap analysis

This document captures what `reclaim` covers today, where the gaps are, and
what to build next. It's written for contributors picking up an issue and for
the maintainer planning releases.

If you want to fill one of these gaps, open an issue first (use the
[New cleanup rule](../.github/ISSUE_TEMPLATE/new_rule.yml) template) so we can
discuss the safety tier and rule shape before you write code.

---

## Where we stand today (v0.1.0)

### Covered well

| Stack / area | How |
|--------------|-----|
| **Node / React / Next** | npm, yarn, pnpm caches; per-project `node_modules`, `.next`, `.open-next` |
| **Flutter / Dart** | `~/.pub-cache`; per-project `.dart_tool` |
| **iOS / Xcode (basic)** | DerivedData, iOS/watchOS/tvOS DeviceSupport, per-project `Pods`, `xcrun simctl delete unavailable` |
| **Android (basic)** | `~/.gradle/{caches,wrapper,daemon}`; per-project `android/.gradle` |
| **Rust** | Per-project `target/`, `~/.cargo/{registry,git}`, `~/.rustup` |
| **Python (basic)** | `__pycache__`, `.venv`/`venv`, pip / uv / poetry caches |
| **Go** | `~/Library/Caches/go-build` |
| **Java/JVM** | Gradle, Maven (`~/.m2/repository`) |
| **Containers** | Docker Desktop VM disk |
| **Browsers** | Chrome, Arc caches (asset cache only — does not log out) |
| **AI / ML** | Hugging Face, Ollama, PyTorch caches |
| **System** | Trash, Homebrew, Puppeteer, Playwright, JetBrains caches |

The catalog is in [`internal/rules/catalog.yaml`](../internal/rules/catalog.yaml).

---

## Gap analysis by stack

Each item below is a candidate rule. Any gap with high real-world impact (≥1 GB
on a typical dev machine) is marked **★**. These should be prioritized first.

### Flutter / iOS specifics

| Gap | Path | Notes |
|-----|------|-------|
| **★ fvm SDK versions** | `~/fvm/versions/*` | Each Flutter SDK is 0.8–2.5 GB. Most users keep stale ones. Needs activity-aware deletion (don't remove the version pinned by an active project). |
| **★ Xcode simulator runtimes** | Managed via `xcrun simctl runtime list/delete` | A single Mac can carry 50+ GB across 5–7 iOS versions. Highest-leverage cleanup we don't fully cover. |
| Xcode ModuleCache / SymbolCache | `~/Library/Developer/Xcode/UserData/IB Support`, `~/Library/Developer/CoreSimulator/Caches`, `~/Library/Developer/Xcode/SymbolCache` | Smaller but reliably present. |
| CocoaPods Pods cache (vs source cache) | `~/Library/Caches/CocoaPods/Pods` | Different from the source cache we already cover. |
| Watchman state | `~/Library/LaunchAgents/...watchman*` | Common in RN/Flutter setups. |
| Simulator app data ("orphaned") | `~/Library/Developer/CoreSimulator/Devices/<id>/data/Containers/Data/Application/<uuid>/` for apps no longer installed | Hard to detect but enormous on long-lived sims. |

### React Native / Expo

| Gap | Path | Notes |
|-----|------|-------|
| **★ Metro bundler cache** | `~/.metro` | RN devs hit this hard; can be GBs. |
| Expo cache | `~/.expo` | Same. |
| Per-project `android/build` | `<project>/android/build` | Distinct from `.gradle`. |
| Per-project `ios/build` | `<project>/ios/build` | Distinct from `Pods`. |
| Hermes / Realm / RN native build artifacts | various | Often inside `node_modules/<pkg>/build`. |

### Node / JS ecosystem

| Gap | Path | Notes |
|-----|------|-------|
| **★ Bun cache** | `~/.bun/install/cache` | Mainstream now; not yet covered. |
| Deno cache | `~/Library/Caches/deno` | Growing user base. |
| **★ Turborepo `.turbo`** | per-project `.turbo` | Common in JS monorepos; can be hundreds of MB. |
| Vite cache | per-project `node_modules/.vite` | Buried inside `node_modules` but distinct artifact. |
| Webpack cache | per-project `node_modules/.cache/webpack` | Same. |
| Esbuild / SWC caches | per-project `node_modules/.cache/...` | Family of small caches. |
| Storybook | `node_modules/.cache/storybook` | |
| **★ Cypress** | `~/Library/Caches/Cypress` | Bins multiple versions; often >2 GB. |
| Playwright (browsers vs binaries) | `~/Library/Caches/ms-playwright`, `~/Library/Caches/ms-playwright-go` | Already partly covered, double-check parity. |

### Python ecosystem (where we're thinnest)

| Gap | Path | Notes |
|-----|------|-------|
| **★ conda environments** | `~/miniconda3/envs/*`, `~/anaconda3/envs/*`, `~/opt/anaconda3/envs/*` | Each env is GBs. Activity-aware deletion needed. |
| **★ pyenv versions** | `~/.pyenv/versions/*` | Like nvm; stale Python builds. |
| Jupyter checkpoints | per-project `.ipynb_checkpoints` | Spam in DS workflows. |
| DVC cache | per-project `.dvc/cache`, `~/.dvc/cache` | Data scientists. |
| MLflow / wandb runs | `~/.mlflow`, `~/.wandb`, per-project `mlruns/` | |
| Pre-commit cache | `~/.cache/pre-commit` | |
| **★ Hugging Face Hub cache** | `~/.cache/huggingface/hub` | Already in catalog as `huggingface-cache` but worth verifying nested paths are covered. |
| Pyright cache | `~/.cache/pyright` | |
| Ruff cache | `.ruff_cache` per project | |
| Mypy cache | `.mypy_cache` per project | |
| Pytest cache | `.pytest_cache` per project | |

### Other stacks worth covering

| Stack | What |
|-------|------|
| **Ruby** | `~/.rbenv/versions`, `~/.rvm`, `~/.bundle/cache`, per-project `vendor/bundle` |
| **PHP** | `~/.composer/cache`, per-project `vendor/` |
| **.NET** | **★ `~/.nuget/packages`** (often very large), per-project `bin/` and `obj/` |
| **Kubernetes** | `~/.kube/cache`, `~/.minikube/cache`, kind clusters |
| **Terraform** | per-project `.terraform/` (provider downloads — often hundreds of MB per project) |
| **Pulumi** | `~/.pulumi/plugins` |
| **Unity / Unreal** | per-project `Library/`, `Temp/`, `Saved/`; `~/Library/Unity/cache` |
| **JetBrains IDEs (per-version)** | `~/Library/Caches/JetBrains/<IDE>2024.x/` is versioned — old versions linger after upgrade |
| **VS Code logs** | `~/Library/Application Support/Code/logs` |
| **VS Code forks** (Cursor, Windsurf, Trae, etc.) | Each has its own `~/Library/Application Support/<name>/` |
| **AI coding assistants** | Cline / Cursor / Windsurf chat history & extension caches |

### Mac-specific opportunities

| Gap | Notes |
|-----|-------|
| **★ Per-version simulator runtime cleanup** | Wrap `xcrun simctl runtime list/delete` so users can deselect a single iOS version, not just `unavailable`. |
| Orphaned app data | `~/Library/Application Support/<app>` and `~/Library/Containers/<bundle>` for apps no longer installed. Surprisingly large; needs careful detection (which bundles are dead?). |
| Time Machine local snapshots | `tmutil listlocalsnapshots /` + `tmutil deletelocalsnapshots <date>`. macOS keeps these. |
| Mail downloads / attachments | `~/Library/Containers/com.apple.mail/Data/Library/Mail Downloads` |
| iCloud / Mobile Documents temp | `~/Library/Mobile Documents/.../tmp` |
| Photos analysis cache | `~/Pictures/Photos Library.photoslibrary/private/com.apple.PhotoAnalysis/...` (read-only — careful) |

---

## Cross-cutting features (not new rules)

These are improvements to the engine and TUI rather than new cleanup targets.

| Feature | Why it matters | Effort |
|---------|----------------|--------|
| **Per-project view** in TUI | Devs think in projects, not paths. Group hits by `git` repo root and let users say "free everything for this project". | Medium |
| **Free space delta in summary** | "Freed 12 GB. Disk: 73 GB → 85 GB free." Builds trust. | Easy |
| **Trash instead of hard delete (opt-in)** | macOS Trash via `osascript` lets users undo. Add a `--trash` flag. | Easy |
| **Profile presets** | `reclaim --preset flutter`, `--preset rn`, `--preset python-ds`. Newcomer onboarding. | Easy |
| **`reclaim doctor`** | After a big cleanup, run a tiny health check (`flutter doctor`, `xcodebuild -version`, etc.) to confirm nothing important was hit. Builds trust. | Medium |
| **Concurrent scanning** | Scanner is sequential (~5s); parallel would be ~1s. | Easy |
| **External rule files** | `~/.reclaim/rules.d/*.yaml` so users add rules without recompiling. | Medium |
| **Linux-aware paths** | The catalog is macOS-first. Add Linux-equivalent paths (XDG cache dirs). | Medium |
| **Activity-aware deletion for SDK managers** | fvm, nvm, pyenv, rbenv: only suggest deleting versions not pinned by any active project. | Medium |
| **Code signing + notarization** | `Developer ID` certificate ($99/yr) + `notarytool` removes the macOS "unidentified developer" warning. | Medium |

---

## Tentative release plan

These are **directional**, not commitments. Each release ships when it's ready.

### v0.1.0 — first cut (current)

- Interactive TUI, plain mode, non-interactive apply
- 38 rules across 8 categories
- Activity-aware project staleness via git
- Path-safety guard (`CheckPathSafe`)
- Audit log to `~/.reclaim/logs/`
- GitHub Actions cross-compile pipeline
- Homebrew formula template

### v0.2.0 — close the dev-stack gap

The high-impact rules above marked **★**:

- fvm SDK versions (activity-aware)
- Xcode per-version simulator runtimes
- Bun cache
- Turborepo `.turbo`
- Cypress
- Metro / Expo
- conda environments
- pyenv versions
- NuGet packages
- Cross-cutting: free space delta in summary

This is the release that makes `reclaim` a serious tool for Flutter / RN /
Node / Python devs.

### v0.3.0 — productization

- Per-project view in TUI
- `--trash` mode (opt-in soft delete)
- `--preset <name>` profiles
- `reclaim doctor`
- External rule files (`~/.reclaim/rules.d/*.yaml`)

### v0.4.0+ — optional / community-driven

- Linux rule pack
- Ruby, PHP, .NET, Unity, Terraform, K8s
- VS Code fork detection (Cursor, Windsurf, etc.)
- Orphaned app data detector
- Time Machine local snapshots
- Code signing + notarization

---

## How to pick up an item

1. **Open an issue** using the
   [New cleanup rule](../.github/ISSUE_TEMPLATE/new_rule.yml) template (for
   rules) or [Feature request](../.github/ISSUE_TEMPLATE/feature_request.yml)
   (for engine/TUI changes). Reference the section here so we know it's tied
   to the roadmap.
2. **Get a thumbs-up** from a maintainer on the safety tier and approach.
3. **Implement** following [CONTRIBUTING.md](../CONTRIBUTING.md). For rules
   this is a YAML edit + verification output. For features it's Go +
   tests.
4. **Open a PR** using the template.

The roadmap is intentionally ambitious. We don't need every item — community
PRs decide which ones land first.
