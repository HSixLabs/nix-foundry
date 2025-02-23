// Package tui provides terminal user interface components.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type uninstallModel struct {
	cursor    int
	step      int
	choice    string
	confirmed bool
	quitting  bool
}

func InitialUninstallModel() uninstallModel {
	return uninstallModel{
		step: 0,
	}
}

func (m uninstallModel) Init() tea.Cmd {
	return nil
}

func (m uninstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m uninstallModel) View() string {
	s := "\n"

	if m.step == 0 {
		s += "Choose what to uninstall:\n\n"
		choices := []string{
			"Remove Nix Foundry only (keeps Nix installation)",
			"Remove Nix Foundry and uninstall Nix",
		}
		for i, choice := range choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else {
		s += "Are you sure?\n\n"
		if m.choice == "nix-foundry" {
			s += "This will remove Nix Foundry.\n"
		} else {
			s += "This will remove Nix Foundry and uninstall Nix.\n"
		}
		s += "\n"
		for i, choice := range []string{"Yes", "No"} {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	}

	s += "\n(use arrow keys to navigate, enter to select)\n"

	return s
}

// RunUninstallTUI runs the uninstallation TUI and returns the user's choices.
func RunUninstallTUI() (bool, bool, error) {
	p := tea.NewProgram(InitialUninstallModel())
	m, err := p.Run()
	if err != nil {
		return false, false, fmt.Errorf("failed to run TUI: %w", err)
	}

	finalModel := m.(uninstallModel)
	if finalModel.quitting {
		return false, false, fmt.Errorf("uninstallation cancelled")
	}

	uninstallNix := finalModel.choice == "all"
	return uninstallNix, finalModel.confirmed, nil
}
