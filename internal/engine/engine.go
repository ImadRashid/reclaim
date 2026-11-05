package engine

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

// Result captures the outcome of cleaning a single hit.
type Result struct {
	RuleID  string
	Path    string
	Freed   int64
	Err     error
	Skipped bool
	Reason  string
}

// Apply deletes the given hits and returns per-hit results.
// onProgress, if non-nil, is called before each deletion.
func Apply(hits []scanner.Hit, ruleByID map[string]rules.Rule, onProgress func(scanner.Hit)) []Result {
	var results []Result
	for _, h := range hits {
		if onProgress != nil {
			onProgress(h)
		}
		r := ruleByID[h.RuleID]

		// The scanner marks items that would lose work (e.g. a worktree with
		// uncommitted changes). Never act on them, whatever was selected.
		if h.Blocked != "" {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  h.Blocked,
			})
			continue
		}

		// Single-process check (legacy ProcessCheck): skip if the app is running.
		if r.ProcessCheck != "" && processRunning(r.ProcessCheck) {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  fmt.Sprintf("%q is running — quit it first", r.ProcessCheck),
			})
			continue
		}

		// Multi-process check (WarnIfRunning): skip if any listed app is running.
		if running := firstRunning(r.WarnIfRunning); running != "" {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  fmt.Sprintf("%q is running — quit it first", running),
			})
			continue
		}

		// Shell-action rules (e.g. `docker system prune`). Run the action
		// command, capture output, and report a single result.
		if r.Action == "shell" && r.ActionCommand != "" {
			if err := runShell(r.ActionCommand); err != nil {
				results = append(results, Result{RuleID: h.RuleID, Err: err})
			} else {
				results = append(results, Result{RuleID: h.RuleID})
			}
			continue
		}

		// Legacy command-based rules (e.g. xcrun simctl delete unavailable)
		// Action: "shell" supersedes this; keep it for backwards compatibility
		// with existing rules that use the older Cmd field.
		if r.Detect == "command" && r.Cmd != "" {
			if err := runShell(r.Cmd); err != nil {
				results = append(results, Result{RuleID: h.RuleID, Err: err})
			} else {
				results = append(results, Result{RuleID: h.RuleID})
			}
			continue
		}

		if r.Action == "git-worktree-remove" {
			results = append(results, removeWorktree(h))
			continue
		}

		// Path-based deletion.
		if h.Path == "" {
			continue
		}
		// Already gone (e.g. it lived inside a worktree removed earlier in
		// this run): report a skip rather than claiming the space was freed.
		if _, err := os.Lstat(h.Path); os.IsNotExist(err) {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  "already removed",
			})
			continue
		}
		if err := CheckPathSafe(h.Path); err != nil {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  fmt.Sprintf("safety check: %v", err),
			})
			continue
		}
		size := h.Size
		if err := os.RemoveAll(h.Path); err != nil {
			results = append(results, Result{RuleID: h.RuleID, Path: h.Path, Err: err})
			continue
		}
		results = append(results, Result{RuleID: h.RuleID, Path: h.Path, Freed: size})
	}
	return results
}

// RuleMap turns a catalog's rule slice into a lookup map.
func RuleMap(c *rules.Catalog) map[string]rules.Rule {
	m := make(map[string]rules.Rule, len(c.Rules))
	for _, r := range c.Rules {
		m[r.ID] = r
	}
	return m
}

func runShell(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func processRunning(name string) bool {
	out, err := exec.Command("pgrep", "-x", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// firstRunning returns the name of the first process from the list that is
// currently running, or "" if none are. Used by WarnIfRunning to decide
// whether to skip a rule's deletion.
func firstRunning(names []string) string {
	for _, n := range names {
		if processRunning(n) {
			return n
		}
	}
	return ""
}
