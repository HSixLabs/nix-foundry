/*
Package tui provides terminal user interface components for Nix Foundry.
It implements interactive terminal interfaces for installation and configuration
using the Bubble Tea framework.
*/
package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

/*
Model represents the TUI model for installation.
It maintains the state of the installation wizard including user selections,
navigation state, and configuration options.
*/
type Model struct {
	languageChoices    []string
	editorChoices      []string
	devToolChoices     []string
	cursor             int
	selected           map[string]map[int]struct{}
	step               int
	manager            string
	shell              string
	confirmed          bool
	quitting           bool
	skipWizard         bool
	chooseOwnLanguages bool
	chooseOwnEditors   bool
	chooseOwnDevTools  bool
}

/*
getCurrentShell retrieves the current user's shell from the environment.
*/
func getCurrentShell() string {
	shell := os.Getenv("SHELL")
	return filepath.Base(shell)
}

/*
InitialModel creates and returns the initial installation model with default values.
*/
func InitialModel() Model {
	return Model{
		languageChoices: []string{
			"Python (python3 + pip)",
			"Node.js (nodejs + npm)",
			"Golang (go)",
			"Java (openjdk + maven)",
			"C/C++ (gcc + make)",
			"I'll choose my own",
		},
		editorChoices: []string{
			"VS Code",
			"Sublime Text",
			"IntelliJ IDEA",
			"Neovim",
			"GNU Emacs",
			"I'll choose my own",
		},
		devToolChoices: []string{
			"Git",
			"Docker",
			"Kubernetes CLI (kubectl)",
			"Terraform",
			"GitHub CLI",
			"I'll choose my own",
		},
		selected: map[string]map[int]struct{}{
			"languages": make(map[int]struct{}),
			"editors":   make(map[int]struct{}),
			"devtools":  make(map[int]struct{}),
		},
		manager: "nix-env",
	}
}

// Init initializes the TUI model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

/*
getMaxCursor returns the maximum cursor value for the current step.
*/
func (m Model) getMaxCursor() int {
	switch m.step {
	case 0:
		return 1
	case 1:
		if !m.skipWizard {
			return 3
		}
	case 2:
		return len(m.languageChoices) - 1
	case 3:
		return len(m.editorChoices) - 1
	case 4:
		return len(m.devToolChoices) - 1
	case 5:
		return 1
	}
	return 0
}

/*
handleKeyPress processes keyboard input and updates the model state.
*/
func (m Model) handleKeyPress(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		maxCursor := m.getMaxCursor()
		if m.cursor < maxCursor {
			m.cursor++
		}

	case "enter":
		m = m.handleEnter()
		if m.step == 5 && m.confirmed {
			return m, tea.Quit
		}

	case "tab", " ":
		if m.step >= 2 && m.step <= 4 {
			m.handleSelection()
		}

	case "right", "l":
		if m.step >= 2 && m.step <= 4 {
			m.step++
			m.cursor = 0
		}

	case "left", "h":
		if m.step >= 1 && m.step <= 5 && !m.skipWizard {
			m.step--
			m.cursor = 0
		}
	}

	return m, nil
}

/*
handleEnter processes the enter key press based on the current step.
*/
func (m Model) handleEnter() Model {
	switch m.step {
	case 0:
		m.skipWizard = m.cursor == 1
		m.step++
		m.cursor = 0
		if m.skipWizard {
			m.step = 5
		}

	case 1:
		if !m.skipWizard {
			shells := []string{"bash", "zsh", "fish", "custom"}
			m.shell = shells[m.cursor]
			m.step++
			m.cursor = 0
		}

	case 2, 3, 4:
		m.handleSelection()

	case 5:
		m.confirmed = m.cursor == 0
	}

	return m
}

// Update handles TUI messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

/*
handleSelection processes selection changes for languages, editors, and dev tools.
*/
func (m *Model) handleSelection() {
	var category string
	var choices []string
	var chooseOwn *bool
	var selections map[int]struct{}

	switch m.step {
	case 2:
		category = "languages"
		choices = m.languageChoices
		chooseOwn = &m.chooseOwnLanguages
		selections = m.selected[category]
	case 3:
		category = "editors"
		choices = m.editorChoices
		chooseOwn = &m.chooseOwnEditors
		selections = m.selected[category]
	case 4:
		category = "devtools"
		choices = m.devToolChoices
		chooseOwn = &m.chooseOwnDevTools
		selections = m.selected[category]
	}

	lastIndex := len(choices) - 1
	if m.cursor == lastIndex {
		m.handleChooseOwnSelection(category, chooseOwn, lastIndex)
	} else {
		m.handleRegularSelection(selections, chooseOwn)
	}
}

/*
handleChooseOwnSelection handles the "I'll choose my own" option selection.
*/
func (m *Model) handleChooseOwnSelection(category string, chooseOwn *bool, lastIndex int) {
	*chooseOwn = !*chooseOwn
	if *chooseOwn {
		m.selected[category] = make(map[int]struct{})
		m.selected[category][lastIndex] = struct{}{}
	} else {
		delete(m.selected[category], lastIndex)
	}
}

/*
handleRegularSelection handles regular option selection.
*/
func (m *Model) handleRegularSelection(selections map[int]struct{}, chooseOwn *bool) {
	if !*chooseOwn {
		if _, ok := selections[m.cursor]; ok {
			delete(selections, m.cursor)
		} else {
			selections[m.cursor] = struct{}{}
		}
	}
}

/*
renderWelcomeScreen renders the initial welcome screen.
*/
func (m Model) renderWelcomeScreen() string {
	s := ColorBold + ColorCyan + "Welcome to Nix Foundry!" + ColorReset + "\n\n"
	s += ColorCyan + "How would you like to configure your environment?" + ColorReset + "\n\n"

	options := []string{
		"Use the installation wizard (recommended for new users)",
		"Configure manually using the CLI or config file",
	}
	for i, option := range options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, option)
	}

	s += "\nNote: You can always modify your configuration later by:\n"
	s += "  • Using CLI commands (run 'nix-foundry --help')\n"
	s += "  • Editing ~/.config/nix-foundry/config.yaml directly\n"

	return s
}

/*
renderShellSelection renders the shell selection screen.
*/
func (m Model) renderShellSelection() string {
	s := ColorCyan + "Choose shell:" + ColorReset + "\n\n"
	currentShell := getCurrentShell()
	shells := []string{"bash", "zsh", "fish", "I'll choose my own"}
	for i, shell := range shells {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		current := ""
		if shell == currentShell {
			current = " " + ColorYellow + "(current)" + ColorReset
		}
		s += fmt.Sprintf("%s %s%s\n", cursor, shell, current)
	}
	return s
}

/*
renderChoiceList renders a list of choices with selection indicators.
*/
func (m Model) renderChoiceList(title string, choices []string, category string, chooseOwn bool) string {
	s := ColorCyan + title + ColorReset + " (space to select, right arrow to continue):\n\n"
	for i, choice := range choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		checked := " "
		if _, ok := m.selected[category][i]; ok {
			checked = ColorGreen + "x" + ColorReset
		}
		if chooseOwn && i != len(choices)-1 {
			s += fmt.Sprintf("%s [%s] %s%s%s\n", cursor, checked, ColorGrey, choice, ColorReset)
		} else {
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		}
	}
	return s
}

/*
renderInstallationSummary renders the final installation summary screen.
*/
func (m Model) renderInstallationSummary() string {
	s := ColorBold + ColorCyan + "Installation Summary" + ColorReset + "\n"
	s += "===================\n\n"

	s += m.renderPackageManagerSection()
	s += m.renderConfigurationSection()
	s += m.renderPostInstallationSection()

	s += ColorCyan + "Proceed with installation?" + ColorReset + "\n\n"
	for i, choice := range []string{"Yes", "No"} {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return s
}

/*
renderPackageManagerSection renders the package manager setup section of the summary.
*/
func (m Model) renderPackageManagerSection() string {
	s := ColorBold + ColorCyan + "1. Package Manager Setup:" + ColorReset + "\n"
	s += "   • Will install and configure Nix package manager\n"
	s += "   • Will create initial Nix configuration\n\n"
	return s
}

/*
renderConfigurationSection renders the configuration section of the summary.
*/
func (m Model) renderConfigurationSection() string {
	if m.skipWizard {
		return m.renderBasicConfigurationSection()
	}
	return m.renderDetailedConfigurationSection()
}

/*
renderBasicConfigurationSection renders the basic configuration section.
*/
func (m Model) renderBasicConfigurationSection() string {
	s := ColorBold + ColorCyan + "2. Configuration:" + ColorReset + "\n"
	s += "   • Basic installation with no initial configuration\n"
	s += "   • You can configure your environment later using:\n"
	s += "     - CLI commands (run 'nix-foundry --help')\n"
	s += "     - Config file at ~/.config/nix-foundry/config.yaml\n\n"
	return s
}

/*
renderDetailedConfigurationSection renders the detailed configuration section.
*/
func (m Model) renderDetailedConfigurationSection() string {
	s := ColorBold + ColorCyan + "2. Shell Configuration:" + ColorReset + "\n"
	if m.shell == "custom" {
		s += "   • You'll configure your shell later\n"
	} else {
		s += fmt.Sprintf("   • Will configure %s as your shell\n", m.shell)
	}
	s += "\n"

	s += ColorBold + ColorCyan + "3. Package Installation:" + ColorReset + "\n"
	s += m.renderPackageSection("Languages", "languages", m.languageChoices)
	s += m.renderPackageSection("Editors", "editors", m.editorChoices)
	s += m.renderPackageSection("Developer Tools", "devtools", m.devToolChoices)

	return s
}

/*
renderPackageSection renders a section of selected packages.
*/
func (m Model) renderPackageSection(title, category string, choices []string) string {
	s := fmt.Sprintf("   %s:", title)
	if _, ok := m.selected[category][len(choices)-1]; ok {
		s += " Will configure later\n"
	} else if len(m.selected[category]) > 0 {
		s += "\n"
		var selected []string
		for i := 0; i < len(choices)-1; i++ {
			if _, ok := m.selected[category][i]; ok {
				selected = append(selected, choices[i])
			}
		}
		for _, item := range selected {
			s += fmt.Sprintf("   • %s\n", item)
		}
	} else {
		s += " None selected\n"
	}
	s += "\n"
	return s
}

/*
renderPostInstallationSection renders the post-installation section.
*/
func (m Model) renderPostInstallationSection() string {
	s := ColorBold + ColorCyan + "4. Post-Installation:" + ColorReset + "\n"
	s += "   • Will create configuration file at ~/.config/nix-foundry/config.yaml\n"
	s += "   • You can modify your configuration at any time using:\n"
	s += "     - CLI commands (run 'nix-foundry --help')\n"
	s += "     - Editing the config file directly\n\n"
	return s
}

/*
renderNavigationHelp renders navigation help text.
*/
func (m Model) renderNavigationHelp() string {
	s := "(use arrow keys to navigate, enter to select)\n"
	if m.step >= 1 && m.step <= 4 && !m.skipWizard {
		s += "(left arrow to go back, "
		if m.step >= 2 && m.step <= 4 {
			s += "space to toggle selection, "
		}
		s += "right arrow to continue)\n"
	}
	return s
}

// View renders the TUI interface and returns the display string.
func (m Model) View() string {
	s := "\n"

	switch m.step {
	case 0:
		s += m.renderWelcomeScreen()
	case 1:
		s += m.renderShellSelection()
	case 2:
		s += m.renderChoiceList("Select programming languages", m.languageChoices, "languages", m.chooseOwnLanguages)
	case 3:
		s += m.renderChoiceList("Select editors", m.editorChoices, "editors", m.chooseOwnEditors)
	case 4:
		s += m.renderChoiceList("Select developer tools", m.devToolChoices, "devtools", m.chooseOwnDevTools)
	case 5:
		s += m.renderInstallationSummary()
	}

	s += "\n" + m.renderNavigationHelp()
	return s
}

/*
getPackageName returns the platform-specific package name for a display name.
*/
func getPackageName(displayName string) string {
	switch {
	case strings.HasPrefix(displayName, "Python"):
		return "python3"
	case strings.HasPrefix(displayName, "Node.js"):
		return "nodejs"
	case strings.HasPrefix(displayName, "Go"):
		return "go"
	case strings.HasPrefix(displayName, "Java"):
		return "openjdk"
	case strings.HasPrefix(displayName, "C/C++"):
		return "gcc"
	case strings.HasPrefix(displayName, "VS Code"):
		return "vscode"
	case strings.HasPrefix(displayName, "IntelliJ"):
		return "jetbrains.idea-community"
	case strings.HasPrefix(displayName, "Kubernetes CLI"):
		return "kubectl"
	case strings.HasPrefix(displayName, "GitHub CLI"):
		return "gh"
	case displayName == "Docker":
		return "docker"
	case displayName == "Terraform":
		return "terraform"
	case displayName == "Git":
		return "git"
	default:
		return strings.ToLower(strings.Split(displayName, " ")[0])
	}
}

/*
RunInstallTUI runs the installation TUI and returns the user's choices.
Returns the selected package manager, shell, packages, and confirmation status.
*/
func RunInstallTUI() (string, string, []string, bool, error) {
	p := tea.NewProgram(InitialModel())
	m, err := p.Run()
	if err != nil {
		return "", "", nil, false, fmt.Errorf("failed to run TUI: %w", err)
	}

	finalModel := m.(Model)
	if finalModel.quitting {
		return "", "", nil, false, fmt.Errorf("installation cancelled")
	}

	var packages []string
	for category, selections := range finalModel.selected {
		var choices []string
		switch category {
		case "languages":
			choices = finalModel.languageChoices
		case "editors":
			choices = finalModel.editorChoices
		case "devtools":
			choices = finalModel.devToolChoices
		}
		for i := range selections {
			if i == len(choices)-1 {
				continue
			}
			packages = append(packages, getPackageName(choices[i]))
		}
	}

	return finalModel.manager, finalModel.shell, packages, finalModel.confirmed, nil
}
