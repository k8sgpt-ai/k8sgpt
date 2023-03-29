package analyze

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	WindowSize    tea.WindowSizeMsg
	tableStyle    = lipgloss.NewStyle().Align(lipgloss.Top, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
	viewportStyle = lipgloss.NewStyle().Align(lipgloss.Bottom, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
)

type model struct {
	table          table.Model
	explainView    viewport.Model
	rows           []table.Row
	explainEnabled bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, cmd2 tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		WindowSize = msg
		m.table.SetWidth(WindowSize.Width)
		m.table.SetHeight(WindowSize.Height / 2)
		tableStyle = lipgloss.NewStyle().Width(WindowSize.Width).Height(WindowSize.Height/2).Align(lipgloss.Top, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
		viewportStyle = lipgloss.NewStyle().Width(WindowSize.Width).Height(WindowSize.Height/3).Align(lipgloss.Bottom, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
		m.explainView = viewport.New(WindowSize.Width, WindowSize.Height/3)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.explainEnabled {
				m.explainView.SetContent(m.table.SelectedRow()[4])
			}
		default:

		}
	}

	m.table, cmd = m.table.Update(msg)
	m.explainView, cmd2 = m.explainView.Update(msg)

	return m, tea.Batch(cmd, cmd2)
}

func (m model) View() string {
	var s string
	if m.explainEnabled {
		s += lipgloss.JoinVertical(lipgloss.Top, tableStyle.Render(m.table.View()), viewportStyle.Render(m.explainView.View()))
	} else {
		s += tableStyle.Render(m.table.View())
	}

	return s
}

func InitTable(rows []table.Row, explainEnabled bool) (tea.Model, tea.Cmd) {
	t := table.New(table.WithColumns([]table.Column{
		{Title: "ID", Width: 3},
		{Title: "Resource", Width: 20},
		{Title: "Fix", Width: 3},
		{Title: "Error", Width: 100},
		{Title: "Details", Width: 0},
	}), table.WithRows(rows), table.WithFocused(true),
		table.WithHeight(WindowSize.Height/2),
		table.WithWidth(WindowSize.Width))

	var v viewport.Model

	if explainEnabled {
		v := viewport.New(WindowSize.Width, WindowSize.Height/3)
		v.SetContent(t.SelectedRow()[4])
	}
	m := model{
		table:          t,
		rows:           rows,
		explainView:    v,
		explainEnabled: explainEnabled,
	}

	return m, func() tea.Msg { return nil }
}

func Render(rows []table.Row, explainEnabled bool) {

	m, _ := InitTable(rows, explainEnabled)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
