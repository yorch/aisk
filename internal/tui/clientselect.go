package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yorch/aisk/internal/client"
)

// ClientSelectModel is a multi-select picker for AI clients.
type ClientSelectModel struct {
	Title    string
	clients  []*client.Client
	cursor   int
	selected map[int]bool
	done     bool
	quitting bool
}

// NewClientSelect creates a new client multi-select model.
func NewClientSelect(title string, clients []*client.Client) ClientSelectModel {
	// Pre-select all detected clients
	selected := make(map[int]bool)
	for i, c := range clients {
		if c.Detected {
			selected[i] = true
		}
	}

	return ClientSelectModel{
		Title:    title,
		clients:  clients,
		selected: selected,
	}
}

func (m ClientSelectModel) Init() tea.Cmd {
	return nil
}

func (m ClientSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.clients)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "a":
			// Select all
			for i := range m.clients {
				m.selected[i] = true
			}
		case "n":
			// Select none
			for i := range m.clients {
				m.selected[i] = false
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ClientSelectModel) View() string {
	if m.done || m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(TitleStyle.Render(m.Title))
	b.WriteString("\n\n")

	for i, c := range m.clients {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		checked := "[ ]"
		style := UncheckedStyle
		if m.selected[i] {
			checked = "[x]"
			style = CheckedStyle
		}

		name := c.Name
		path := c.GlobalPath
		if path == "" {
			path = c.ProjectPath
		}

		line := fmt.Sprintf("%s %s %s", cursor, checked, name)
		if m.cursor == i {
			line = SelectedStyle.Render(fmt.Sprintf("%s [%s] %s", cursor, func() string {
				if m.selected[i] {
					return "x"
				}
				return " "
			}(), name))
		} else {
			line = style.Render(line)
		}

		pathInfo := lipgloss.NewStyle().Foreground(DarkGray).Render(fmt.Sprintf("  (%s)", path))

		b.WriteString(line + pathInfo + "\n")
	}

	b.WriteString(HelpStyle.Render("\n  space: toggle | enter: confirm | a: all | n: none | q: quit"))

	return b.String()
}

// SelectedClients returns the clients that were selected.
func (m ClientSelectModel) SelectedClients() []*client.Client {
	var result []*client.Client
	for i, c := range m.clients {
		if m.selected[i] {
			result = append(result, c)
		}
	}
	return result
}

// Cancelled returns true if the user quit without confirming.
func (m ClientSelectModel) Cancelled() bool {
	return m.quitting
}

// RunClientSelect runs the interactive client picker and returns selected clients.
func RunClientSelect(title string, clients []*client.Client) ([]*client.Client, error) {
	model := NewClientSelect(title, clients)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(ClientSelectModel)
	if final.Cancelled() {
		return nil, fmt.Errorf("cancelled")
	}

	return final.SelectedClients(), nil
}
