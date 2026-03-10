package commands

import (
	"github.com/avitacco/jig/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Execute() error {
	app := NewApp()

	rootCmd := &cobra.Command{
		Use:   "jig",
		Short: "A tool for building and publishing Puppet modules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			debug, _ := cmd.Flags().GetBool("debug")

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			app.Config = cfg

			if debug {
				app.Logger.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().String("config", "", "Path to config file")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")

	rootCmd.AddCommand(app.newCmd())

	return rootCmd.Execute()
}
