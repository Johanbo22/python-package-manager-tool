package tui

import (
	"fmt"

	"github.com/Johanbo22/python-package-manager-tool/internal/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D67F4")).Padding(0, 1).Bold(true)

	modeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#6146AB")).Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0")).MarginTop(1)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).MarginTop(1)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1)
)

type ApplicationState int

const (
	StateViewingList ApplicationState = iota
	StateSearchingPyPi
)

type item struct {
	name    string
	version string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return fmt.Sprintf("Version: %s", i.version) }
func (i item) FilterValue() string { return i.name }

type MainModel struct {
	State         ApplicationState
	PythonClient  *client.PythonBridgeClient
	List          list.Model
	SearchInput   textinput.Model
	Spinner       spinner.Model
	IsLoading     bool
	StatusMessage string
	Err           error
	Width         int
	Height        int
}

func InitialModel() MainModel {
	ti := textinput.New()
	ti.Placeholder = "Type package name to install"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Installed Packages"
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return MainModel{
		State:        StateViewingList,
		PythonClient: client.NewPythonBridgeClient(),
		SearchInput:  ti,
		List:         l,
		Spinner:      s,
		IsLoading:    false,
	}
}

func (m MainModel) Init() tea.Cmd {
	return fetchPackagesCmd(m.PythonClient)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v-4)

	case tea.KeyMsg:
		if m.IsLoading && msg.String() != "ctrl+c" {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			return m, tea.Quit
		case "tab":
			if m.State == StateViewingList {
				m.State = StateSearchingPyPi
				m.StatusMessage = ""
				m.Err = nil
			} else {
				m.State = StateViewingList
				m.StatusMessage = ""
				m.Err = nil
			}
			return m, nil
		}
		if m.State == StateViewingList {
			switch msg.String() {
			case "d":
				if selectedItem, ok := m.List.SelectedItem().(item); ok {
					m.StatusMessage = fmt.Sprintf("Uninstalling %s...", selectedItem.name)
					cmds = append(cmds, deletePackageCmd(m.PythonClient, selectedItem.name), m.Spinner.Tick)
					return m, tea.Batch(cmds...)
				}
			}
		} else if m.State == StateSearchingPyPi {
			switch msg.String() {
			case "enter":
				pkgToInstall := m.SearchInput.Value()
				if pkgToInstall != "" {
					m.StatusMessage = fmt.Sprintf("Installing %s...", pkgToInstall)
					m.SearchInput.SetValue("")
					cmds = append(cmds, installPackageCmd(m.PythonClient, pkgToInstall), m.Spinner.Tick)
					return m, tea.Batch(cmds...)
				}
			}
		}
	case spinner.TickMsg:
		if m.IsLoading {
			var spinCmd tea.Cmd
			m.Spinner, spinCmd = m.Spinner.Update(msg)
			return m, spinCmd
		}
	case packagesMsg:
		items := make([]list.Item, len(msg.packages))
		for i, pkg := range msg.packages {
			items[i] = item{name: pkg.Name, version: pkg.Version}
		}
		cmd = m.List.SetItems(items)
		cmds = append(cmds, cmd)
		m.IsLoading = false
		if m.StatusMessage == "" {
			m.StatusMessage = fmt.Sprintf("Loaded %d packages", len(msg.packages))
		}

	case statusMsg:
		m.StatusMessage = msg.message
		m.IsLoading = false
		m.Err = nil
		cmds = append(cmds, fetchPackagesCmd(m.PythonClient))

	case errMsg:
		m.Err = msg.err
		m.IsLoading = false
		m.StatusMessage = ""
	}

	if m.State == StateViewingList {
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	var mode string
	if m.State == StateViewingList {
		mode = modeStyle.Render("MODE: Browse")
	} else {
		mode = modeStyle.Render("MODE: Install")
	}
	header := headerStyle.Render(" Python Package Manager Tool") + mode

	var content string
	if m.State == StateViewingList {
		content = m.List.View()
	} else {
		content = fmt.Sprintf(
			"\n Type the name of PyPI package to install:\n\n %s\n\n",
			m.SearchInput.View(),
		)
	}

	var status string
	if m.IsLoading {
		status = fmt.Sprintf("%s %s", m.Spinner.View(), statusStyle.Render(m.StatusMessage))
	} else if m.Err != nil {
		status = errorStyle.Render(fmt.Sprintf("Error: %v", m.Err))
	} else {
		status = statusStyle.Render(m.StatusMessage)
	}

	var help string
	if m.State == StateViewingList {
		help = helpStyle.Render("• [d] uninstall pacakge • [tab] switch to installer • [q] quit")
	} else {
		help = helpStyle.Render("• [enter] install package • [tab] swtich to browser • [q] quit")
	}

	return docStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s", header, content, status, help))
}

type packagesMsg struct {
	packages []client.LocalPackage
}

type statusMsg struct {
	message string
}

type errMsg struct {
	err error
}

func fetchPackagesCmd(c *client.PythonBridgeClient) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := c.GetInstalledPackages()
		if err != nil {
			return errMsg{err}
		}
		return packagesMsg{pkgs}
	}
}

func installPackageCmd(c *client.PythonBridgeClient, name string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.FetchPackageFromPyPI(name)
		if err != nil {
			return errMsg{fmt.Errorf("package '%s' not found on PyPI", name)}
		}

		msg, err := c.InstallPackages(name)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{msg}
	}
}

func deletePackageCmd(c *client.PythonBridgeClient, name string) tea.Cmd {
	return func() tea.Msg {
		msg, err := c.DeletePackage(name)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg{msg}
	}
}
