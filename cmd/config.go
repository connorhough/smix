package cmd

import (
	"fmt"

	"github.com/connorhough/smix/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage smix configuration",
		Long:  `Get and set smix configuration values.`,
	}

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "get <key>",
			Short: "Get a configuration value",
			Long:  `Get a configuration value by key.`,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				value, err := config.GetValue(args[0])
				if err != nil {
					return err
				}
				fmt.Println(value)
				return nil
			},
		},
		&cobra.Command{
			Use:   "set <key> <value>",
			Short: "Set a configuration value",
			Long:  `Set a configuration value by key.`,
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return config.SetValue(args[0], args[1])
			},
		},
	)

	return configCmd
}