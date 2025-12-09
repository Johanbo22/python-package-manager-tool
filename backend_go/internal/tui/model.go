package tui

import (
	"fmt"

	"github.com/Johanbo22/python-package-manager-tool/internal/client"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ApplicationState int

const (
	StateViewingList ApplicationState = iota
	StateSearchingPyPI
)

type MainModel struct {
	State             ApplicationState
	PythonClient      *client.PythonBridgeClient
	InstalledPackages []client.LocalPackage
	SearchInput       textinput.Model
	StatusMessage     string
	CursorIndex       int
	Err               error
}

func InitialModel() MainModel {
	ti := textinput.New()
	ti.Placeholder = "Type package name to install"
	ti.Focus()

	return MainModel{
		State:        StateViewingList,
		PythonClient: client.NewPythonBridgeClient(),
		SearchInput:  ti,
		CursorIndex:  0,
	}
}

func (m MainModel) Init() tea.Cmd {
	return fetchPackagesCmd(m.PythonClient)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.State == StateViewingList {
				m.State = StateSearchingPyPI
			} else {
				m.State = StateViewingList
			}
		case "up":
			if m.CursorIndex > 0 {
				m.CursorIndex--
			}
		case "down":
			if m.CursorIndex < len(m.InstalledPackages)-1 {
				m.CursorIndex++
			}
		case "d":
			if m.State == StateViewingList && len(m.InstalledPackages) > 0 {
				pkgToDelete := m.InstalledPackages[m.CursorIndex].Name
				m.StatusMessage = "Uninstalling " + pkgToDelete + "..."
				return m, deletePackageCmd(m.PythonClient, pkgToDelete)
			}
		case "enter":
			if m.State == StateSearchingPyPI {
				pkgToInstall := m.SearchInput.Value()
				m.StatusMessage = "Installing " + pkgToInstall + "..."
				return m, installPackageCmd(m.PythonClient, pkgToInstall)
			}
		}

	case packagesMsg:
		m.InstalledPackages = msg.packages

	case statusMsg:
		m.StatusMessage = msg.message
		return m, fetchPackagesCmd(m.PythonClient)

	case errMsg:
		m.Err = msg.err
		m.StatusMessage = "Error: " + msg.err.Error()
	}

	if m.State == StateSearchingPyPI {
		m.SearchInput, cmd = m.SearchInput.Update(msg)
	}

	return m, cmd
}

func (m MainModel) View() string {
	s := "Python Package Manager TUI\n\n"

	if m.State == StateViewingList {
		s += "Installed Packages (Press 'd' to delete, 'Tab' to install new packages):\n"
		for i, pkg := range m.InstalledPackages {
			cursor := " "
			if m.CursorIndex == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s (%s)\n", cursor, pkg.Name, pkg.Version)
		}
	} else {
		s += "Install new package (Press 'Enter' to install, 'Tab' to go back):\n"
		s += m.SearchInput.View() + "\n"
	}

	s += "\nStatus: " + m.StatusMessage + "\n"
	s += "\n[q] Quit | [Tab] Switch Mode"
	return s
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
			return errMsg{fmt.Errorf("packages not found on PyPI")}
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
