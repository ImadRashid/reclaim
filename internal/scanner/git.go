package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FindProjectRoot walks up from a path to find the nearest directory containing .git.
// Returns "" if none found before hitting the user's home directory or filesystem root.
func FindProjectRoot(p string) string {
	home, _ := os.UserHomeDir()
	cur := p
	for {
		if cur == "/" || cur == home || cur == "" {
			return ""
		}
		if _, err := os.Stat(filepath.Join(cur, ".git")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

// LastGitActivityDays returns the number of days since the most recent commit
// in the given repo, or -1 if it can't be determined.
func LastGitActivityDays(repoRoot string) int {
	if repoRoot == "" {
		return -1
	}
	cmd := exec.Command("git", "-C", repoRoot, "log", "-1", "--format=%ct")
	out, err := cmd.Output()
	if err != nil {
		return -1
	}
	ts, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return -1
	}
	return int(time.Since(time.Unix(ts, 0)).Hours() / 24)
}

// StalenessLabel returns a short label like "active", "1mo", "stale (1y)" for display.
func StalenessLabel(days int) string {
	switch {
	case days < 0:
		return "no git"
	case days <= 14:
		return "active"
	case days <= 60:
		return "recent"
	case days <= 180:
		return "idle"
	case days <= 365:
		return "stale"
	default:
		return "abandoned"
	}
}
