package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "registers specified containerised application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		containers.Connect()
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

