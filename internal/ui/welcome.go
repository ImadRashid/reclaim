package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ImadRashid/reclaim/internal/rules"
)

// MenuChoice is what the user picked from the welcome screen.
type MenuChoice int

const (
	ChoiceNone MenuChoice = iota // user quit without picking
	ChoiceScanAndPick                // open the full TUI
	ChoiceQuickCleanSafe             // scan + delete safe items + summary
	ChoiceBrowseByCategory           // pick a category, then scan that one only
	ChoiceLastLog                    // show the last cleanup log
	ChoiceAbout                      // show keys / safety / docs
)

// menuItem is one row in the welcome menu.
type menuItem struct {
	choice MenuChoice
	label  string
	hint   string
}

var menuItems = []menuItem{
	{ChoiceScanAndPick, "Quick scan & pick", "Scan everything and choose what to clean"},
	{ChoiceQuickCleanSafe, "Quick clean (safe defaults)", "Scan and clean only items marked ✓ safe"},
	{ChoiceBrowseByCategory, "Browse by category", "Pick a single category to scan"},
	{ChoiceLastLog, "Last cleanup log", "See what was deleted last time"},
	{ChoiceAbout, "About / help", "Keys, safety model, docs"},
	{ChoiceNone, "Quit", "Exit without doing anything"},
}

// WelcomeModel is a Bubble Tea model that shows the entry-point menu.
type WelcomeModel struct {
	cursor       int
	chosen       MenuChoice
	cancelled    bool
	categoryView bool        // when true, we're in the second menu (categories)
	categories   []rules.Category
	pickedCat    string
}

func NewWelcome(catalog *rules.Catalog) WelcomeModel {
	return WelcomeModel{categories: catalog.Categories}
}

// Chosen returns the menu choice the user committed to (after Run completes).
// Returns ChoiceNone if the user quit.
func (m WelcomeModel) Chosen() MenuChoice { return m.chosen }

// PickedCategory returns the category id when Chosen() == ChoiceBrowseByCategory.
func (m WelcomeModel) PickedCategory() string { return m.pickedCat }

func (m WelcomeModel) Init() tea.Cmd { return nil }

func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.categoryView {
		return m.updateCategoryView(keyMsg)
	}
	return m.updateMainView(keyMsg)
}

func (m WelcomeModel) updateMainView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
	case "enter", " ":
		picked := menuItems[m.cursor]
		if picked.choice == ChoiceNone {
			m.cancelled = true
			return m, tea.Quit
		}
		if picked.choice == ChoiceBrowseByCategory {
			m.categoryView = true
			m.cursor = 0
			return m, nil
		}
		m.chosen = picked.choice
		return m, tea.Quit
	}
	return m, nil
}

func (m WelcomeModel) updateCategoryView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		// Back to main menu, not a full quit.
		m.categoryView = false
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.categories)-1 {
			m.cursor++
		}
	case "enter", " ":
		m.pickedCat = m.categories[m.cursor].ID
		m.chosen = ChoiceBrowseByCategory
		return m, tea.Quit
	}
	return m, nil
}

func (m WelcomeModel) View() string {
	if m.cancelled {
		return ""
	}
	if m.categoryView {
		return m.viewCategoryPick()
	}
	return m.viewMain()
}

var (
	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2)
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	cursorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
)

func (m WelcomeModel) viewMain() string {
	var b strings.Builder
	b.WriteString(bannerStyle.Render("🧹 reclaim — developer-aware Mac cleaner"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render(
		"Find and free disk space taken by build caches, package managers,\n" +
			"IDE artifacts, and stale project dependencies."))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Nothing is deleted without your explicit confirmation."))
	b.WriteString("\n\n")
	b.WriteString("What would you like to do?\n\n")

	// Determine longest label for column alignment.
	maxLabel := 0
	for _, it := range menuItems {
		if len(it.label) > maxLabel {
			maxLabel = len(it.label)
		}
	}

	for i, it := range menuItems {
		marker := "  "
		label := it.label
		hint := it.hint
		line := fmt.Sprintf("%s%-*s    %s", marker, maxLabel, label, hint)
		if i == m.cursor {
			line = cursorStyle.Render(fmt.Sprintf("▶ %-*s", maxLabel, label)) + "    " + hintStyle.Render(hint)
		} else {
			line = itemStyle.Render(label) + strings.Repeat(" ", maxLabel-len(label)+4) + hintStyle.Render(hint)
		}
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ select   enter confirm   q quit"))
	return b.String()
}

func (m WelcomeModel) viewCategoryPick() string {
	var b strings.Builder
	b.WriteString(bannerStyle.Render("🧹 reclaim — pick a category"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Scan only this category. Faster than the full scan."))
	b.WriteString("\n\n")

	maxLabel := 0
	for _, c := range m.categories {
		if len(c.Name) > maxLabel {
			maxLabel = len(c.Name)
		}
	}
	for i, c := range m.categories {
		label := c.Name
		hint := c.Description
		var line string
		if i == m.cursor {
			line = cursorStyle.Render(fmt.Sprintf("▶ %-*s", maxLabel, label)) + "    " + hintStyle.Render(hint)
		} else {
			line = itemStyle.Render(label) + strings.Repeat(" ", maxLabel-len(label)+4) + hintStyle.Render(hint)
		}
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ select   enter confirm   esc back"))
	return b.String()
}
