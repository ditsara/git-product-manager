package main

import "testing"

func TestParseGitLogLine(t *testing.T) {
	line := "abc123|Alice Example|1707372000|Move to todo"
	info, ok := parseGitLogLine(line)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if info.hash != "abc123" {
		t.Errorf("unexpected hash: %s", info.hash)
	}
	if info.author != "Alice Example" {
		t.Errorf("unexpected author: %s", info.author)
	}
	if info.message != "Move to todo" {
		t.Errorf("unexpected message: %s", info.message)
	}
}

func TestParseStatusDiff(t *testing.T) {
	diff := `diff --git a/.pm/tickets/TEST-1.md b/.pm/tickets/TEST-1.md
index 123..456 100644
--- a/.pm/tickets/TEST-1.md
+++ b/.pm/tickets/TEST-1.md
@@
-status: backlog
+status: "in-progress"
`
	oldStatus, newStatus, ok := parseStatusDiff(diff)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if oldStatus != "backlog" {
		t.Errorf("unexpected old status: %s", oldStatus)
	}
	if newStatus != "in-progress" {
		t.Errorf("unexpected new status: %s", newStatus)
	}
}

func TestParseStatusDiffNoMatch(t *testing.T) {
	diff := `diff --git a/.pm/tickets/TEST-1.md b/.pm/tickets/TEST-1.md
index 123..456 100644
--- a/.pm/tickets/TEST-1.md
+++ b/.pm/tickets/TEST-1.md
@@
-title: "Old"
+title: "New"
`
	_, _, ok := parseStatusDiff(diff)
	if ok {
		t.Fatalf("expected no match")
	}
}
