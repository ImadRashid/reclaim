// Package update implements the once-daily background update check.
//
// This is the ONLY place in reclaim that touches the network. It asks the
// GitHub releases API for the latest tag at most once every 24 hours
// (cached in ~/.reclaim/update-check.json), never blocks or delays the
// UI, and never downloads or replaces anything — it only prints a one-line
// notice telling the user how to upgrade. Opt out with
// RECLAIM_NO_UPDATE_CHECK=1 or --no-update-check.
package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	releaseAPI = "https://api.github.com/repos/ImadRashid/reclaim/releases/latest"
	cacheTTL   = 24 * time.Hour
)

// Start launches the check in the background and returns a function to call
// once, at exit. That function returns a one-line upgrade notice, or "" if
// the user is current, the check is disabled, or the result isn't in yet
// (it never waits — a slow network means no notice, not a slow exit).
func Start(current string) func() string {
	if os.Getenv("RECLAIM_NO_UPDATE_CHECK") != "" {
		return func() string { return "" }
	}
	ch := make(chan string, 1)
	go func() { ch <- latestVersion() }()
	return func() string {
		select {
		case latest := <-ch:
			if newer(latest, current) {
				return notice(latest, current)
			}
		default:
		}
		return ""
	}
}

type cache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

func cachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".reclaim", "update-check.json")
}

// latestVersion returns the newest release tag (e.g. "v0.2.0"), from the
// cache when it's fresh, otherwise from the GitHub API. Returns "" on any
// failure — an update notice is never worth surfacing an error for.
func latestVersion() string {
	p := cachePath()
	if p != "" {
		if b, err := os.ReadFile(p); err == nil {
			var c cache
			if json.Unmarshal(b, &c) == nil && time.Since(c.CheckedAt) < cacheTTL {
				return c.Latest
			}
		}
	}

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("GET", releaseAPI, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "reclaim-update-check")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	if p != "" && release.TagName != "" {
		if b, err := json.Marshal(cache{CheckedAt: time.Now(), Latest: release.TagName}); err == nil {
			_ = os.MkdirAll(filepath.Dir(p), 0o755)
			_ = os.WriteFile(p, b, 0o644)
		}
	}
	return release.TagName
}

// newer reports whether version a is strictly newer than b. Versions are
// "vX.Y.Z" or "X.Y.Z"; anything unparsable compares as not-newer so dev
// builds and API garbage never trigger a notice.
func newer(a, b string) bool {
	av, aok := parse(a)
	bv, bok := parse(b)
	if !aok || !bok {
		return false
	}
	for i := 0; i < 3; i++ {
		if av[i] != bv[i] {
			return av[i] > bv[i]
		}
	}
	return false
}

func parse(v string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// notice formats the one-line upgrade hint, matched to how this binary was
// installed: Homebrew-managed binaries should be upgraded through brew,
// anything else via the releases page.
func notice(latest, current string) string {
	how := "download: https://github.com/ImadRashid/reclaim/releases/latest"
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		if strings.Contains(exe, "/Cellar/") || strings.Contains(exe, "/homebrew/") || strings.Contains(exe, "/linuxbrew/") {
			how = "run: brew upgrade reclaim"
		}
	}
	return fmt.Sprintf("↑ reclaim %s is available (you have %s) — %s",
		strings.TrimPrefix(latest, "v"), strings.TrimPrefix(current, "v"), how)
}
