package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume rockets from where they left off",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		slog.Debug("Resuming... " + constants.ApplicationName)
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Could not connect to podman", "error", err)
			return
		}
		err = startRouter(conn)
		if err != nil {
			slog.Debug("Failed to start application router", "error", err)
			return
		}

		err = launchAll(conn)
		if err != nil {
			slog.Debug("Failed to start application router", "error", err)
			return
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
