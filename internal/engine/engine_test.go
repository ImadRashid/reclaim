package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

func TestApply_DeletesAllowedPaths(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "to-delete")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "x"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	hits := []scanner.Hit{{RuleID: "test", Path: target, Size: 5}}
	ruleByID := map[string]rules.Rule{
		"test": {ID: "test", Safety: "safe"},
	}

	results := Apply(hits, ruleByID, nil)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if results[0].Skipped {
		t.Fatalf("unexpected skip: %s", results[0].Reason)
	}
	if results[0].Freed != 5 {
		t.Errorf("expected Freed=5, got %d", results[0].Freed)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("target should be deleted, got err=%v", err)
	}
}

func TestApply_RejectsSystemPath(t *testing.T) {
	hits := []scanner.Hit{{RuleID: "rogue", Path: "/usr/local"}}
	ruleByID := map[string]rules.Rule{
		"rogue": {ID: "rogue", Safety: "safe"},
	}
	results := Apply(hits, ruleByID, nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Errorf("expected Skipped=true for system path; got %+v", results[0])
	}
}

func TestRuleMap(t *testing.T) {
	cat := &rules.Catalog{
		Rules: []rules.Rule{
			{ID: "a"}, {ID: "b"},
		},
	}
	m := RuleMap(cat)
	if _, ok := m["a"]; !ok {
		t.Error("missing rule a")
	}
	if _, ok := m["b"]; !ok {
		t.Error("missing rule b")
	}
}
