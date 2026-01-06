package cmd

import (
	"ayayushsharma/rocket/constants"
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "rocket",
	Short: "Launch for your webapps to your own hardware.\n\n" +
		"No longer you depend on the cloud apps which you are not sure of how they \n" +
		"track your data. Use open source versions of your beloved apps and host on\n" +
		"your own machine.",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var programLevel slog.LevelVar
	programLevel.Set(slog.LevelDebug)

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: &programLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $XDG_CONFIG_HOME/rocket/rocket.yaml)",
	)
}

func initializeConfig(cmd *cobra.Command) error {
	viper.SetEnvPrefix(strings.ToUpper(constants.ApplicationName))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		slog.Debug("Mac Os config settings")
		homeDir, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(
			homeDir + "/" +
				".config" + "/" +
				constants.ApplicationName + "/",
		)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// ignore "file not found" errors, but panic on any other error
	if err := viper.ReadInConfig(); err != nil {
		var ignorableError viper.ConfigFileNotFoundError
		if !errors.As(err, &ignorableError) {
			return err
		}
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	slog.Debug("Configuration initialized", "config", viper.ConfigFileUsed())
	return nil
}
