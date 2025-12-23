package cmd

import (
	"context"
	"fmt"
	"strings"

	"os"

	"github.com/connorhough/smix/internal/gca"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func newPRCmd() *cobra.Command {
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Work with Pull Requests",
		Long:  `Commands to work with Pull Requests on GitHub.`,
	}

	prCmd.AddCommand(newPRReviewCmd())

	return prCmd
}

func newPRReviewCmd() *cobra.Command {
	var useExistingDir string

	cmd := &cobra.Command{
		Use:   "review <repo> <pr_number>",
		Short: "Fetch and process gemini-code-assist feedback from a GitHub PR",
		Long: `Fetch gemini-code-assist feedback from a GitHub PR and launch Claude Code sessions to analyze and implement the suggested changes.
The repo argument should be in the format "owner/name" (e.g. "octocat/Hello-World").
The pr_number argument should be the PR number (e.g. 123).

To process an existing gca_review folder without fetching, use the --dir flag.`,
		Args: func(cmd *cobra.Command, args []string) error {
			// If --dir is set, allow 0 args, otherwise require 2
			if useExistingDir != "" {
				return cobra.NoArgs(cmd, args)
			}
			return cobra.ExactArgs(2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var outputDir string

			// If using existing directory, skip fetching
			if useExistingDir != "" {
				outputDir = useExistingDir
				fmt.Printf("Using existing directory: %s\n", outputDir)
			} else {
				// Parse repo owner and name
				parts := strings.Split(args[0], "/")
				if len(parts) != 2 {
					return fmt.Errorf("invalid repo format. Expected 'owner/name', got '%s'", args[0])
				}
				repoOwner := parts[0]
				repoName := parts[1]

				// Parse PR number
				var prNumber int
				if _, err := fmt.Sscanf(args[1], "%d", &prNumber); err != nil {
					return fmt.Errorf("invalid PR number: %w", err)
				}

				// Create GitHub client
				ctx := context.Background()
				var client *github.Client
				if token := os.Getenv("GITHUB_TOKEN"); token != "" {
					ts := oauth2.StaticTokenSource(
						&oauth2.Token{AccessToken: token},
					)
					tc := oauth2.NewClient(ctx, ts)
					client = github.NewClient(tc)
				} else {
					client = github.NewClient(nil)
				}

				// Create output directory
				outputDir = fmt.Sprintf("./gca_review_pr%d", prNumber)

				// Fetch reviews
				if err := gca.FetchReviews(ctx, client, repoOwner, repoName, prNumber, outputDir); err != nil {
					return fmt.Errorf("failed to fetch reviews: %w", err)
				}
			}

			// Process reviews
			if err := gca.ProcessReviews(outputDir); err != nil {
				return fmt.Errorf("failed to process reviews: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&useExistingDir, "dir", "", "Use existing gca_review directory instead of fetching from GitHub")

	return cmd
}
