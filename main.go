// reclaim — a developer-aware Mac cleaner CLI.
//
// Usage:
//
//	reclaim                        # interactive TUI (scan + pick + apply)
//	reclaim --plain                # plain-text scan (no TUI)
//	reclaim --apply                # apply directly without TUI (uses safe defaults)
//	reclaim --apply -c <category>  # restrict apply to one category
//	reclaim --version
//	reclaim --help
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ImadRashid/reclaim/internal/engine"
	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
	"github.com/ImadRashid/reclaim/internal/ui"
)

const version = "0.1.0"

type opts struct {
	apply      bool
	plain      bool
	category   string
	wantHelp   bool
	wantVer    bool
}

func main() {
	o := parseArgs(os.Args[1:])
	if o.wantHelp {
		printHelp()
		return
	}
	if o.wantVer {
		fmt.Printf("reclaim %s\n", version)
		return
	}

	cat, err := rules.Load()
	if err != nil {
		die("load rules: %v", err)
	}
	ruleByID := engine.RuleMap(cat)

	hits := runScan(cat)

	// Filter by category if requested.
	if o.category != "" {
		hits = filterByCategory(hits, ruleByID, o.category)
	}

	if len(hits) == 0 {
		fmt.Println("Nothing to clean. Your system is tidy. ✨")
		return
	}

	// Mode selection:
	//   --plain      => text report only
	//   --apply      => non-interactive apply (uses safe defaults: safe + confirm)
	//   default      => TUI
	if o.plain {
		printReport(cat, hits, ruleByID)
		fmt.Println("\nRun without --plain for the interactive picker, or --apply to delete safely.")
		return
	}

	if o.apply {
		// Non-interactive apply: take all safe + confirm hits.
		runNonInteractiveApply(hits, ruleByID)
		return
	}

	// Default: interactive TUI.
	model := ui.New(cat, hits)
	final, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		die("tui: %v", err)
	}
	uiModel, ok := final.(ui.Model)
	if !ok || !uiModel.Confirmed() {
		fmt.Println("Cancelled — nothing was deleted.")
		return
	}
	selected := uiModel.Selected()
	if len(selected) == 0 {
		fmt.Println("No selections — nothing was deleted.")
		return
	}
	applyAndReport(selected, ruleByID)
}

func parseArgs(args []string) opts {
	var o opts
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--apply", "-a":
			o.apply = true
		case "--plain", "-p":
			o.plain = true
		case "--category", "-c":
			if i+1 < len(args) {
				o.category = args[i+1]
				i++
			}
		case "--version", "-v":
			o.wantVer = true
		case "--help", "-h":
			o.wantHelp = true
		}
	}
	return o
}

func runScan(cat *rules.Catalog) []scanner.Hit {
	fmt.Printf("\033[1m🧹 reclaim v%s\033[0m — scanning…\n", version)
	hits := scanner.Scan(context.Background(), cat, func(id string) {
		fmt.Printf("\r  %-50s", id)
	})
	fmt.Print("\r" + strings.Repeat(" ", 60) + "\r")
	return hits
}

func filterByCategory(hits []scanner.Hit, ruleByID map[string]rules.Rule, cat string) []scanner.Hit {
	var out []scanner.Hit
	for _, h := range hits {
		if ruleByID[h.RuleID].Category == cat {
			out = append(out, h)
		}
	}
	return out
}

// runNonInteractiveApply deletes safe + confirm hits after a single y/N prompt.
func runNonInteractiveApply(hits []scanner.Hit, ruleByID map[string]rules.Rule) {
	var toClean []scanner.Hit
	var total int64
	for _, h := range hits {
		r := ruleByID[h.RuleID]
		if r.Safety == "dangerous" {
			continue
		}
		toClean = append(toClean, h)
		total += h.Size
	}
	if len(toClean) == 0 {
		fmt.Println("Nothing in scope.")
		return
	}
	fmt.Printf("\033[1;33mAbout to free %s across %d items.\033[0m Continue? [y/N] ", formatSize(total), len(toClean))
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
		fmt.Println("Cancelled.")
		return
	}
	applyAndReport(toClean, ruleByID)
}

func applyAndReport(toClean []scanner.Hit, ruleByID map[string]rules.Rule) {
	results := engine.Apply(toClean, ruleByID, func(h scanner.Hit) {
		fmt.Printf("  cleaning %s …\n", abbreviatePath(h.Path))
	})

	logPath, _ := engine.LogRun(results)

	var freed int64
	var failed int
	for _, r := range results {
		switch {
		case r.Err != nil:
			failed++
			fmt.Printf("  \033[31m✗ %s: %v\033[0m\n", abbreviatePath(r.Path), r.Err)
		case r.Skipped:
			fmt.Printf("  \033[33m· skipped %s — %s\033[0m\n", abbreviatePath(r.Path), r.Reason)
		default:
			freed += r.Freed
		}
	}
	fmt.Println()
	fmt.Printf("\033[1;32m✓ Freed %s\033[0m", formatSize(freed))
	if failed > 0 {
		fmt.Printf(" (\033[31m%d failed\033[0m)", failed)
	}
	fmt.Println()
	if logPath != "" {
		fmt.Printf("Log: %s\n", abbreviatePath(logPath))
	}
}

func printReport(cat *rules.Catalog, hits []scanner.Hit, ruleByID map[string]rules.Rule) {
	type group struct {
		category rules.Category
		ruleHits map[string][]scanner.Hit
	}
	groups := map[string]*group{}
	for _, c := range cat.Categories {
		groups[c.ID] = &group{category: c, ruleHits: map[string][]scanner.Hit{}}
	}
	for _, h := range hits {
		r := ruleByID[h.RuleID]
		g, ok := groups[r.Category]
		if !ok {
			continue
		}
		g.ruleHits[r.ID] = append(g.ruleHits[r.ID], h)
	}

	var grandTotal int64
	for _, c := range cat.Categories {
		g := groups[c.ID]
		if len(g.ruleHits) == 0 {
			continue
		}
		var catTotal int64
		for _, hs := range g.ruleHits {
			for _, h := range hs {
				catTotal += h.Size
			}
		}
		grandTotal += catTotal
		fmt.Printf("\033[1;36m▸ %s — %s\033[0m\n", g.category.Name, formatSize(catTotal))
		ruleIDs := make([]string, 0, len(g.ruleHits))
		for rid := range g.ruleHits {
			ruleIDs = append(ruleIDs, rid)
		}
		sort.Slice(ruleIDs, func(i, j int) bool {
			return totalSize(g.ruleHits[ruleIDs[i]]) > totalSize(g.ruleHits[ruleIDs[j]])
		})
		for _, rid := range ruleIDs {
			r := ruleByID[rid]
			hs := g.ruleHits[rid]
			marker := safetyMarker(r.Safety)
			fmt.Printf("    %s  %-32s  %10s  %s\n", marker, rid, formatSize(totalSize(hs)), shortDesc(r.Description))
			limit := 3
			if len(hs) <= limit {
				for _, h := range hs {
					fmt.Printf("        · %s (%s)\n", abbreviatePath(h.Path), formatSize(h.Size))
				}
			} else {
				for _, h := range hs[:limit] {
					fmt.Printf("        · %s (%s)\n", abbreviatePath(h.Path), formatSize(h.Size))
				}
				fmt.Printf("        · …and %d more\n", len(hs)-limit)
			}
		}
		fmt.Println()
	}
	fmt.Printf("\033[1mTotal cleanable: %s across %d items\033[0m\n", formatSize(grandTotal), len(hits))
	fmt.Println("Legend:  ✓ safe   ⚠ confirm   ✗ dangerous")
}

func totalSize(hs []scanner.Hit) int64 {
	var t int64
	for _, h := range hs {
		t += h.Size
	}
	return t
}

func safetyMarker(s string) string {
	switch s {
	case "safe":
		return "\033[32m✓\033[0m"
	case "confirm":
		return "\033[33m⚠\033[0m"
	case "dangerous":
		return "\033[31m✗\033[0m"
	}
	return " "
}

func formatSize(n int64) string {
	const (
		kb = 1 << 10
		mb = 1 << 20
		gb = 1 << 30
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/gb)
	case n >= mb:
		return fmt.Sprintf("%.0f MB", float64(n)/mb)
	case n >= kb:
		return fmt.Sprintf("%.0f KB", float64(n)/kb)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func abbreviatePath(p string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

func shortDesc(d string) string {
	if len(d) <= 60 {
		return d
	}
	return d[:57] + "..."
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "reclaim: "+format+"\n", args...)
	os.Exit(1)
}

func printHelp() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf(`reclaim — a developer-aware Mac cleaner

USAGE:
    %s                            Interactive TUI (scan + pick + apply)
    %s --plain                    Print plain-text report only
    %s --apply                    Non-interactive apply (safe + confirm items)
    %s --apply -c build-caches    Limit to one category
    %s --version
    %s --help

FLAGS:
    -a, --apply               Apply without TUI (single y/N prompt)
    -p, --plain               Plain text scan, no TUI
    -c, --category <id>       Restrict to a category
    -v, --version             Print version
    -h, --help                Show this help

KEYS (in TUI):
    ↑/↓/j/k    navigate
    space      toggle selection (or expand category)
    →/l, ←/h   expand / collapse category
    a / n      select all safe / select none
    enter      apply selected (with confirmation)
    q / esc    quit

SAFETY:
    ✓ safe      Regenerates automatically — deletion has no impact
    ⚠ confirm   Should be reviewed before deleting
    ✗ dangerous Skipped by default — use targeted commands manually

LOG:
    Each apply run is logged to ~/.reclaim/logs/YYYY-MM-DD.jsonl
`, exe, exe, exe, exe, exe, exe)
}
