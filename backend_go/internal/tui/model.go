package tui

import (
	"fmt"

	"github.com/Johanbo22/python-package-manager-tool/internal/client"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ApplicationState int

const (
	StateViewingList ApplicationState = iota
	StateSearchingPyPi
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
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
	l.SetShowHelp(true)

	return MainModel{
		State:        StateViewingList,
		PythonClient: client.NewPythonBridgeClient(),
		SearchInput:  ti,
		List:         l,
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
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.State == StateViewingList {
				m.State = StateSearchingPyPi
				m.StatusMessage = "Mode: Install New Package"
			} else {
				m.State = StateViewingList
				m.StatusMessage = "Mode: Browse Packages"
			}
			return m, nil
		}
		if m.State == StateViewingList {
			switch msg.String() {
			case "d":
				if selectedItem, ok := m.List.SelectedItem().(item); ok {
					m.StatusMessage = "Uninstalling " + selectedItem.name + "..."
					return m, deletePackageCmd(m.PythonClient, selectedItem.name)
				}
			}
		} else if m.State == StateSearchingPyPi {
			switch msg.String() {
			case "enter":
				pkgToInstall := m.SearchInput.Value()
				if pkgToInstall != "" {
					m.StatusMessage = "Installing " + pkgToInstall + "..."
					m.SearchInput.SetValue("")
					return m, installPackageCmd(m.PythonClient, pkgToInstall)
				}
			}
		}
	case packagesMsg:
		items := make([]list.Item, len(msg.packages))
		for i, pkg := range msg.packages {
			items[i] = item{name: pkg.Name, version: pkg.Version}
		}
		cmd = m.List.SetItems(items)
		cmds = append(cmds, cmd)
		m.StatusMessage = fmt.Sprintf("Updated: %d packages found", len(msg.packages))

	case statusMsg:
		m.StatusMessage = msg.message
		cmds = append(cmds, fetchPackagesCmd(m.PythonClient))

	case errMsg:
		m.Err = msg.err
		m.StatusMessage = "Error: " + msg.err.Error()
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
	if m.State == StateViewingList {
		return docStyle.Render(m.List.View())
	}
	return docStyle.Render(fmt.Sprintf(
		"Install New Package\n\n%s\n\n%s\n\n[Tab] Back to List | [Enter] Install | [Ctrl+C] Quit",
		m.SearchInput.View(),
		m.StatusMessage,
	))
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
