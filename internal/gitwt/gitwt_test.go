package gitwt

import (
	"testing"
	"time"
)

func TestIsPrecious(t *testing.T) {
	precious := []string{
		".env", ".env.local", "apps/api/.env.production", ".dev.vars",
		"android/key.properties", "android/app/upload.jks", "certs/server.pem",
		"AuthKey_ABC.p8", "ios/Runner/GoogleService-Info.plist", ".npmrc",
	}
	for _, p := range precious {
		if !IsPrecious(p) {
			t.Errorf("IsPrecious(%q) = false, want true", p)
		}
	}
	regenerable := []string{
		"node_modules/", "apps/web/.next/", "dist/", ".env.example", ".env.sample",
		"apps/api/src/generated/prisma/", "build/", ".turbo/", "coverage/", "debug.log",
	}
	for _, p := range regenerable {
		if IsPrecious(p) {
			t.Errorf("IsPrecious(%q) = true, want false", p)
		}
	}
}

func TestBlocker(t *testing.T) {
	now := time.Now()
	idle := Status{LastActivity: now.Add(-48 * time.Hour)}
	if got := Blocker(Worktree{}, idle, 24*time.Hour, now); got != "" {
		t.Errorf("clean idle worktree blocked: %q", got)
	}
	cases := []struct {
		name string
		w    Worktree
		s    Status
	}{
		{"locked", Worktree{Locked: true}, idle},
		{"dirty", Worktree{}, Status{Dirty: 2, LastActivity: idle.LastActivity}},
		{"precious", Worktree{}, Status{Precious: []string{".env"}, LastActivity: idle.LastActivity}},
		{"orphaned", Worktree{}, Status{Orphaned: 1, LastActivity: idle.LastActivity}},
		{"active", Worktree{}, Status{LastActivity: now.Add(-time.Hour)}},
	}
	for _, c := range cases {
		if got := Blocker(c.w, c.s, 24*time.Hour, now); got == "" {
			t.Errorf("%s: expected a blocker", c.name)
		}
	}
}
