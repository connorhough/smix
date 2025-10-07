package cmd

import (
	"github.com/connorhough/smix/internal/example"
	"github.com/spf13/cobra"
)

func newExampleCmd() *cobra.Command {
	var (
		message string
		count   int
	)

	exampleCmd := &cobra.Command{
		Use:   "example",
		Short: "Example command demonstrating the proper pattern",
		Long: `This command serves as a template for adding new commands.
It shows how CLI commands should delegate to internal packages for business logic.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Delegate to internal package
			return example.Run(message, count)
		},
	}

	// Define flags
	exampleCmd.Flags().StringVar(&message, "message", "Hello from smix!", "Message to print")
	exampleCmd.Flags().IntVar(&count, "count", 1, "Number of times to print the message")

	return exampleCmd
}