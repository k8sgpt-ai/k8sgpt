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

type MainModel struct {
}

type model struct {
	table table.Model
	view  viewport.Model
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
		m.table.SetHeight(WindowSize.Height / 2)
		tableStyle = lipgloss.NewStyle().Width(WindowSize.Width).Height(WindowSize.Height/2).Align(lipgloss.Top, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
		viewportStyle = lipgloss.NewStyle().Width(WindowSize.Width).Height(WindowSize.Height/3).Align(lipgloss.Bottom, lipgloss.Left).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))
		m.view = viewport.New(WindowSize.Width, WindowSize.Height/3)

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
			//m.view.SetContent(m.table.SelectedRow()[4])
		default:
			m.view.SetContent(m.table.SelectedRow()[4])
		}
	}
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m model) View() string {
	var s string
	//s += tableStyle.Render(m.table.View())
	s += lipgloss.JoinVertical(lipgloss.Top, tableStyle.Render(m.table.View()), viewportStyle.Render(m.view.View()))

	//	s += viewportStyle.Render(m.view.View())
	return s
}

func InitTable(rows []table.Row) (tea.Model, tea.Cmd) {
	t := table.New(table.WithColumns([]table.Column{
		{Title: "ID", Width: 3},
		{Title: "Resource", Width: 20},
		{Title: "Fix", Width: 3},
		{Title: "Error", Width: 100},
		{Title: "Details", Width: 0},
	}), table.WithRows(rows), table.WithFocused(true),
		table.WithHeight(WindowSize.Height/2),
		table.WithWidth(WindowSize.Width))

	s := table.DefaultStyles()

	t.SetStyles(s)

	v := viewport.New(WindowSize.Width, WindowSize.Height/3)
	v.SetContent(t.SelectedRow()[4])
	m := model{
		table: t,
		rows:  rows,
		view:  v,
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
