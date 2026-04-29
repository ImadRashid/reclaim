package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ImadRashid/reclaim/internal/rules"
)

// Hit represents a single discovered cleanable target.
type Hit struct {
	RuleID         string
	Path           string
	Size           int64
	ProjectRoot    string // git repo root for project-* rules; "" otherwise
	StalenessDays  int    // days since last git activity for the project; -1 if N/A
}

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return p
}

// DirSize sums file sizes under a directory tree.
// It silently skips files we can't stat (e.g. macOS-protected).
func DirSize(root string) int64 {
	var total int64
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
		total += info.Size()
		return nil
	})
	return total
}

// Scan walks the catalog and returns all hits, with sizes.
// progress, if non-nil, is called once per rule with the rule id.
func Scan(ctx context.Context, cat *rules.Catalog, progress func(ruleID string)) []Hit {
	var hits []Hit
	for _, r := range cat.Rules {
		if progress != nil {
			progress(r.ID)
		}
		select {
		case <-ctx.Done():
			return hits
		default:
		}
		hits = append(hits, scanRule(r)...)
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
		return nil // commands report 0 size; engine reruns them
	}
	return nil
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
