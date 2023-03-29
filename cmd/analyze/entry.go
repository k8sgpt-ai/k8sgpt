package analyze

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
)

type Entry struct {
	viewport   viewport.Model
	currentRow table.Row
	rows       []table.Row
}

func (m Entry) Init() tea.Cmd {
	return nil
}

func (m Entry) View() string {
	return lipgloss.NewStyle().Align(lipgloss.Bottom).Render(m.viewport.View())
}

func (m Entry) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport = viewport.New(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return InitTable(m.rows)
		}

	}
	m.setViewportContent()
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}
func (m *Entry) setViewportContent() {
	m.viewport.SetContent(m.currentRow[4])
}
func InitEntry(currentRow table.Row, rows []table.Row) *Entry {

	currentRow[4] = color.GreenString(currentRow[4])
	m := Entry{
		rows:       rows,
		currentRow: currentRow,
	}
	m.viewport = viewport.New(100, 100)
	m.viewport.Style = lipgloss.NewStyle().Align(lipgloss.Left)
	m.setViewportContent()

	return &m
}
