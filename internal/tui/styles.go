package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Purple   = lipgloss.Color("99")
	Cyan     = lipgloss.Color("86")
	Green    = lipgloss.Color("42")
	Yellow   = lipgloss.Color("214")
	Red      = lipgloss.Color("196")
	Gray     = lipgloss.Color("245")
	DarkGray = lipgloss.Color("239")
	White    = lipgloss.Color("255")

	// Title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Purple).
			MarginBottom(1)

	// Subtitle / description
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			MarginBottom(1)

	// Selected item
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true)

	// Checked item
	CheckedStyle = lipgloss.NewStyle().
			Foreground(Green)

	// Unchecked item
	UncheckedStyle = lipgloss.NewStyle().
			Foreground(DarkGray)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray).
			MarginTop(1)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	// Status indicators
	DoneIndicator    = SuccessStyle.Render("*")
	PendingIndicator = UncheckedStyle.Render("-")
	ActiveIndicator  = lipgloss.NewStyle().Foreground(Yellow).Render("o")
)
