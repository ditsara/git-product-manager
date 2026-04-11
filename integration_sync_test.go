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
