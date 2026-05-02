package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var showFlags struct {
	noComments bool
	noPager    bool
}

var showCmd = &cobra.Command{
	Use:               "show <id>",
	Short:             "Show a single ticket",
	Long:              `Shows the details of a single ticket.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTicketIDs,
	Example:           "  pm show PROJ-123",
	Run: func(cmd *cobra.Command, args []string) {
		ticketID := args[0]
		ticketPath := getTicketPath(ticketID)

		content, err := os.ReadFile(ticketPath)
		if err != nil {
			log.Fatalf("Error reading ticket file: %v", err)
		}

		parts := strings.SplitN(string(content), "---", 3)
		if len(parts) < 3 {
			fmt.Println("Error: Invalid ticket format. Missing YAML front matter.")
			os.Exit(1)
		}

		var buf bytes.Buffer
		fmt.Fprintln(&buf, "---")
		fmt.Fprintln(&buf, strings.TrimSpace(parts[1]))
		fmt.Fprintln(&buf, "---")
		fmt.Fprintln(&buf, strings.TrimSpace(parts[2]))

		// Extract actual ticket ID from path for comment directory
		actualTicketID := strings.TrimSuffix(filepath.Base(ticketPath), ".md")

		// Collect comments if not suppressed
		if !showFlags.noComments {
			collectComments(actualTicketID, &buf)
		}

		displayOutput(&buf)
	},
}

func init() {
	showCmd.Flags().BoolVar(&showFlags.noComments, "no-comments", false, "Hide comments")
	showCmd.Flags().BoolVar(&showFlags.noPager, "no-pager", false, "Write output directly to stdout, skipping bat/less")
	rootCmd.AddCommand(showCmd)
}

// displayOutput writes buf to a pager (bat → less) when stdout is a TTY,
// or directly to stdout otherwise.
func displayOutput(buf *bytes.Buffer) {
	usePager := !showFlags.noPager && term.IsTerminal(int(os.Stdout.Fd()))

	if usePager {
		if pagerPath, err := exec.LookPath("bat"); err == nil {
			runPager(pagerPath, []string{"--language=md", "--paging=always"}, buf)
			return
		}
		if pagerPath, err := exec.LookPath("less"); err == nil {
			runPager(pagerPath, []string{"-R"}, buf)
			return
		}
	}

	os.Stdout.Write(buf.Bytes())
}

// runPager pipes buf through the given pager command.
func runPager(path string, args []string, buf *bytes.Buffer) {
	pager := exec.Command(path, args...)
	pager.Stdin = buf
	pager.Stdout = os.Stdout
	pager.Stderr = os.Stderr
	if err := pager.Run(); err != nil {
		// Fall back to plain stdout if pager fails
		os.Stdout.Write(buf.Bytes())
	}
}

// collectComments writes formatted comments for ticketID into buf.
func collectComments(ticketID string, buf *bytes.Buffer) {
	comments, err := ticket.ListCommentsForTicket(ticketID, ".pm")
	if err != nil || len(comments) == 0 {
		return
	}

	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Fprintf(buf, "Comments (%d):\n", len(comments))
	fmt.Fprintln(buf)

	for i, comment := range comments {
		createdAt := comment.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC")
		if comment.UpdatedAt.After(comment.CreatedAt) {
			updatedAt := comment.UpdatedAt.UTC().Format("2006-01-02 15:04:05 UTC")
			fmt.Fprintf(buf, "@%s (%s, edited %s)\n", comment.Author, createdAt, updatedAt)
		} else {
			fmt.Fprintf(buf, "@%s (%s)\n", comment.Author, createdAt)
		}
		fmt.Fprintln(buf, comment.Body)

		if i < len(comments)-1 {
			fmt.Fprintln(buf)
		}
	}
}

// displayComments is kept for backward compatibility with existing tests.
func displayComments(ticketID string) {
	var buf bytes.Buffer
	collectComments(ticketID, &buf)
	os.Stdout.Write(buf.Bytes())
}
