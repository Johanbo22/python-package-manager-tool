package main

import (
	"fmt"
	"os"

	"github.com/Johanbo22/python-package-manager-tool/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirupsen/logrus"
)

func main() {
	file, err := os.OpenFile("../logs/app_go.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.Info("Failed to log to file, using default stderr")
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if os.Getenv("MANAGER_API_KEY") == "" {
		fmt.Println("Error: MANAGER_API_KEY environment variable is not set")
		os.Exit(1)
	}

	program := tea.NewProgram(tui.InitialModel())
	if _, err := program.Run(); err != nil {
		logrus.Fatal("Error running app", err)
		os.Exit(1)
	}
}
