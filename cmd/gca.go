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

func newGCACmd() *cobra.Command {
	gcaCmd := &cobra.Command{
		Use:   "gca",
		Short: "Work with gemini-code-assist feedback",
		Long: `Commands to fetch and process feedback from gemini-code-assist bot on GitHub PRs.

GitHub authentication:
  The gca command uses the GitHub API to fetch PR feedback. It will automatically
  use a GITHUB_TOKEN environment variable if set. Without a token, it uses
  anonymous access which has stricter rate limits. For higher rate limits and
  access to private repositories, set a GitHub personal access token in the
  GITHUB_TOKEN environment variable.`,
	}

	gcaCmd.AddCommand(newGCAReviewCmd())

	return gcaCmd
}

func newGCAReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <repo> <pr_number>",
		Short: "Fetch and process gemini-code-assist feedback from a GitHub PR",
		Long: `Fetch gemini-code-assist feedback from a GitHub PR and process it with an LLM to generate code patches.
The repo argument should be in the format "owner/name" (e.g. "octocat/Hello-World").
The pr_number argument should be the PR number (e.g. 123).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			outputDir := fmt.Sprintf("./gca_review_pr%d", prNumber)

			// Fetch reviews
			if err := gca.FetchReviews(ctx, client, repoOwner, repoName, prNumber, outputDir); err != nil {
				return fmt.Errorf("failed to fetch reviews: %w", err)
			}

			// Process reviews
			if err := gca.ProcessReviews(ctx, outputDir); err != nil {
				return fmt.Errorf("failed to process reviews: %w", err)
			}

			return nil
		},
	}

	return cmd
}
