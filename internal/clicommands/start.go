package clicommands

import (
	"fmt"
	"log"
	"strconv"

	"github.com/kennyfellows/go-kengrok/internal/reversetunneler"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
  Use:   "start [port] [subdomain]",
  Short: "Start a new proxy tunnel",
  Args:  cobra.ExactArgs(2),
  RunE: func(cmd *cobra.Command, args []string) error {

    port, err := strconv.Atoi(args[0])

    if err != nil {
      return fmt.Errorf("invalid port number: %s", args[0])
    }

    subdomain := args[1]

    if port < 1 || port > 65535 {
      return fmt.Errorf("port must be between 1 and 65535")
    }

    fmt.Printf("Starting proxy: port=%d, subdomain=%s\n", port, subdomain)

    err = reversetunneler.StartTunnel( port, subdomain )

    if err != nil {
      log.Fatalf("Error starting tunnel %v", err)
    }

    return nil
  },
  Example: "kengrok start 3000 myapp",
}

func init() {
  rootCmd.AddCommand(startCmd)
}
