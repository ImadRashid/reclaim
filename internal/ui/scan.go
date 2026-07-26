// Interactive scanning screen. The scan itself runs in a goroutine and
// streams events into the Bubble Tea program, so the user sees an animated
// spinner, which rule is being checked, a live tally of what's been found,
// and can cancel with q instead of staring at frozen printf output.
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ImadRashid/reclaim/internal/rules"
	"github.com/ImadRashid/reclaim/internal/scanner"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type scanRuleMsg struct {
	ruleID       string
	index, total int
}
type scanHitsMsg struct{ hits []scanner.Hit }
type scanDoneMsg struct{ hits []scanner.Hit }
type scanTickMsg struct{}

type ScanModel struct {
	version   string
	events    chan tea.Msg
	cancel    context.CancelFunc
	current   string
	index     int
	total     int
	found     int
	foundSize int64
	start     time.Time
	frame     int
	hits      []scanner.Hit
	done      bool
	cancelled bool
}

// RunScan performs a catalog scan behind an interactive progress screen.
// Returns the hits and whether the user cancelled.
func RunScan(cat *rules.Catalog, version string) ([]scanner.Hit, bool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan tea.Msg, 64)
	go func() {
		hits := scanner.ScanEvents(ctx, cat,
			func(ruleID string, index, total int) {
				events <- scanRuleMsg{ruleID: ruleID, index: index, total: total}
			},
			func(hs []scanner.Hit) {
				events <- scanHitsMsg{hits: hs}
			})
		events <- scanDoneMsg{hits: hits}
	}()

	m := ScanModel{version: version, events: events, cancel: cancel, start: time.Now()}
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		cancel()
		return nil, false, err
	}
	sm := final.(ScanModel)
	if sm.cancelled {
		cancel()
		return nil, true, nil
	}
	return sm.hits, false, nil
}

func (m ScanModel) listen() tea.Cmd {
	return func() tea.Msg { return <-m.events }
}

func scanTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return scanTickMsg{} })
}

func (m ScanModel) Init() tea.Cmd {
	return tea.Batch(m.listen(), scanTick())
}

func (m ScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanRuleMsg:
		m.current = msg.ruleID
		m.index = msg.index
		m.total = msg.total
		return m, m.listen()
	case scanHitsMsg:
		m.hits = append(m.hits, msg.hits...)
		for _, h := range msg.hits {
			m.found++
			m.foundSize += h.Size
		}
		return m, m.listen()
	case scanDoneMsg:
		m.hits = msg.hits
		m.done = true
		return m, tea.Quit
	case scanTickMsg:
		m.frame++
		if m.done {
			return m, nil
		}
		return m, scanTick()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			m.cancel()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ScanModel) View() string {
	if m.done || m.cancelled {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("🧹 reclaim v%s — scanning", m.version)))
	b.WriteString("\n\n")

	spin := spinnerFrames[m.frame%len(spinnerFrames)]
	rule := m.current
	if rule == "" {
		rule = "starting…"
	}
	b.WriteString(fmt.Sprintf("  %s checking %s", categoryStyle.Render(spin), rule))
	if m.total > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  (%d/%d)", m.index, m.total)))
	}
	b.WriteString("\n\n")

	elapsed := time.Since(m.start).Round(time.Second)
	b.WriteString(fmt.Sprintf("  found %s across %d items so far",
		totalStyle.Render(formatSize(m.foundSize)), m.found))
	b.WriteString(dimStyle.Render(fmt.Sprintf("  ·  %s elapsed", elapsed)))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  q cancel"))
	return b.String()
}
