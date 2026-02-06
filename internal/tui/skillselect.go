package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yorch/aisk/internal/skill"
)

// SkillSelectModel is a filterable skill browser.
type SkillSelectModel struct {
	skills   []*skill.Skill
	filtered []*skill.Skill
	cursor   int
	filter   string
	done     bool
	quitting bool
}

// NewSkillSelect creates a new skill browser model.
func NewSkillSelect(skills []*skill.Skill) SkillSelectModel {
	return SkillSelectModel{
		skills:   skills,
		filtered: skills,
	}
}

func (m SkillSelectModel) Init() tea.Cmd {
	return nil
}

func (m SkillSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.filtered) > 0 {
				m.done = true
				return m, tea.Quit
			}
		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}
		default:
			if len(msg.String()) == 1 {
				m.filter += msg.String()
				m.applyFilter()
			}
		}
	}

	return m, nil
}

func (m *SkillSelectModel) applyFilter() {
	if m.filter == "" {
		m.filtered = m.skills
	} else {
		var filtered []*skill.Skill
		lower := strings.ToLower(m.filter)
		for _, s := range m.skills {
			if strings.Contains(strings.ToLower(s.Frontmatter.Name), lower) ||
				strings.Contains(strings.ToLower(s.DirName), lower) {
				filtered = append(filtered, s)
			}
		}
		m.filtered = filtered
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m SkillSelectModel) View() string {
	if m.done || m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(TitleStyle.Render("Select a skill"))
	b.WriteString("\n\n")

	// Filter bar
	if m.filter != "" {
		filterDisplay := lipgloss.NewStyle().Foreground(Yellow).Render("/" + m.filter)
		b.WriteString(filterDisplay + "\n\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(Gray).Render("  No matching skills.\n"))
	} else {
		for i, s := range m.filtered {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			name := s.Frontmatter.Name
			version := lipgloss.NewStyle().Foreground(DarkGray).Render(s.DisplayVersion())

			if m.cursor == i {
				name = SelectedStyle.Render(name)
				line := fmt.Sprintf("%s %s  %s", cursor, name, version)
				b.WriteString(line + "\n")
			} else {
				line := fmt.Sprintf("%s %s  %s", cursor, name, version)
				b.WriteString(line + "\n")
			}
		}
	}

	b.WriteString(HelpStyle.Render("\n  type to filter | enter: select | esc: quit"))

	return b.String()
}

// SelectedSkill returns the skill that was selected, or nil.
func (m SkillSelectModel) SelectedSkill() *skill.Skill {
	if m.done && m.cursor < len(m.filtered) {
		return m.filtered[m.cursor]
	}
	return nil
}

// Cancelled returns true if the user quit without selecting.
func (m SkillSelectModel) Cancelled() bool {
	return m.quitting
}

// RunSkillSelect runs the interactive skill browser.
func RunSkillSelect(skills []*skill.Skill) (*skill.Skill, error) {
	model := NewSkillSelect(skills)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(SkillSelectModel)
	if final.Cancelled() {
		return nil, fmt.Errorf("cancelled")
	}

	selected := final.SelectedSkill()
	if selected == nil {
		return nil, fmt.Errorf("no skill selected")
	}

	return selected, nil
}
