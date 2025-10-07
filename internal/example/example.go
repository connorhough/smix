package example

import (
	"fmt"
	"os"
)

// Run demonstrates business logic separation by taking input parameters
// from the CLI and performing some business logic (printing a message multiple times)
func Run(message string, count int) error {
	// Validate input
	if count < 0 {
		return fmt.Errorf("count must be non-negative")
	}

	// Business logic: print the message 'count' times
	for i := 0; i < count; i++ {
		fmt.Fprintln(os.Stdout, message)
	}

	return nil
}