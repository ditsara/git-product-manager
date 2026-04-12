package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationSyncAcrossBranches(t *testing.T) {
	pmBinary := buildPMBinary(t)
	root := t.TempDir()

	remote := filepath.Join(root, "remote.git")
	mainRepo := filepath.Join(root, "main-repo")
	featureOne := filepath.Join(root, "feature-one")
	featureTwo := filepath.Join(root, "feature-two")
	featureThree := filepath.Join(root, "feature-three")
	verifyRepo := filepath.Join(root, "verify-repo")

	runGit(t, root, "init", "--bare", remote)
	runGit(t, root, "clone", remote, mainRepo)
	runGit(t, mainRepo, "checkout", "-b", "main")
	configureGitIdentity(t, mainRepo)

	initWorkspace(t, pmBinary, mainRepo, "SYNC")
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "Initial pm setup")
	runGit(t, mainRepo, "push", "-u", "origin", "main")
	runGit(t, root, "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/main")

	runGit(t, root, "clone", remote, featureOne)
	runGit(t, featureOne, "checkout", "-b", "feature/one", "origin/main")
	configureGitIdentity(t, featureOne)
	ensurePMDirs(t, featureOne)

	runGit(t, root, "clone", remote, featureTwo)
	runGit(t, featureTwo, "checkout", "-b", "feature/two", "origin/main")
	configureGitIdentity(t, featureTwo)
	ensurePMDirs(t, featureTwo)

	runGit(t, root, "clone", remote, featureThree)
	runGit(t, featureThree, "checkout", "-b", "feature/three", "origin/main")
	configureGitIdentity(t, featureThree)
	ensurePMDirs(t, featureThree)

	if output, err := runPM(t, pmBinary, featureOne, "new", "Shared ticket from feature one"); err != nil {
		t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
	}

	pushOutput, err := runPM(t, pmBinary, featureOne, "sync", "push")
	if err != nil {
		t.Fatalf("pm sync push failed: %v\nOutput: %s", err, pushOutput)
	}
	if !strings.Contains(pushOutput, "Pushed") {
		t.Fatalf("pm sync push did not report success:\n%s", pushOutput)
	}

	pullOutput, err := runPM(t, pmBinary, featureThree, "sync", "pull")
	if err != nil {
		t.Fatalf("pm sync pull failed: %v\nOutput: %s", err, pullOutput)
	}
	if !strings.Contains(pullOutput, "Pulled") {
		t.Fatalf("pm sync pull did not report success:\n%s", pullOutput)
	}

	sharedTicket := filepath.Join(featureThree, ".pm", "tickets", "SYNC-1.md")
	if _, err := os.Stat(sharedTicket); err != nil {
		t.Fatalf("expected pulled ticket %s: %v", sharedTicket, err)
	}

	combinedOutput, err := runPM(t, pmBinary, featureTwo, "sync")
	if err != nil {
		t.Fatalf("pm sync failed: %v\nOutput: %s", err, combinedOutput)
	}
	if !strings.Contains(combinedOutput, "Pulled") || !strings.Contains(combinedOutput, "No .pm changes to push") {
		t.Fatalf("pm sync combined did not run clean pull/push sequence:\n%s", combinedOutput)
	}

	sharedTicket = filepath.Join(featureTwo, ".pm", "tickets", "SYNC-1.md")
	if _, err := os.Stat(sharedTicket); err != nil {
		t.Fatalf("expected combined sync to pull ticket %s: %v", sharedTicket, err)
	}

	if output, err := runPM(t, pmBinary, featureTwo, "new", "Second ticket from feature two"); err != nil {
		t.Fatalf("pm new failed: %v\nOutput: %s", err, output)
	}

	pushOutput, err = runPM(t, pmBinary, featureTwo, "sync", "push")
	if err != nil {
		t.Fatalf("pm sync push failed: %v\nOutput: %s", err, pushOutput)
	}
	if !strings.Contains(pushOutput, "Pushed") {
		t.Fatalf("pm sync push did not push successfully:\n%s", pushOutput)
	}

	runGit(t, root, "clone", remote, verifyRepo)
	runGit(t, verifyRepo, "checkout", "main")

	for _, ticketID := range []string{"SYNC-1.md", "SYNC-2.md"} {
		path := filepath.Join(verifyRepo, ".pm", "tickets", ticketID)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected synced ticket %s in main branch: %v", ticketID, err)
		}
	}
}

func TestSyncPullRejectsUncommittedPMChanges(t *testing.T) {
	pmBinary := buildPMBinary(t)
	root := t.TempDir()

	remote := filepath.Join(root, "remote.git")
	mainRepo := filepath.Join(root, "main-repo")
	featureRepo := filepath.Join(root, "feature-repo")

	runGit(t, root, "init", "--bare", remote)
	runGit(t, root, "clone", remote, mainRepo)
	runGit(t, mainRepo, "checkout", "-b", "main")
	configureGitIdentity(t, mainRepo)

	initWorkspace(t, pmBinary, mainRepo, "SYNC")
	runGit(t, mainRepo, "add", ".")
	runGit(t, mainRepo, "commit", "-m", "Initial pm setup")
	runGit(t, mainRepo, "push", "-u", "origin", "main")
	runGit(t, root, "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/main")

	runGit(t, root, "clone", remote, featureRepo)
	runGit(t, featureRepo, "checkout", "-b", "feature/dirty", "origin/main")
	configureGitIdentity(t, featureRepo)
	ensurePMDirs(t, featureRepo)

	dirtyPath := filepath.Join(featureRepo, ".pm", "tickets", "DIRTY.md")
	if err := os.WriteFile(dirtyPath, []byte("dirty"), 0644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	output, err := runPM(t, pmBinary, featureRepo, "sync", "pull")
	if err == nil {
		t.Fatalf("expected pm sync pull to fail with uncommitted .pm changes\nOutput: %s", output)
	}
	if !strings.Contains(output, "uncommitted changes in .pm") {
		t.Fatalf("expected uncommitted-changes error, got:\n%s", output)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func configureGitIdentity(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
}

func ensurePMDirs(t *testing.T, repo string) {
	t.Helper()
	for _, dir := range []string{
		filepath.Join(repo, ".pm", "tickets"),
		filepath.Join(repo, ".pm", "milestones"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
}

// runPMStderr runs the pm binary and returns stdout+stderr separately.
func runPMStderr(t *testing.T, pmBinary, workDir string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := exec.Command(pmBinary, args...)
	cmd.Dir = workDir
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func TestSymmetryWarningOnSync(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "SYM")

	ticketsDir := filepath.Join(workspace, ".pm", "tickets")

	// A depends_on B, but B does not block A — broken symmetry.
	ticketA := `---
id: SYM-1
title: Ticket A
type: task
status: todo
priority: medium
depends_on: [SYM-2]
blocks: []
related: []
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-01T00:00:00Z"
---
`
	ticketB := `---
id: SYM-2
title: Ticket B
type: task
status: todo
priority: medium
depends_on: []
blocks: []
related: []
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-01T00:00:00Z"
---
`
	if err := os.WriteFile(filepath.Join(ticketsDir, "SYM-1.md"), []byte(ticketA), 0644); err != nil {
		t.Fatalf("write SYM-1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketsDir, "SYM-2.md"), []byte(ticketB), 0644); err != nil {
		t.Fatalf("write SYM-2: %v", err)
	}

	// pm list triggers a cache sync which should print the symmetry warning to stderr.
	_, stderr, _ := runPMStderr(t, pmBinary, workspace, "list")

	if !strings.Contains(stderr, "SYM-1 depends_on SYM-2") {
		t.Errorf("expected symmetry warning on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "pm link SYM-2 SYM-1 --type blocks") {
		t.Errorf("expected pm link fix command on stderr, got:\n%s", stderr)
	}
}

func TestSymmetryWarningRelated(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "REL")

	ticketsDir := filepath.Join(workspace, ".pm", "tickets")

	// A related B, but B does not relate back to A.
	ticketA := `---
id: REL-1
title: Ticket A
type: task
status: todo
priority: medium
depends_on: []
blocks: []
related: [REL-2]
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-01T00:00:00Z"
---
`
	ticketB := `---
id: REL-2
title: Ticket B
type: task
status: todo
priority: medium
depends_on: []
blocks: []
related: []
created_at: "2026-01-01T00:00:00Z"
updated_at: "2026-01-01T00:00:00Z"
---
`
	if err := os.WriteFile(filepath.Join(ticketsDir, "REL-1.md"), []byte(ticketA), 0644); err != nil {
		t.Fatalf("write REL-1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketsDir, "REL-2.md"), []byte(ticketB), 0644); err != nil {
		t.Fatalf("write REL-2: %v", err)
	}

	_, stderr, _ := runPMStderr(t, pmBinary, workspace, "list")

	if !strings.Contains(stderr, "REL-1 related REL-2") {
		t.Errorf("expected related symmetry warning on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "pm link REL-2 REL-1 --type related") {
		t.Errorf("expected pm link fix command on stderr, got:\n%s", stderr)
	}
}

func TestNoSymmetryWarningWhenClean(t *testing.T) {
	pmBinary := buildPMBinary(t)
	workspace := t.TempDir()
	initGitRepo(t, workspace)
	initWorkspace(t, pmBinary, workspace, "CLN")

	// Create tickets via pm new + pm link so symmetry is correct from the start.
	if _, err := runPM(t, pmBinary, workspace, "new", "Ticket A"); err != nil {
		t.Fatalf("pm new A: %v", err)
	}
	if _, err := runPM(t, pmBinary, workspace, "new", "Ticket B"); err != nil {
		t.Fatalf("pm new B: %v", err)
	}
	if _, err := runPM(t, pmBinary, workspace, "link", "CLN-1", "CLN-2", "--type", "depends-on"); err != nil {
		t.Fatalf("pm link: %v", err)
	}

	_, stderr, _ := runPMStderr(t, pmBinary, workspace, "list")
	if strings.Contains(stderr, "⚠") {
		t.Errorf("expected no warnings for clean data, got stderr:\n%s", stderr)
	}
}
