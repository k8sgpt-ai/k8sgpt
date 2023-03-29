package analyze

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	WindowSize tea.WindowSizeMsg
)

type MainModel struct {
}

type model struct {
	table table.Model
	rows  []table.Row
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowSize = msg
		m.table.SetWidth(WindowSize.Width)
		m.table.SetHeight(WindowSize.Height - (WindowSize.Height / 9))
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
			entry := InitEntry(m.table.SelectedRow(), m.rows)
			return entry.Update(WindowSize)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func InitTable(rows []table.Row) (tea.Model, tea.Cmd) {
	t := table.New(table.WithColumns([]table.Column{
		{Title: "ID", Width: 3},
		{Title: "Resource", Width: 20},
		{Title: "Fix", Width: 3},
		{Title: "Error", Width: 100},
		{Title: "Details", Width: 0},
	}), table.WithRows(rows), table.WithFocused(true),
		table.WithHeight(WindowSize.Height-(WindowSize.Height/9)),
		table.WithWidth(WindowSize.Width))

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	m := model{
		table: t,
		rows:  rows,
	}

	return m, func() tea.Msg { return nil }
}

func Render(rows []table.Row) {

	m, _ := InitTable(rows)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
