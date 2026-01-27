package cmd

import (
	"fmt"
	"time"

	"github.com/connorhough/smix/internal/resume"
	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	var at string
	var message string
	var filter string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Schedule a continue message",
		Long:  `Wait until a specific time and type a message into the active terminal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse time
			// Try multiple formats
			var targetTime time.Time
			now := resume.SystemClock.Now()

			// Support 12h and 24h formats, with/without seconds, with/without space for AM/PM
			formats := []string{
				"15:04", "15:04:05",
				"3:04PM", "3:04 PM",
				"3:04pm", "3:04 pm",
			}
			parsed := false

			for _, f := range formats {
				t, e := time.ParseInLocation(f, at, time.Local)
				if e == nil {
					// time.Parse returns year 0, set to today
					targetTime = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
					parsed = true
					break
				}
			}

			if !parsed {
				return fmt.Errorf("could not parse time '%s'. Supported formats: HH:MM, HH:MM:SS, HH:MM PM", at)
			}

			return resume.Run(cmd.Context(), targetTime, message, filter)
		},
	}

	cmd.Flags().StringVar(&at, "at", "", "Time to resume (e.g., '14:30')")
	_ = cmd.MarkFlagRequired("at")
	cmd.Flags().StringVar(&message, "message", "continue", "Message to type")
	cmd.Flags().StringVar(&filter, "filter", "", "Substring required in active window title for safety")

	return cmd
}
