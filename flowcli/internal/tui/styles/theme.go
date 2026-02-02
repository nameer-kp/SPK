package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#10B981")
	Muted     = lipgloss.Color("#6B7280")
	Error     = lipgloss.Color("#EF4444")
	Success   = lipgloss.Color("#10B981")
	Warning   = lipgloss.Color("#F59E0B")

	// Text styles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	Subtitle = lipgloss.NewStyle().
			Foreground(Muted)

	StatusSuccess = lipgloss.NewStyle().
			Foreground(Success)

	StatusError = lipgloss.NewStyle().
			Foreground(Error)

	StatusPending = lipgloss.NewStyle().
			Foreground(Muted)

	StatusRunning = lipgloss.NewStyle().
			Foreground(Warning)

	// Layout styles
	Container = lipgloss.NewStyle().
			Padding(1, 2)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Muted).
		Padding(1, 2)

	// Menu styles
	MenuItem = lipgloss.NewStyle().
			PaddingLeft(2)

	MenuItemSelected = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(Primary).
				Bold(true)

	// Help text
	Help = lipgloss.NewStyle().
		Foreground(Muted).
		MarginTop(1)
)
