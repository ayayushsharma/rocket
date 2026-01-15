//go:build !production
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checking health of application contaienrs !TODO",

	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
