package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressItem represents a single install operation.
type ProgressItem struct {
	Label    string
	Detail   string
	Status   ProgressStatus
}

// ProgressStatus tracks the state of an operation.
type ProgressStatus int

const (
	StatusPending ProgressStatus = iota
	StatusActive
	StatusDone
	StatusError
)

// ProgressModel displays install/update progress.
type ProgressModel struct {
	Title string
	Items []ProgressItem
	done  bool
}

// NewProgress creates a new progress view.
func NewProgress(title string, items []ProgressItem) ProgressModel {
	return ProgressModel{
		Title: title,
		Items: items,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m ProgressModel) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render(m.Title))
	b.WriteString("\n\n")

	completed := 0
	for _, item := range m.Items {
		var indicator string
		switch item.Status {
		case StatusDone:
			indicator = DoneIndicator
			completed++
		case StatusActive:
			indicator = ActiveIndicator
		case StatusError:
			indicator = ErrorStyle.Render("!")
		default:
			indicator = PendingIndicator
		}

		label := item.Label
		detail := lipgloss.NewStyle().Foreground(DarkGray).Render(item.Detail)
		b.WriteString(fmt.Sprintf("  %s %s  %s\n", indicator, label, detail))
	}

	// Progress bar
	total := len(m.Items)
	if total > 0 {
		b.WriteString("\n")
		barWidth := 20
		filled := (completed * barWidth) / total
		bar := strings.Repeat("=", filled) + strings.Repeat("-", barWidth-filled)
		progress := lipgloss.NewStyle().Foreground(Cyan).Render(fmt.Sprintf("  [%s] %d/%d", bar, completed, total))
		b.WriteString(progress)
	}

	return b.String()
}

// PrintProgress outputs a static progress view (non-interactive).
func PrintProgress(title string, items []ProgressItem) {
	m := NewProgress(title, items)
	fmt.Println(m.View())
}
