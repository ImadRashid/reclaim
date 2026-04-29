package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := ExpandHome("~/foo"); got != filepath.Join(home, "foo") {
		t.Fatalf("ExpandHome(~/foo) = %q, want %q", got, filepath.Join(home, "foo"))
	}
	if got := ExpandHome("/abs/path"); got != "/abs/path" {
		t.Fatalf("ExpandHome left absolute path alone, got %q", got)
	}
}

func TestDirSize(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(strings.Repeat("x", 100)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got := DirSize(dir); got != 300 {
		t.Fatalf("DirSize = %d, want 300", got)
	}
}

func TestStalenessLabel(t *testing.T) {
	cases := map[int]string{
		-1:   "no git",
		0:    "active",
		20:   "recent",
		90:   "idle",
		200:  "stale",
		500:  "abandoned",
	}
	for days, want := range cases {
		if got := StalenessLabel(days); got != want {
			t.Errorf("StalenessLabel(%d) = %q, want %q", days, got, want)
		}
	}
}
