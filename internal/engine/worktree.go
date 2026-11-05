package engine

import (
	"fmt"
	"time"

	"github.com/ImadRashid/reclaim/internal/gitwt"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

// removeWorktree removes a linked git worktree via `git worktree remove`.
//
// Safety layers, in order:
//  1. CheckPathSafe — same guard as every other deletion.
//  2. The path must be a linked worktree of h.ProjectRoot right now (never
//     the main worktree, never an arbitrary directory).
//  3. A fresh gitwt.Inspect must find nothing to lose — state may have
//     changed since the scan.
//  4. git itself refuses to remove a dirty or locked worktree; we never
//     pass --force, so that refusal stands.
//
// The branch and its commits are kept; only the working directory and its
// git admin entry go away.
func removeWorktree(h scanner.Hit) Result {
	res := Result{RuleID: h.RuleID, Path: h.Path}
	skip := func(reason string) Result {
		res.Skipped = true
		res.Reason = reason
		return res
	}

	if err := CheckPathSafe(h.Path); err != nil {
		return skip(fmt.Sprintf("safety check: %v", err))
	}
	if h.ProjectRoot == "" {
		return skip("safety check: worktree has no owning repository")
	}
	w, err := gitwt.Find(h.ProjectRoot, h.Path)
	if err != nil {
		return skip(fmt.Sprintf("safety check: %v", err))
	}
	st, err := gitwt.Inspect(w)
	if err != nil {
		return skip(fmt.Sprintf("couldn't inspect worktree: %v", err))
	}
	if reason := gitwt.Blocker(w, st, scanner.WorktreeMinIdle, time.Now()); reason != "" {
		return skip(reason)
	}

	// Measure now, not at scan time: nested items (node_modules) may have
	// been cleaned earlier in this run.
	size := scanner.DirSize(w.Path)
	if err := gitwt.Remove(w); err != nil {
		res.Err = err
		return res
	}
	res.Freed = size
	return res
}
