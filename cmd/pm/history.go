package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ditsara/git-product-manager/internal/ticket"
	"github.com/spf13/cobra"
)

type commitInfo struct {
	hash    string
	author  string
	time    time.Time
	message string
}

type historyEntry struct {
	time       time.Time
	author     string
	fromStatus string
	toStatus   string
	message    string
	isCreation bool
}

var historyCmd = &cobra.Command{
	Use:               "history <id>",
	Short:             "Show ticket status change history",
	Long:              "Show an audit trail of status changes for a ticket by parsing git history.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTicketIDs,
	Example:           "  pm history GPM-123",
	RunE:              runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	inputID := args[0]
	ticketPath := findTicketByID(inputID)
	actualTicketID := strings.TrimSuffix(filepath.Base(ticketPath), ".md")

	entries, err := buildHistory(ticketPath)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return fmt.Errorf("no status history found for %s", actualTicketID)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].time.Equal(entries[j].time) {
			if entries[i].isCreation != entries[j].isCreation {
				return entries[i].isCreation
			}
			return entries[i].message < entries[j].message
		}
		return entries[i].time.Before(entries[j].time)
	})

	fmt.Printf("State Change History for %s:\n\n", actualTicketID)
	for _, entry := range entries {
		timestamp := entry.time.UTC().Format("2006-01-02 15:04")
		if entry.isCreation {
			fmt.Printf("%s  %s  Created (status: %s)\n", timestamp, entry.author, entry.toStatus)
		} else {
			fmt.Printf("%s  %s  %s → %s\n", timestamp, entry.author, entry.fromStatus, entry.toStatus)
		}

		if strings.TrimSpace(entry.message) != "" {
			fmt.Printf("    commit: %s\n", entry.message)
		}
	}

	return nil
}

func buildHistory(ticketPath string) ([]historyEntry, error) {
	logLines, err := getGitLog(ticketPath)
	if err != nil {
		return nil, err
	}
	if len(logLines) == 0 {
		return nil, fmt.Errorf("ticket not in git history")
	}

	entries := []historyEntry{}

	for _, line := range logLines {
		info, ok := parseGitLogLine(line)
		if !ok {
			continue
		}

		diff, err := getCommitDiff(info.hash, ticketPath)
		if err != nil {
			continue
		}

		oldStatus, newStatus, ok := parseStatusDiff(diff)
		if !ok {
			continue
		}

		entries = append(entries, historyEntry{
			time:       info.time,
			author:     info.author,
			fromStatus: oldStatus,
			toStatus:   newStatus,
			message:    info.message,
			isCreation: false,
		})
	}

	creation, err := getCreationEntry(ticketPath)
	if err != nil {
		return nil, err
	}
	entries = append(entries, creation)

	return entries, nil
}

func getGitLog(ticketPath string) ([]string, error) {
	cmd := exec.Command("git", "log", "--follow", "--format=%H|%an|%at|%s", "--", ticketPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, gitError(output, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []string{}, nil
	}
	return lines, nil
}

func parseGitLogLine(line string) (commitInfo, bool) {
	parts := strings.SplitN(line, "|", 4)
	if len(parts) != 4 {
		return commitInfo{}, false
	}

	unixTime, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
	if err != nil {
		return commitInfo{}, false
	}

	return commitInfo{
		hash:    strings.TrimSpace(parts[0]),
		author:  strings.TrimSpace(parts[1]),
		time:    time.Unix(unixTime, 0).UTC(),
		message: strings.TrimSpace(parts[3]),
	}, true
}

func getCommitDiff(hash string, ticketPath string) (string, error) {
	cmd := exec.Command("git", "show", hash, "--", ticketPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", gitError(output, err)
	}
	return string(output), nil
}

func parseStatusDiff(diff string) (string, string, bool) {
	var oldStatus string
	var newStatus string

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "-status:") || strings.HasPrefix(line, "- status:") {
			oldStatus = parseStatusValue(line)
		}
		if strings.HasPrefix(line, "+status:") || strings.HasPrefix(line, "+ status:") {
			newStatus = parseStatusValue(line)
		}
	}

	if oldStatus == "" || newStatus == "" {
		return "", "", false
	}

	return oldStatus, newStatus, true
}

func parseStatusValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	value := strings.TrimSpace(parts[1])
	if idx := strings.Index(value, " #"); idx != -1 {
		value = value[:idx]
	}
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")
	return value
}

func getCreationEntry(ticketPath string) (historyEntry, error) {
	cmd := exec.Command("git", "log", "--follow", "--diff-filter=A", "--format=%H|%an|%at|%s", "--", ticketPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return historyEntry{}, gitError(output, err)
	}
	line := strings.TrimSpace(string(output))
	if line == "" {
		return historyEntry{}, fmt.Errorf("ticket not in git history")
	}
	info, ok := parseGitLogLine(strings.SplitN(line, "\n", 2)[0])
	if !ok {
		return historyEntry{}, fmt.Errorf("failed to parse creation commit")
	}

	fileContent, err := getFileAtCommit(info.hash, ticketPath)
	if err != nil {
		return historyEntry{}, err
	}

	t, err := ticket.Parse([]byte(fileContent))
	if err != nil {
		return historyEntry{}, fmt.Errorf("failed to parse ticket from creation commit: %w", err)
	}

	return historyEntry{
		time:       info.time,
		author:     info.author,
		fromStatus: "",
		toStatus:   t.Status,
		message:    info.message,
		isCreation: true,
	}, nil
}

func getFileAtCommit(hash string, ticketPath string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", hash, ticketPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", gitError(output, err)
	}
	return string(output), nil
}

func gitError(output []byte, err error) error {
	msg := strings.TrimSpace(string(output))
	if msg == "" {
		return err
	}
	if strings.Contains(msg, "not a git repository") || strings.Contains(msg, "fatal: not a git repository") {
		return errors.New("git not available or not a git repository")
	}
	if strings.Contains(msg, "does not have any commits yet") {
		return errors.New("ticket not in git history")
	}
	if strings.Contains(msg, "not found") || strings.Contains(msg, "pathspec") {
		return errors.New("ticket not in git history")
	}
	return fmt.Errorf("git error: %s", msg)
}
