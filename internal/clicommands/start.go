package clicommands

import (
	"fmt"
	"log"
	"path/filepath"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
  Use:   "start [port] [subdomain]",
  Short: "Start a new proxy tunnel",
  Args:  cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {
    port := args[0]
    subdomain := args[1]

    execPath, err := os.Executable()
    if err != nil {
      return fmt.Errorf("Failed to get executable path: %v", err)
    }

    scriptPath := filepath.Join( filepath.Dir(execPath), "start_tunnel")

    command := exec.Command("bash", scriptPath, port, subdomain)

    command.Stdout = os.Stdout
    command.Stderr = os.Stderr
    command.Stdin  = os.Stdin

    err = command.Run()

    if err != nil {
      log.Fatalf("Script execution failed: %v", err)
    }

    return nil
  },
  Example: "kengrok start 3000 myapp",
}

func init() {
  rootCmd.AddCommand(startCmd)
}
