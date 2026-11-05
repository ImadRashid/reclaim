package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ImadRashid/reclaim/internal/gitwt"
	"github.com/ImadRashid/reclaim/internal/rules"
)

// WorktreeMinIdle is how long a worktree must be untouched before it can be
// selected. Recent activity suggests a person or agent is working in it.
var WorktreeMinIdle = 24 * time.Hour

// scanWorktrees finds git repositories under ScanRoot and emits one hit per
// linked worktree, wherever on disk (under $HOME) that worktree lives.
// Worktrees that would lose work are still emitted — with Blocked set — so
// the user sees why that space isn't reclaimable.
func scanWorktrees(r rules.Rule) []Hit {
	root := ExpandHome(r.ScanRoot)
	if _, err := os.Stat(root); err != nil {
		return nil
	}
	home, _ := os.UserHomeDir()
	home = gitwt.Resolve(home)
	now := time.Now()

	var hits []Hit
	seen := map[string]bool{}
	for _, repo := range findRepos(root, r.ScanMaxDepth) {
		list, err := gitwt.List(repo)
		if err != nil {
			continue
		}
		for _, w := range list {
			if w.Prunable || seen[w.Path] {
				continue
			}
			seen[w.Path] = true
			if !strings.HasPrefix(w.Path, home+string(os.PathSeparator)) {
				continue // never offer anything outside $HOME
			}
			if fi, err := os.Stat(w.Path); err != nil || !fi.IsDir() {
				continue
			}
			hit := Hit{
				RuleID:        r.ID,
				Path:          w.Path,
				Size:          DirSize(w.Path),
				ProjectRoot:   w.MainPath,
				StalenessDays: -1,
			}
			st, err := gitwt.Inspect(w)
			if err != nil {
				hit.Blocked = "couldn't inspect: " + err.Error()
			} else {
				hit.Note = gitwt.Label(w, st, now)
				hit.Blocked = gitwt.Blocker(w, st, WorktreeMinIdle, now)
			}
			hits = append(hits, hit)
		}
	}
	return hits
}

// findRepos returns main-worktree roots (directories whose .git is a
// directory) under root. It doesn't descend into repos, node_modules, or
// hidden directories, so it stays fast on large project trees.
func findRepos(root string, maxDepth int) []string {
	if maxDepth == 0 {
		maxDepth = 4
	}
	rootDepth := strings.Count(root, string(os.PathSeparator))
	var repos []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if strings.Count(p, string(os.PathSeparator))-rootDepth > maxDepth {
			return filepath.SkipDir
		}
		base := d.Name()
		if p != root && (base == "node_modules" || strings.HasPrefix(base, ".")) {
			return filepath.SkipDir
		}
		if fi, err := os.Stat(filepath.Join(p, ".git")); err == nil && fi.IsDir() {
			repos = append(repos, p)
			return filepath.SkipDir
		}
		return nil
	})
	return repos
}
