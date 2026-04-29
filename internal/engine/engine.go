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

		// Process check: skip if a named app is running.
		if r.ProcessCheck != "" && processRunning(r.ProcessCheck) {
			results = append(results, Result{
				RuleID:  h.RuleID,
				Path:    h.Path,
				Skipped: true,
				Reason:  fmt.Sprintf("%q is running — quit it first", r.ProcessCheck),
			})
			continue
		}

		// Command-based rules (e.g. xcrun simctl delete unavailable)
		if r.Detect == "command" && r.Cmd != "" {
			if err := runShell(r.Cmd); err != nil {
				results = append(results, Result{RuleID: h.RuleID, Err: err})
			} else {
				results = append(results, Result{RuleID: h.RuleID})
			}
			continue
		}

		// Path-based deletion.
		if h.Path == "" {
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
