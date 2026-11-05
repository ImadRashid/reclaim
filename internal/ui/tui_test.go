package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

// manyRowsModel builds a picker with one category and n items, sized so
// exactly 5 rows are visible.
func manyRowsModel(n int) Model {
	cat := &rules.Catalog{
		Categories: []rules.Category{{ID: "build"}},
		Rules:      []rules.Rule{{ID: "cache", Category: "build", Safety: "safe"}},
	}
	hits := make([]scanner.Hit, n)
	for i := range hits {
		hits[i] = scanner.Hit{RuleID: "cache", Path: "/p", Size: int64(n - i)}
	}
	m := New(cat, hits)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 14}) // listHeight = 5
	return next.(Model)
}

func TestScroll_CursorMovesFreelyInsideWindow(t *testing.T) {
	m := manyRowsModel(20) // 21 rows incl. category header
	for i := 0; i < 20; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.cursor != 20 || m.offset != 16 {
		t.Fatalf("at bottom: cursor=%d offset=%d, want 20/16", m.cursor, m.offset)
	}
	// Moving up within the window must not scroll.
	for i := 0; i < 4; i++ {
		m = press(m, tea.KeyMsg{Type: tea.KeyUp})
		if m.offset != 16 {
			t.Fatalf("after %d ups: offset=%d, want 16 (window should stay put)", i+1, m.offset)
		}
	}
	if m.cursor != 16 {
		t.Fatalf("cursor=%d, want 16 (top of window)", m.cursor)
	}
	// One more up crosses the top edge and scrolls by exactly one row.
	m = press(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 15 || m.offset != 15 {
		t.Fatalf("crossing top edge: cursor=%d offset=%d, want 15/15", m.cursor, m.offset)
	}
}

func TestScroll_CollapseNeverLeavesBlankWindow(t *testing.T) {
	m := manyRowsModel(20)
	m = press(m, runeKey('G'))
	// Jump to the category header and collapse: only 1 row remains.
	m = press(m, runeKey('g'))
	m = press(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.offset != 0 {
		t.Fatalf("offset=%d after collapse, want 0", m.offset)
	}
}

func TestRenderRow_LockReasonSurvivesNarrowTerminal(t *testing.T) {
	m := testModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 90, Height: 40})
	m = next.(Model)
	m.hits[3].Note = "feat/a-very-long-branch-name-that-would-push-everything-off-screen · merged · 19d idle"
	for _, row := range m.rows {
		if row.IsCategory || row.HitIndex != 3 {
			continue
		}
		line := m.renderRow(row, false)
		if w := lipgloss.Width(line); w > 90 {
			t.Errorf("row width %d exceeds terminal width 90", w)
		}
		if !strings.Contains(line, "🔒 dirty") {
			t.Errorf("lock reason clipped from row: %q", line)
		}
		return
	}
	t.Fatal("blocked row not found")
}
