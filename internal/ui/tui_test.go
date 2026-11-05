package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

func testModel() Model {
	cat := &rules.Catalog{
		Categories: []rules.Category{{ID: "build"}, {ID: "git-worktrees"}},
		Rules: []rules.Rule{
			{ID: "cache", Category: "build", Safety: "safe"},
			{ID: "git-worktrees", Category: "git-worktrees", Safety: "confirm", OptIn: true},
		},
	}
	hits := []scanner.Hit{
		{RuleID: "cache", Path: "/a", Size: 10},                                  // 0
		{RuleID: "cache", Path: "/blocked", Size: 10, Blocked: "in use"},         // 1
		{RuleID: "git-worktrees", Path: "/wt-clean", Size: 10},                   // 2
		{RuleID: "git-worktrees", Path: "/wt-dirty", Size: 10, Blocked: "dirty"}, // 3
	}
	return New(cat, hits)
}

func press(m Model, key tea.KeyMsg) Model {
	next, _ := m.Update(key)
	return next.(Model)
}

func runeKey(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func cursorTo(t *testing.T, m Model, path string) Model {
	t.Helper()
	for i, row := range m.rows {
		if !row.IsCategory && m.hits[row.HitIndex].Path == path {
			m.cursor = i
			return m
		}
	}
	t.Fatalf("no row for %s", path)
	return m
}

func TestNew_PreselectsOnlyUnblockedSafeNonOptIn(t *testing.T) {
	m := testModel()
	want := map[int]bool{0: true}
	for i := range m.hits {
		if m.selected[i] != want[i] {
			t.Errorf("hit %d (%s) selected=%v, want %v", i, m.hits[i].Path, m.selected[i], want[i])
		}
	}
}

func TestSelectAll_SkipsBlockedAndOptIn(t *testing.T) {
	m := press(testModel(), runeKey('a'))
	if m.selected[1] || m.selected[2] || m.selected[3] {
		t.Fatalf("'a' selected blocked or opt-in items: %v", m.selected)
	}
}

func TestSpace_CannotSelectBlockedItem(t *testing.T) {
	m := cursorTo(t, testModel(), "/wt-dirty")
	m = press(m, tea.KeyMsg{Type: tea.KeySpace})
	if m.selected[3] {
		t.Fatal("space selected a blocked worktree")
	}
}

func TestSpace_CanSelectOptInItemDeliberately(t *testing.T) {
	m := cursorTo(t, testModel(), "/wt-clean")
	m = press(m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.selected[2] {
		t.Fatal("space should select a clean opt-in worktree")
	}
}

func TestCategoryToggle_SkipsBlocked(t *testing.T) {
	m := testModel()
	for i, row := range m.rows {
		if row.IsCategory && row.CategoryID == "git-worktrees" {
			m.cursor = i
		}
	}
	m = press(m, tea.KeyMsg{Type: tea.KeySpace})
	if !m.selected[2] || m.selected[3] {
		t.Fatalf("category toggle: clean=%v dirty=%v, want true/false", m.selected[2], m.selected[3])
	}
}
