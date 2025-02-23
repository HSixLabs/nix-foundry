// Package tui provides terminal user interface components.
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices   []string
	cursor    int
	selected  map[int]struct{}
	step      int
	manager   string
	shell     string
	confirmed bool
	quitting  bool
}

func getCurrentShell() string {
	shell := os.Getenv("SHELL")
	return filepath.Base(shell)
}

func InitialModel() model {
	return model{
		choices: []string{
			"tmux (terminal multiplexer)",
			"git (version control)",
			"fzf (fuzzy finder)",
			"ripgrep (fast search)",
			"fd (find alternative)",
			"jq (JSON processor)",
			"yq (YAML processor)",
			"htop (process viewer)",
			"neovim (text editor)",
			"tree (directory viewer)",
			"bat (cat alternative)",
			"exa (ls alternative)",
			"delta (git diff viewer)",
			"direnv (environment manager)",
			"gh (GitHub CLI)",
			"starship (shell prompt)",
			"lazygit (git TUI)",
			"bottom (system monitor)",
		},
		selected: make(map[int]struct{}),
		manager:  "nix-env",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			switch m.step {
			case 0:
				if m.cursor < 2 {
					m.cursor++
				}
			case 1:
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case 2:
				if m.cursor < 1 {
					m.cursor++
				}
			}

		case "enter":
			switch m.step {
			case 0:
				shells := []string{"bash", "zsh", "fish"}
				m.shell = shells[m.cursor]
				m.step++
				m.cursor = 0

			case 1:
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}

			case 2:
				m.confirmed = m.cursor == 0
				return m, tea.Quit
			}

		case "tab", " ":
			if m.step == 1 {
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}

		case "right", "l":
			if m.step == 1 {
				m.step++
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "\n"

	switch m.step {
	case 0:
		s += "Choose shell:\n\n"
		currentShell := getCurrentShell()
		shells := []string{"bash", "zsh", "fish"}
		for i, shell := range shells {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			current := ""
			if shell == currentShell {
				current = " (current)"
			}
			s += fmt.Sprintf("%s %s%s\n", cursor, shell, current)
		}

	case 1:
		s += "Select optional packages (space to select, right arrow to continue):\n\n"
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "x"
			}
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}

	case 2:
		s += "Configuration Summary:\n"
		s += fmt.Sprintf("Package Manager: %s\n", m.manager)
		s += fmt.Sprintf("Shell: %s\n", m.shell)
		s += "Core Packages: git, curl, wget\n"
		if len(m.selected) > 0 {
			s += "Optional Packages:\n"
			for i := range m.selected {
				s += fmt.Sprintf("  - %s\n", strings.Split(m.choices[i], " ")[0])
			}
		}
		s += "\nProceed with installation?\n\n"
		for i, choice := range []string{"Yes", "No"} {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	}

	s += "\n(use arrow keys to navigate, enter to select)\n"
	if m.step == 1 {
		s += "(space to toggle selection, right arrow to continue)\n"
	}

	return s
}

// RunInstallTUI runs the installation TUI and returns the user's choices.
func RunInstallTUI() (string, string, []string, bool, error) {
	p := tea.NewProgram(InitialModel())
	m, err := p.Run()
	if err != nil {
		return "", "", nil, false, fmt.Errorf("failed to run TUI: %w", err)
	}

	finalModel := m.(model)
	if finalModel.quitting {
		return "", "", nil, false, fmt.Errorf("installation cancelled")
	}

	var packages []string
	for i := range finalModel.selected {
		pkg := strings.Split(finalModel.choices[i], " ")[0]
		packages = append(packages, pkg)
	}

	return finalModel.manager, finalModel.shell, packages, finalModel.confirmed, nil
}
