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
	// Disk usage is block-granular, so three 100-byte files occupy at least
	// 300 bytes (typically 3 × 4 KB) — never less than their content.
	got := DirSize(dir)
	if got < 300 {
		t.Fatalf("DirSize = %d, want >= 300", got)
	}
	if got > 3*64*1024 {
		t.Fatalf("DirSize = %d, implausibly large for three 100-byte files", got)
	}
}

// A sparse file's logical length can be huge while it occupies almost no
// disk. Docker.raw is the real-world case: 460 GB nominal, ~30 GB on disk.
// DirSize must report what deleting it would free, not the nominal length.
func TestDirSize_SparseFileCountsAllocatedBlocks(t *testing.T) {
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, "disk.raw"))
	if err != nil {
		t.Fatal(err)
	}
	const nominal = 1 << 30 // 1 GB
	if err := f.Truncate(nominal); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if got := DirSize(dir); got >= nominal/2 {
		t.Fatalf("DirSize = %d for a sparse 1 GB file; want allocated size, not logical length", got)
	}
}

func TestDirSize_HardLinksCountedOnce(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	if err := os.WriteFile(a, []byte(strings.Repeat("x", 64*1024)), 0o644); err != nil {
		t.Fatal(err)
	}
	single := DirSize(dir)
	if err := os.Link(a, filepath.Join(dir, "b")); err != nil {
		t.Skipf("hard links unsupported: %v", err)
	}
	if got := DirSize(dir); got != single {
		t.Fatalf("DirSize with hard link = %d, want %d (same as single file)", got, single)
	}
}

func TestStalenessLabel(t *testing.T) {
	cases := map[int]string{
		-1:  "no git",
		0:   "active",
		20:  "recent",
		90:  "idle",
		200: "stale",
		500: "abandoned",
	}
	for days, want := range cases {
		if got := StalenessLabel(days); got != want {
			t.Errorf("StalenessLabel(%d) = %q, want %q", days, got, want)
		}
	}
}
