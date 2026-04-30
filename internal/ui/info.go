package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// PrintAbout writes the static help / about text to stdout.
// Called from main when the user picks "About / help" from the welcome menu.
func PrintAbout(version string) {
	fmt.Println()
	fmt.Println(bannerStyle.Render("🧹 reclaim — about"))
	fmt.Println()
	fmt.Println("Version:    " + version)
	fmt.Println("Homepage:   https://github.com/ImadRashid/reclaim")
	fmt.Println("Docs:       https://github.com/ImadRashid/reclaim/blob/main/README.md")
	fmt.Println("Roadmap:    https://github.com/ImadRashid/reclaim/blob/main/docs/ROADMAP.md")
	fmt.Println("License:    MIT")
	fmt.Println()
	fmt.Println("Safety levels:")
	fmt.Println("    " + safeStyle.Render("✓ safe") + "      Regenerates automatically — deletion has no impact")
	fmt.Println("    " + confirmStyle.Render("⚠ confirm") + "   Should be reviewed before deleting")
	fmt.Println("    " + dangerStyle.Render("✗ dangerous") + " Skipped by default — handle manually")
	fmt.Println()
	fmt.Println("TUI keys (during the picker):")
	fmt.Println("    ↑/↓ / j/k     navigate")
	fmt.Println("    space         toggle item or select/deselect a whole category")
	fmt.Println("    →/l, ←/h      expand / collapse category")
	fmt.Println("    a / n         select all non-dangerous / deselect all")
	fmt.Println("    enter         apply selected (with confirmation)")
	fmt.Println("    q / esc       quit without changes")
	fmt.Println()
	fmt.Println("Run history is written to ~/.reclaim/logs/YYYY-MM-DD.jsonl after every apply.")
	fmt.Println()
	fmt.Println("Skip the welcome menu with flags:")
	fmt.Println("    reclaim --plain               # plain-text scan only")
	fmt.Println("    reclaim --apply               # scan + delete safe items, single y/N prompt")
	fmt.Println("    reclaim --apply -c <id>       # restrict to one category")
	fmt.Println("    reclaim --pick                # go straight to the picker (skip welcome)")
	fmt.Println("    reclaim --help                # full reference")
	fmt.Println()
}

// PrintLastLog reads the latest JSONL entry from ~/.reclaim/logs/ and
// prints a human-readable summary. It's resilient to a missing log dir.
func PrintLastLog() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Could not locate home directory.")
		return
	}
	logDir := filepath.Join(home, ".reclaim", "logs")

	entries, err := os.ReadDir(logDir)
	if err != nil || len(entries) == 0 {
		fmt.Println()
		fmt.Println(bannerStyle.Render("🧹 reclaim — last cleanup log"))
		fmt.Println()
		fmt.Println("No cleanup runs recorded yet.")
		fmt.Println("Logs will appear at " + abbreviatePath(logDir) + " after your first cleanup.")
		fmt.Println()
		return
	}

	// Most recent log file by name (YYYY-MM-DD.jsonl sorts naturally).
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})
	latest := filepath.Join(logDir, entries[0].Name())

	type entry struct {
		Time    time.Time `json:"time"`
		RuleID  string    `json:"rule"`
		Path    string    `json:"path"`
		Freed   int64     `json:"freed_bytes"`
		Failed  bool      `json:"failed"`
		Error   string    `json:"error"`
		Skipped bool      `json:"skipped"`
		Reason  string    `json:"reason"`
	}

	f, err := os.Open(latest)
	if err != nil {
		fmt.Println("Could not read", latest, ":", err)
		return
	}
	defer f.Close()

	var rows []entry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var e entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			rows = append(rows, e)
		}
	}

	if len(rows) == 0 {
		fmt.Println("Latest log is empty.")
		return
	}

	// Summarize the most recent contiguous run (entries within 60s of each other).
	cutoff := rows[len(rows)-1].Time.Add(-60 * time.Second)
	var totalFreed int64
	var deleted, skipped, failed int
	for _, e := range rows {
		if e.Time.Before(cutoff) {
			continue
		}
		switch {
		case e.Failed:
			failed++
		case e.Skipped:
			skipped++
		default:
			deleted++
			totalFreed += e.Freed
		}
	}

	fmt.Println()
	fmt.Println(bannerStyle.Render("🧹 reclaim — last cleanup log"))
	fmt.Println()
	fmt.Println("Log file:   " + abbreviatePath(latest))
	fmt.Println("Last run:   " + rows[len(rows)-1].Time.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Printf("  %s deleted, %s skipped, %s failed — %s freed\n",
		safeStyle.Render(fmt.Sprintf("%d items", deleted)),
		confirmStyle.Render(fmt.Sprintf("%d", skipped)),
		dangerStyle.Render(fmt.Sprintf("%d", failed)),
		formatSize(totalFreed))
	fmt.Println()

	// Show the last 10 deletions for context.
	fmt.Println("Recent deletions:")
	shown := 0
	for i := len(rows) - 1; i >= 0 && shown < 10; i-- {
		e := rows[i]
		if e.Failed || e.Skipped {
			continue
		}
		fmt.Printf("  · %-50s  %s\n", abbreviatePath(truncatePath(e.Path, 50)), formatSize(e.Freed))
		shown++
	}
	fmt.Println()
}

func truncatePath(p string, max int) string {
	if len(p) <= max {
		return p
	}
	return "…" + p[len(p)-max+1:]
}

// SuppressUnused keeps the linter quiet about strings used only via templates.
var _ = strings.Builder{}
