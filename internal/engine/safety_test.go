package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckPathSafe_RejectsSystemPaths(t *testing.T) {
	for _, p := range []string{
		"/", "/System", "/usr", "/etc", "/Users", "/Library",
	} {
		if err := CheckPathSafe(p); err == nil {
			t.Errorf("CheckPathSafe(%q) should have rejected it", p)
		}
	}
}

func TestCheckPathSafe_RejectsHomeItself(t *testing.T) {
	home, _ := os.UserHomeDir()
	if err := CheckPathSafe(home); err == nil {
		t.Errorf("CheckPathSafe(home) should have rejected it")
	}
}

func TestCheckPathSafe_RejectsUnknownTopLevelHomeDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	bad := filepath.Join(home, "Documents")
	if err := CheckPathSafe(bad); err == nil {
		t.Errorf("CheckPathSafe(%q) should have rejected — not in allowlist", bad)
	}
}

func TestCheckPathSafe_AllowsKnownCacheDirs(t *testing.T) {
	home, _ := os.UserHomeDir()
	for _, sub := range []string{".gradle", ".npm", ".cache", ".pub-cache"} {
		p := filepath.Join(home, sub, "anything")
		if err := CheckPathSafe(p); err != nil {
			t.Errorf("CheckPathSafe(%q) unexpectedly rejected: %v", p, err)
		}
	}
}

func TestCheckPathSafe_AllowsTempDirs(t *testing.T) {
	dir := t.TempDir()
	if err := CheckPathSafe(dir); err != nil {
		t.Errorf("CheckPathSafe(%q) for temp dir should be allowed: %v", dir, err)
	}
}

func TestCheckPathSafe_RejectsEmpty(t *testing.T) {
	if err := CheckPathSafe(""); err == nil {
		t.Error("CheckPathSafe(\"\") should fail")
	}
}
