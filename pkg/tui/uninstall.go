package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// UninstallModel represents the TUI model for the uninstall process.
type UninstallModel struct {
	cursor    int
	step      int
	choice    string
	confirmed bool
	quitting  bool
}

// InitialUninstallModel creates a new uninstall TUI model with default values.
func InitialUninstallModel() UninstallModel {
	return UninstallModel{
		step: 0,
	}
}

// Init initializes the uninstall TUI model.
func (m UninstallModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the uninstall TUI model state.
func (m UninstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < 1 {
				m.cursor++
			}

		case "enter":
			if m.step == 0 {
				if m.cursor == 0 {
					m.choice = "nix-foundry"
				} else {
					m.choice = "all"
				}
				m.step++
				m.cursor = 0
			} else {
				m.confirmed = m.cursor == 0
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the uninstall TUI interface.
func (m UninstallModel) View() string {
	s := "\n"

	if m.step == 0 {
		s += ColorBold + ColorCyan + "Choose what to uninstall:" + ColorReset + "\n\n"
		choices := []string{
			"Remove Nix Foundry only (keeps Nix installation)",
			"Remove Nix Foundry and uninstall Nix",
		}
		for i, choice := range choices {
			cursor := " "
			if m.cursor == i {
				cursor = ColorCyan + ">" + ColorReset
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else {
		s += ColorBold + ColorCyan + "Confirmation" + ColorReset + "\n"
		s += "============\n\n"

		s += ColorBold + ColorCyan + "The following will be removed:" + ColorReset + "\n"
		if m.choice == "nix-foundry" {
			s += "• Nix Foundry configuration and files\n"
			s += "• Shell configuration for Nix Foundry\n"
			s += ColorGreen + "\nNote: Your Nix installation will be preserved." + ColorReset + "\n"
		} else {
			s += "• Nix Foundry configuration and files\n"
			s += "• Shell configuration for Nix Foundry\n"
			s += "• Nix package manager and daemon services\n"
			s += "• All packages installed through Nix\n"
			s += "• Nix store directory (/nix)\n"
			s += "• User Nix profiles and channels\n"
			s += "• System and user shell configurations\n"
			s += "• Nix-related cache and configuration files\n"
			s += ColorYellow + "\nWarning: This will completely remove Nix and all associated data." + ColorReset + "\n"
			s += ColorRed + "This action cannot be undone!" + ColorReset + "\n"
		}
		s += "\n"
		s += ColorCyan + "Are you sure you want to proceed?" + ColorReset + "\n\n"

		for i, choice := range []string{"Yes", "No"} {
			cursor := " "
			if m.cursor == i {
				cursor = ColorCyan + ">" + ColorReset
			}
			if i == 0 {
				s += fmt.Sprintf("%s %s%s%s\n", cursor, ColorRed, choice, ColorReset)
			} else {
				s += fmt.Sprintf("%s %s\n", cursor, choice)
			}
		}
	}

	s += "\n(use arrow keys to navigate, enter to select)\n"

	return s
}

// RunUninstallTUI runs the uninstall TUI and returns user choices.
func RunUninstallTUI() (bool, bool, error) {
	p := tea.NewProgram(InitialUninstallModel())
	m, err := p.Run()
	if err != nil {
		return false, false, fmt.Errorf("failed to run TUI: %w", err)
	}

	finalModel := m.(UninstallModel)
	if finalModel.quitting {
		return false, false, fmt.Errorf("uninstallation cancelled")
	}

	uninstallNix := finalModel.choice == "all"
	return uninstallNix, finalModel.confirmed, nil
}
