// Package gitwt inspects linked git worktrees and decides whether one can be
// removed without losing work.
//
// It is shared by the scanner (to classify worktrees for display) and the
// engine (to re-check immediately before removal, since state can change
// between scan and apply). Everything here is read-only: git is invoked with
// GIT_OPTIONAL_LOCKS=0 so `git status` never rewrites the index of a
// worktree another tool may be using.
package gitwt

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Worktree is one entry from `git worktree list --porcelain`.
type Worktree struct {
	Path     string // absolute, symlinks resolved
	MainPath string // the repo's main worktree
	Branch   string // short branch name; "" when detached
	Locked   bool
	Prunable bool // git reports the directory as missing
}

// Status describes what removing a worktree would lose.
type Status struct {
	Dirty         int      // modified tracked files + untracked files
	Precious      []string // ignored files that look irreplaceable (.env, keys)
	Orphaned      int      // commits reachable from HEAD but from no branch, tag or remote
	Unpushed      int      // commits not on any remote (still kept on the branch)
	Merged        bool     // HEAD is contained in the default branch
	DefaultBranch string   // e.g. "origin/main"; "" if unknown
	LastActivity  time.Time
}

// Run executes git with optional locks disabled and returns stdout.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// Resolve returns an absolute, symlink-free form of p so paths reported by
// git (which resolves symlinks) compare equal to paths we construct.
func Resolve(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// List returns the linked (non-main) worktrees of the repository at repo.
func List(repo string) ([]Worktree, error) {
	out, err := run(repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var all []Worktree
	var cur *Worktree
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		key, val, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			all = append(all, Worktree{Path: Resolve(val)})
			cur = &all[len(all)-1]
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(val, "refs/heads/")
			}
		case "locked":
			if cur != nil {
				cur.Locked = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	if len(all) == 0 {
		return nil, nil
	}
	// The first entry is always the main worktree.
	main := all[0].Path
	linked := make([]Worktree, 0, len(all)-1)
	for _, w := range all[1:] {
		w.MainPath = main
		linked = append(linked, w)
	}
	return linked, nil
}

// Find returns the linked worktree at path belonging to the repo whose main
// worktree is mainPath, or an error if path is not such a worktree (including
// when path is the main worktree itself).
func Find(mainPath, path string) (Worktree, error) {
	want := Resolve(path)
	if want == Resolve(mainPath) {
		return Worktree{}, fmt.Errorf("%s is the main worktree, not a linked one", path)
	}
	list, err := List(mainPath)
	if err != nil {
		return Worktree{}, err
	}
	for _, w := range list {
		if w.Path == want {
			return w, nil
		}
	}
	return Worktree{}, fmt.Errorf("%s is not a linked worktree of %s", path, mainPath)
}

// Inspect gathers everything needed to decide whether w is safe to remove.
func Inspect(w Worktree) (Status, error) {
	var s Status

	// Activity first, before any git command touches metadata.
	gitDir, err := run(w.Path, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return s, err
	}
	gitDir = strings.TrimSpace(gitDir)
	for _, f := range []string{"HEAD", "index"} {
		if fi, err := os.Stat(filepath.Join(gitDir, f)); err == nil && fi.ModTime().After(s.LastActivity) {
			s.LastActivity = fi.ModTime()
		}
	}
	if out, err := run(w.Path, "log", "-1", "--format=%ct"); err == nil {
		if ts, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64); err == nil {
			if t := time.Unix(ts, 0); t.After(s.LastActivity) {
				s.LastActivity = t
			}
		}
	}

	// Uncommitted work: modified tracked files and untracked files.
	out, err := run(w.Path, "status", "--porcelain=v1", "--untracked-files=normal", "--ignore-submodules=none")
	if err != nil {
		return s, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			s.Dirty++
		}
	}

	// Ignored files git would silently delete along with the worktree.
	out, err = run(w.Path, "status", "--porcelain=v1", "--ignored=matching", "--untracked-files=normal")
	if err != nil {
		return s, err
	}
	for _, line := range strings.Split(out, "\n") {
		if p, ok := strings.CutPrefix(line, "!! "); ok && IsPrecious(p) {
			s.Precious = append(s.Precious, strings.TrimSuffix(p, "/"))
		}
	}

	// Commits that exist only via this worktree's HEAD (detached work).
	if n, err := count(w.Path, "rev-list", "--count", "HEAD", "--not", "--branches", "--tags", "--remotes"); err == nil {
		s.Orphaned = n
	} else {
		return s, err
	}
	if n, err := count(w.Path, "rev-list", "--count", "HEAD", "--not", "--remotes"); err == nil {
		s.Unpushed = n
	}

	s.DefaultBranch = defaultBranch(w.Path)
	if s.DefaultBranch != "" {
		_, err := run(w.Path, "merge-base", "--is-ancestor", "HEAD", s.DefaultBranch)
		s.Merged = err == nil
	}
	return s, nil
}

func count(dir string, args ...string) (int, error) {
	out, err := run(dir, args...)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func defaultBranch(dir string) string {
	if out, err := run(dir, "rev-parse", "--abbrev-ref", "origin/HEAD"); err == nil {
		if b := strings.TrimSpace(out); b != "" && b != "origin/HEAD" {
			return b
		}
	}
	for _, b := range []string{"origin/main", "origin/master", "main", "master"} {
		if _, err := run(dir, "rev-parse", "--verify", "--quiet", b+"^{commit}"); err == nil {
			return b
		}
	}
	return ""
}

// preciousNames are gitignored files that are commonly the only copy of
// something: local secrets and signing material. Matched on basename.
var preciousNames = map[string]bool{
	".npmrc":                   true,
	".pypirc":                  true,
	"key.properties":           true, // Flutter/Android release signing
	"google-services.json":     true,
	"GoogleService-Info.plist": true,
}

var preciousExts = []string{".pem", ".key", ".p8", ".p12", ".pfx", ".jks", ".keystore", ".mobileprovision"}

// IsPrecious reports whether an ignored path looks irreplaceable.
func IsPrecious(p string) bool {
	base := filepath.Base(strings.TrimSuffix(p, "/"))
	if preciousNames[base] {
		return true
	}
	for _, prefix := range []string{".env", ".dev.vars"} {
		if base == prefix || strings.HasPrefix(base, prefix+".") {
			switch strings.TrimPrefix(base, prefix+".") {
			case "example", "sample", "template", "dist":
				return false
			}
			return true
		}
	}
	lower := strings.ToLower(base)
	for _, ext := range preciousExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// Blocker returns why the worktree must not be removed, or "" if removing it
// loses nothing beyond regenerable ignored files. minIdle is how long the
// worktree must have been untouched; activity inside that window is treated
// as "someone may be working here right now".
func Blocker(w Worktree, s Status, minIdle time.Duration, now time.Time) string {
	switch {
	case w.Locked:
		return "locked (git worktree unlock to allow)"
	case s.Dirty > 0:
		return plural(s.Dirty, "uncommitted change")
	case len(s.Precious) > 0:
		more := ""
		if len(s.Precious) > 1 {
			more = fmt.Sprintf(" +%d more", len(s.Precious)-1)
		}
		return fmt.Sprintf("ignored file worth keeping: %s%s", s.Precious[0], more)
	case s.Orphaned > 0:
		return plural(s.Orphaned, "commit") + " not on any branch"
	case minIdle > 0 && now.Sub(s.LastActivity) < minIdle:
		return fmt.Sprintf("active in the last %dh", int(minIdle.Hours()))
	}
	return ""
}

// Label is a short human description, e.g. "fix/foo · merged · 13d idle".
func Label(w Worktree, s Status, now time.Time) string {
	parts := []string{}
	if w.Branch != "" {
		parts = append(parts, w.Branch)
	} else {
		parts = append(parts, "detached")
	}
	switch {
	case s.Merged:
		parts = append(parts, "merged")
	case s.Unpushed > 0:
		parts = append(parts, plural(s.Unpushed, "unpushed commit"))
	case s.DefaultBranch != "":
		parts = append(parts, "pushed, not merged")
	}
	if !s.LastActivity.IsZero() {
		parts = append(parts, fmt.Sprintf("%dd idle", int(now.Sub(s.LastActivity).Hours()/24)))
	}
	return strings.Join(parts, " · ")
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// Remove runs `git worktree remove` (without --force) from the main
// worktree. git refuses if the worktree is dirty, has untracked files, or
// is locked. The branch is left in place.
func Remove(w Worktree) error {
	_, err := run(w.MainPath, "worktree", "remove", w.Path)
	return err
}
