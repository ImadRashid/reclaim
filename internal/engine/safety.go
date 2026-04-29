package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// forbiddenPaths is a hard-coded list of paths reclaim must never delete,
// no matter what a rule says. This protects against authoring mistakes in
// the YAML catalog and against malicious third-party rule files.
var forbiddenPaths = []string{
	"/",
	"/System",
	"/Library",
	"/Applications",
	"/usr",
	"/bin",
	"/sbin",
	"/etc",
	"/var",
	"/private",
	"/opt",
	"/tmp",
	"/Users",
}

// CheckPathSafe returns nil if it's safe for reclaim to delete `path`,
// or a non-nil error explaining why it isn't.
//
// Rules:
//   1. The path must be absolute and resolvable.
//   2. It must not be exactly any of the forbiddenPaths.
//   3. It must be inside the user's home directory (or /tmp for tests).
//   4. It must not equal the user's home directory itself.
func CheckPathSafe(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	clean := filepath.Clean(abs)

	for _, f := range forbiddenPaths {
		if clean == f {
			return fmt.Errorf("refusing to delete protected system path %q", clean)
		}
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		if clean == home {
			return fmt.Errorf("refusing to delete the home directory itself")
		}
		// Allow anything under home, plus standard temp locations (used in tests).
		allowed := strings.HasPrefix(clean, home+string(os.PathSeparator)) ||
			strings.HasPrefix(clean, "/tmp"+string(os.PathSeparator)) ||
			strings.HasPrefix(clean, "/var/folders"+string(os.PathSeparator)) ||
			strings.HasPrefix(clean, "/private/var/folders"+string(os.PathSeparator)) ||
			strings.HasPrefix(clean, "/private/tmp"+string(os.PathSeparator))
		if !allowed {
			return fmt.Errorf("refusing to delete path outside home: %q", clean)
		}
	}

	// Defensive: never delete two-component paths like /foo or ~/foo
	// (where foo is a major dotfile/library).
	if home != "" {
		rel, err := filepath.Rel(home, clean)
		if err == nil && rel != "" && !strings.Contains(rel, string(os.PathSeparator)) {
			// Direct child of home — allow only known cleanable directories.
			allowedTopLevel := map[string]bool{
				".gradle":        true,
				".npm":            true,
				".yarn":           true,
				".pnpm-store":     true,
				".pub-cache":      true,
				".cocoapods":      true,
				".cache":          true,
				".cargo":          true,
				".rustup":         true,
				".docker":         true,
				".m2":             true,
				".local":          true,
				".reclaim":        true,
				".Trash":          true,
			}
			if !allowedTopLevel[rel] {
				return fmt.Errorf("refusing to delete top-level home directory %q (not in allowlist)", rel)
			}
		}
	}
	return nil
}
