package scanner

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ImadRashid/reclaim/internal/rules"
)

// Hit represents a single discovered cleanable target.
type Hit struct {
	RuleID        string
	Path          string
	Size          int64
	ProjectRoot   string // git repo root for project-* rules; main worktree for git-worktrees; "" otherwise
	StalenessDays int    // days since last git activity for the project; -1 if N/A
	Note          string // short status shown next to the item (e.g. "fix/foo · merged · 13d idle")
	Blocked       string // non-empty: why this item can't be cleaned; UI and engine refuse it
}

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return p
}

// DirSize reports how much disk space a directory tree occupies — i.e. what
// deleting it would actually free. It counts allocated blocks rather than
// logical file length, so sparse files (Docker.raw, VM disk images) report
// their real footprint instead of their nominal size, and it counts each
// hard-linked file once. Files we can't stat (e.g. macOS-protected) are
// silently skipped.
func DirSize(root string) int64 {
	var total int64
	seen := map[fileID]bool{}
	_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		size, id, linked := diskUsage(info)
		if linked {
			if seen[id] {
				return nil
			}
			seen[id] = true
		}
		total += size
		return nil
	})
	return total
}

// Scan walks the catalog and returns all hits, with sizes.
// progress, if non-nil, is called once per rule with the rule id.
func Scan(ctx context.Context, cat *rules.Catalog, progress func(ruleID string)) []Hit {
	var onRule func(ruleID string, index, total int)
	if progress != nil {
		onRule = func(ruleID string, _, _ int) { progress(ruleID) }
	}
	return ScanEvents(ctx, cat, onRule, nil)
}

// ScanEvents is Scan with richer progress reporting, for interactive UIs.
// onRule, if non-nil, fires before each rule with its position in the
// catalog; onHits, if non-nil, fires after each rule that produced hits.
// Both are called from the goroutine running the scan.
func ScanEvents(ctx context.Context, cat *rules.Catalog, onRule func(ruleID string, index, total int), onHits func([]Hit)) []Hit {
	var hits []Hit
	total := len(cat.Rules)
	for i, r := range cat.Rules {
		if onRule != nil {
			onRule(r.ID, i+1, total)
		}
		select {
		case <-ctx.Done():
			return hits
		default:
		}
		found := scanRule(r)
		if len(found) > 0 {
			hits = append(hits, found...)
			if onHits != nil {
				onHits(found)
			}
		}
	}
	return hits
}

func scanRule(r rules.Rule) []Hit {
	switch r.Detect {
	case "dir-exists":
		return scanFixedPaths(r)
	case "scan":
		return scanPattern(r)
	case "command":
		return scanCommandRule(r)
	case "git-worktrees":
		return scanWorktrees(r)
	}
	return nil
}

// scanCommandRule emits a single virtual hit for a rule whose presence is
// gated on a probe command. The hit has no path and zero size — it's a
// signal to the UI/engine that this rule is available. The probe command,
// if specified via detect_command, must exit 0.
func scanCommandRule(r rules.Rule) []Hit {
	if r.DetectCommand != "" {
		if err := exec.Command("sh", "-c", r.DetectCommand).Run(); err != nil {
			return nil
		}
	}
	// Synthetic hit: surfaces the rule in the UI without a real path.
	return []Hit{{RuleID: r.ID}}
}

func scanFixedPaths(r rules.Rule) []Hit {
	var out []Hit
	for _, p := range r.Paths {
		expanded := ExpandHome(p)
		fi, err := os.Stat(expanded)
		if err != nil || !fi.IsDir() {
			continue
		}
		out = append(out, Hit{
			RuleID: r.ID,
			Path:   expanded,
			Size:   DirSize(expanded),
		})
	}
	return out
}

// scanPattern walks ScanRoot looking for directories matching ScanFor.
// ScanFor may be a single name or "a|b" alternation.
func scanPattern(r rules.Rule) []Hit {
	root := ExpandHome(r.ScanRoot)
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	names := strings.Split(r.ScanFor, "|")
	maxDepth := r.ScanMaxDepth
	if maxDepth == 0 {
		maxDepth = 5
	}

	var hits []Hit
	rootDepth := strings.Count(root, string(os.PathSeparator))

	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		depth := strings.Count(p, string(os.PathSeparator)) - rootDepth
		if depth > maxDepth {
			return filepath.SkipDir
		}
		base := filepath.Base(p)
		matched := false
		for _, n := range names {
			if base == n {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
		// Optional: path must contain a substring (e.g. "ios" for Pods).
		if r.PathContains != "" && !strings.Contains(p, string(os.PathSeparator)+r.PathContains+string(os.PathSeparator)) {
			return nil
		}
		// Optional: must have a sibling file (e.g. Cargo.toml for Rust target).
		if r.RequiresSibling != "" {
			if _, err := os.Stat(filepath.Join(filepath.Dir(p), r.RequiresSibling)); err != nil {
				return nil
			}
		}
		hit := Hit{
			RuleID: r.ID,
			Path:   p,
			Size:   DirSize(p),
		}
		if r.ActivityCheck {
			root := FindProjectRoot(filepath.Dir(p))
			hit.ProjectRoot = root
			hit.StalenessDays = LastGitActivityDays(root)
		} else {
			hit.StalenessDays = -1
		}
		hits = append(hits, hit)
		// Don't descend into matched directories.
		return filepath.SkipDir
	})
	return hits
}
