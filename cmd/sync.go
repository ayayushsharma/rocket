package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"

	"ayayushsharma/rocket/workspace"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs application configs with user changes to the manifests",

	RunE: func(cmd *cobra.Command, args []string) (err error) {

		err = workspace.SyncRouter()
		if err != nil {
			slog.Debug("Failed to sync router configs")
			return
		}
		return nil

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
