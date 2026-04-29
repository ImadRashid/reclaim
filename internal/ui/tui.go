// Package ui implements the interactive Bubble Tea TUI for reclaim.
//
// Layout:
//   - Top: title + total selected size
//   - Middle: scrollable list of categories and items (categories can be expanded/collapsed)
//   - Bottom: help bar with keybindings
//
// Keys:
//   ↑/↓/j/k  navigate
//   space    toggle selection (or expand category)
//   →/l      expand category
//   ←/h      collapse category
//   a        select all safe items
//   n        select none
//   enter    apply selected (with confirmation)
//   q/esc    quit
package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	categoryStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("87"))
	selectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("231"))
	safeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	confirmStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dangerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	totalStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("215"))
)

// Row is one visible line in the list — either a category header or a hit.
type Row struct {
	IsCategory bool
	CategoryID string
	HitIndex   int  // index into Model.hits when IsCategory=false
	Indent     int
}

type Model struct {
	catalog   *rules.Catalog
	ruleByID  map[string]rules.Rule
	hits      []scanner.Hit
	selected  map[int]bool // hit index -> selected
	expanded  map[string]bool
	rows      []Row
	cursor    int
	width     int
	height    int
	confirming bool
	confirmed  bool
	cancelled  bool
}

// New constructs a TUI model. By default, all "safe" hits are pre-selected
// and all categories are expanded.
func New(catalog *rules.Catalog, hits []scanner.Hit) Model {
	ruleByID := map[string]rules.Rule{}
	for _, r := range catalog.Rules {
		ruleByID[r.ID] = r
	}
	m := Model{
		catalog:  catalog,
		ruleByID: ruleByID,
		hits:     hits,
		selected: map[int]bool{},
		expanded: map[string]bool{},
	}
	for _, c := range catalog.Categories {
		m.expanded[c.ID] = true
	}
	for i, h := range hits {
		r := ruleByID[h.RuleID]
		if r.Safety == "safe" {
			m.selected[i] = true
		}
	}
	m.rebuildRows()
	return m
}

// Confirmed reports whether the user pressed enter to apply.
func (m Model) Confirmed() bool { return m.confirmed }

// Selected returns the hits the user chose.
func (m Model) Selected() []scanner.Hit {
	var out []scanner.Hit
	for i, h := range m.hits {
		if m.selected[i] {
			out = append(out, h)
		}
	}
	return out
}

func (m *Model) rebuildRows() {
	m.rows = nil
	// Group hits by category.
	byCategory := map[string][]int{}
	for i, h := range m.hits {
		r := m.ruleByID[h.RuleID]
		byCategory[r.Category] = append(byCategory[r.Category], i)
	}
	for _, cat := range m.catalog.Categories {
		idxs, ok := byCategory[cat.ID]
		if !ok || len(idxs) == 0 {
			continue
		}
		// Sort items within category by size desc.
		sort.Slice(idxs, func(a, b int) bool {
			return m.hits[idxs[a]].Size > m.hits[idxs[b]].Size
		})
		m.rows = append(m.rows, Row{IsCategory: true, CategoryID: cat.ID})
		if m.expanded[cat.ID] {
			for _, i := range idxs {
				m.rows = append(m.rows, Row{HitIndex: i, Indent: 1})
			}
		}
	}
	if m.cursor >= len(m.rows) {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y", "Y", "enter":
				m.confirmed = true
				return m, tea.Quit
			case "n", "N", "esc", "q":
				m.confirming = false
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "g", "home":
			m.cursor = 0
		case "G", "end":
			m.cursor = len(m.rows) - 1
		case "right", "l":
			row := m.rows[m.cursor]
			if row.IsCategory {
				m.expanded[row.CategoryID] = true
				m.rebuildRows()
			}
		case "left", "h":
			row := m.rows[m.cursor]
			if row.IsCategory {
				m.expanded[row.CategoryID] = false
				m.rebuildRows()
			}
		case " ":
			row := m.rows[m.cursor]
			if row.IsCategory {
				// Toggle: if any item in category is selected, deselect all; else select all "safe"
				anySelected := false
				for i, h := range m.hits {
					if m.ruleByID[h.RuleID].Category == row.CategoryID && m.selected[i] {
						anySelected = true
						break
					}
				}
				for i, h := range m.hits {
					if m.ruleByID[h.RuleID].Category != row.CategoryID {
						continue
					}
					if anySelected {
						delete(m.selected, i)
					} else if m.ruleByID[h.RuleID].Safety != "dangerous" {
						m.selected[i] = true
					}
				}
			} else {
				if m.selected[row.HitIndex] {
					delete(m.selected, row.HitIndex)
				} else {
					m.selected[row.HitIndex] = true
				}
			}
		case "a":
			for i, h := range m.hits {
				if m.ruleByID[h.RuleID].Safety != "dangerous" {
					m.selected[i] = true
				}
			}
		case "n":
			m.selected = map[int]bool{}
		case "enter":
			if len(m.Selected()) == 0 {
				return m, nil
			}
			m.confirming = true
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.cancelled {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("🧹 reclaim — interactive cleanup"))
	b.WriteString("\n")

	totalSelected := int64(0)
	totalAll := int64(0)
	for i, h := range m.hits {
		totalAll += h.Size
		if m.selected[i] {
			totalSelected += h.Size
		}
	}
	b.WriteString(fmt.Sprintf("Selected: %s of %s\n\n",
		totalStyle.Render(formatSize(totalSelected)),
		dimStyle.Render(formatSize(totalAll))))

	// Calculate visible window for scrolling.
	listHeight := m.height - 8 // account for title, totals, help
	if listHeight < 5 {
		listHeight = 5
	}
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}
	end := start + listHeight
	if end > len(m.rows) {
		end = len(m.rows)
	}

	for i := start; i < end; i++ {
		row := m.rows[i]
		line := m.renderRow(row, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad to bottom.
	for i := end - start; i < listHeight; i++ {
		b.WriteString("\n")
	}

	if m.confirming {
		b.WriteString("\n")
		b.WriteString(confirmStyle.Render(
			fmt.Sprintf("⚠ Delete %d items, freeing %s? [y/N]",
				len(m.Selected()), formatSize(totalSelected))))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render(
			"↑/↓ move  space toggle  ←/→ collapse  a all  n none  enter apply  q quit"))
	}
	return b.String()
}

func (m Model) renderRow(row Row, isCursor bool) string {
	var line string
	if row.IsCategory {
		cat := m.findCategory(row.CategoryID)
		expanded := m.expanded[row.CategoryID]
		marker := "▶ "
		if expanded {
			marker = "▼ "
		}
		// Compute category total + selected total
		var total, selected int64
		var count, selCount int
		for i, h := range m.hits {
			if m.ruleByID[h.RuleID].Category != row.CategoryID {
				continue
			}
			total += h.Size
			count++
			if m.selected[i] {
				selected += h.Size
				selCount++
			}
		}
		text := fmt.Sprintf("%s%s — %s [%d/%d items, %s selected]",
			marker, cat.Name, formatSize(total), selCount, count, formatSize(selected))
		line = categoryStyle.Render(text)
	} else {
		h := m.hits[row.HitIndex]
		r := m.ruleByID[h.RuleID]
		check := "[ ]"
		if m.selected[row.HitIndex] {
			check = "[✓]"
		}
		marker := safetyMarker(r.Safety)
		path := abbreviatePath(h.Path)
		if len(path) > 50 {
			path = "…" + path[len(path)-49:]
		}
		extra := ""
		if h.StalenessDays >= 0 {
			extra = " " + dimStyle.Render("("+scanner.StalenessLabel(h.StalenessDays)+")")
		}
		line = fmt.Sprintf("    %s %s %-50s  %10s%s",
			check, marker, path, formatSize(h.Size), extra)
	}
	if isCursor {
		line = selectedStyle.Render(strings.TrimRight(line, " "))
	}
	return line
}

func (m Model) findCategory(id string) rules.Category {
	for _, c := range m.catalog.Categories {
		if c.ID == id {
			return c
		}
	}
	return rules.Category{ID: id, Name: id}
}

func safetyMarker(s string) string {
	switch s {
	case "safe":
		return safeStyle.Render("✓")
	case "confirm":
		return confirmStyle.Render("⚠")
	case "dangerous":
		return dangerStyle.Render("✗")
	}
	return " "
}

func formatSize(n int64) string {
	const (
		kb = 1 << 10
		mb = 1 << 20
		gb = 1 << 30
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/gb)
	case n >= mb:
		return fmt.Sprintf("%.0f MB", float64(n)/mb)
	case n >= kb:
		return fmt.Sprintf("%.0f KB", float64(n)/kb)
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func abbreviatePath(p string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}
