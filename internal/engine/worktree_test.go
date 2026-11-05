package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

// wtFixture is a throwaway repo with one linked worktree on branch "feature".
type wtFixture struct {
	t    *testing.T
	main string
	wt   string
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir,
		"-c", "user.name=test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func newWorktreeFixture(t *testing.T) wtFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	// Isolate from the developer's git config (hooks, signing, etc.).
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Treat every worktree as idle unless a test says otherwise.
	prev := scanner.WorktreeMinIdle
	scanner.WorktreeMinIdle = 0
	t.Cleanup(func() { scanner.WorktreeMinIdle = prev })

	root := t.TempDir()
	main := filepath.Join(root, "repo")
	if err := os.MkdirAll(main, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, main, "init", "-q")
	write(t, filepath.Join(main, "README.md"), "hello\n")
	write(t, filepath.Join(main, ".gitignore"), "node_modules/\n.env*\n")
	git(t, main, "add", ".")
	git(t, main, "commit", "-q", "-m", "init")

	wt := filepath.Join(root, "wt-feature")
	git(t, main, "worktree", "add", "-q", "-b", "feature", wt)
	return wtFixture{t: t, main: main, wt: wt}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f wtFixture) apply(path string) Result {
	f.t.Helper()
	hits := []scanner.Hit{{RuleID: "git-worktrees", Path: path, ProjectRoot: f.main, Size: 1}}
	ruleByID := map[string]rules.Rule{
		"git-worktrees": {ID: "git-worktrees", Safety: "confirm", Action: "git-worktree-remove"},
	}
	results := Apply(hits, ruleByID, nil)
	if len(results) != 1 {
		f.t.Fatalf("expected 1 result, got %d", len(results))
	}
	return results[0]
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f wtFixture) assertRefused(res Result, wantReason string) {
	f.t.Helper()
	if !res.Skipped {
		f.t.Fatalf("expected skip, got %+v", res)
	}
	if !strings.Contains(res.Reason, wantReason) {
		f.t.Errorf("reason = %q, want it to mention %q", res.Reason, wantReason)
	}
	if !exists(f.wt) || !exists(f.main) {
		f.t.Fatalf("refused removal must leave the worktree and repo on disk")
	}
}

func TestApply_RemovesCleanWorktreeKeepsBranch(t *testing.T) {
	f := newWorktreeFixture(t)
	// Regenerable ignored content doesn't block removal.
	write(t, filepath.Join(f.wt, "node_modules", "pkg", "index.js"), "x")
	// Work committed on the branch is kept after removal.
	write(t, filepath.Join(f.wt, "feature.txt"), "done\n")
	git(t, f.wt, "add", ".")
	git(t, f.wt, "commit", "-q", "-m", "feature work")

	res := f.apply(f.wt)
	if res.Err != nil || res.Skipped {
		t.Fatalf("expected removal, got %+v", res)
	}
	if exists(f.wt) {
		t.Fatal("worktree directory should be gone")
	}
	if res.Freed <= 0 {
		t.Errorf("Freed = %d, want > 0", res.Freed)
	}
	if !exists(filepath.Join(f.main, "README.md")) {
		t.Fatal("main worktree must be untouched")
	}
	if got := git(t, f.main, "log", "-1", "--format=%s", "feature"); got != "feature work" {
		t.Errorf("branch tip = %q, want the committed work to survive", got)
	}
}

func TestApply_RefusesWorktreeWithModifiedFile(t *testing.T) {
	f := newWorktreeFixture(t)
	write(t, filepath.Join(f.wt, "README.md"), "edited\n")
	f.assertRefused(f.apply(f.wt), "uncommitted")
}

func TestApply_RefusesWorktreeWithUntrackedFile(t *testing.T) {
	f := newWorktreeFixture(t)
	write(t, filepath.Join(f.wt, "notes.txt"), "draft\n")
	f.assertRefused(f.apply(f.wt), "uncommitted")
}

func TestApply_RefusesWorktreeWithIgnoredSecret(t *testing.T) {
	f := newWorktreeFixture(t)
	write(t, filepath.Join(f.wt, "apps", "api", ".env.local"), "SECRET=1\n")
	f.assertRefused(f.apply(f.wt), ".env.local")
}

func TestApply_RefusesLockedWorktree(t *testing.T) {
	f := newWorktreeFixture(t)
	git(t, f.main, "worktree", "lock", f.wt)
	f.assertRefused(f.apply(f.wt), "locked")
}

func TestApply_RefusesDetachedCommitsOnNoBranch(t *testing.T) {
	f := newWorktreeFixture(t)
	git(t, f.wt, "checkout", "-q", "--detach")
	write(t, filepath.Join(f.wt, "orphan.txt"), "x\n")
	git(t, f.wt, "add", ".")
	git(t, f.wt, "commit", "-q", "-m", "detached work")
	f.assertRefused(f.apply(f.wt), "not on any branch")
}

func TestApply_RefusesRecentlyActiveWorktree(t *testing.T) {
	f := newWorktreeFixture(t)
	scanner.WorktreeMinIdle = 24 * time.Hour
	f.assertRefused(f.apply(f.wt), "active")
}

func TestApply_RefusesMainWorktree(t *testing.T) {
	f := newWorktreeFixture(t)
	res := f.apply(f.main)
	if !res.Skipped || !strings.Contains(res.Reason, "main worktree") {
		t.Fatalf("expected main worktree refusal, got %+v", res)
	}
	if !exists(filepath.Join(f.main, "README.md")) {
		t.Fatal("main worktree must be untouched")
	}
}

func TestApply_RefusesDirectoryThatIsNotAWorktree(t *testing.T) {
	f := newWorktreeFixture(t)
	other := filepath.Join(filepath.Dir(f.main), "just-a-folder")
	write(t, filepath.Join(other, "file.txt"), "keep me\n")
	res := f.apply(other)
	if !res.Skipped || !strings.Contains(res.Reason, "not a linked worktree") {
		t.Fatalf("expected refusal, got %+v", res)
	}
	if !exists(filepath.Join(other, "file.txt")) {
		t.Fatal("non-worktree directory must be untouched")
	}
}

func TestApply_RefusesWorktreeOutsideSafePaths(t *testing.T) {
	f := newWorktreeFixture(t)
	res := f.apply("/usr/local")
	if !res.Skipped || !strings.Contains(res.Reason, "safety check") {
		t.Fatalf("expected safety refusal, got %+v", res)
	}
}

func TestApply_NeverActsOnBlockedHit(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "keep")
	write(t, filepath.Join(target, "x"), "data")
	hits := []scanner.Hit{{RuleID: "test", Path: target, Size: 4, Blocked: "3 uncommitted changes"}}
	ruleByID := map[string]rules.Rule{"test": {ID: "test", Safety: "safe"}}

	results := Apply(hits, ruleByID, nil)
	if len(results) != 1 || !results[0].Skipped || results[0].Reason != "3 uncommitted changes" {
		t.Fatalf("expected blocked skip, got %+v", results)
	}
	if !exists(target) {
		t.Fatal("blocked item must not be deleted")
	}
}

func TestApply_AlreadyRemovedPathIsNotCountedAsFreed(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	hits := []scanner.Hit{{RuleID: "test", Path: missing, Size: 100}}
	ruleByID := map[string]rules.Rule{"test": {ID: "test", Safety: "safe"}}
	results := Apply(hits, ruleByID, nil)
	if len(results) != 1 || !results[0].Skipped || results[0].Freed != 0 {
		t.Fatalf("expected skip with nothing freed, got %+v", results)
	}
}
