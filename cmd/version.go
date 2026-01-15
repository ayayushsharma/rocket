package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"ayayushsharma/rocket/constants"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(constants.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
