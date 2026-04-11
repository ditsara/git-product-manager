package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ditsara/git-product-manager/internal/cache"
	"github.com/ditsara/git-product-manager/internal/config"
	"github.com/spf13/cobra"
)

type syncRuntime struct {
	project    *config.Project
	repoRoot   string
	gitDir     string
	pmPath     string
	currentRef string
	syncBranch string
	upstream   string
	remote     string
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize .pm data across branches",
	Long: `Synchronize the .pm directory with the configured sync branch.

Examples:
  pm sync        # Pull, then push
  pm sync pull   # Pull .pm changes from the sync branch
  pm sync push   # Push current .pm changes to the sync branch`,
	Args: cobra.NoArgs,
	RunE: runSyncCombined,
}

var syncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .pm changes from the sync branch",
	Args:  cobra.NoArgs,
	RunE:  runSyncPull,
}

var syncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push current .pm changes to the sync branch",
	Args:  cobra.NoArgs,
	RunE:  runSyncPush,
}

func init() {
	syncCmd.AddCommand(syncPullCmd, syncPushCmd)
	rootCmd.AddCommand(syncCmd)
}

func runSyncCombined(cmd *cobra.Command, args []string) error {
	backupDir, hasBackup, err := backupPMDir(".pm")
	if err != nil {
		return err
	}
	if hasBackup {
		defer os.RemoveAll(backupDir)
	}

	if err := runSyncPull(cmd, args); err != nil {
		return err
	}
	if err := runSyncPush(cmd, args); err != nil {
		if hasBackup {
			restoreErr := restorePMDir(".pm", backupDir)
			if restoreErr != nil {
				return fmt.Errorf("%w (rollback failed: %v)", err, restoreErr)
			}
			if ensureErr := cache.EnsureCacheReady(".pm"); ensureErr != nil {
				return fmt.Errorf("%w (rollback cache bootstrap failed: %v)", err, ensureErr)
			}
			if cacheErr := cache.SyncCache(".pm"); cacheErr != nil {
				return fmt.Errorf("%w (rollback cache sync failed: %v)", err, cacheErr)
			}
		}
		return err
	}
	return nil
}

func runSyncPull(cmd *cobra.Command, args []string) error {
	rt, err := prepareSyncRuntime()
	if err != nil {
		return err
	}

	if rt.currentRef == rt.syncBranch {
		fmt.Printf("Already on sync branch %q\n", rt.syncBranch)
		return nil
	}

	if err := ensureNoUncommittedPMChanges(); err != nil {
		return err
	}
	if err := fetchSyncBranch(rt); err != nil {
		return err
	}

	base, err := gitOutput("", "merge-base", "HEAD", rt.upstream)
	if err != nil {
		return fmt.Errorf("failed to compute merge base with %s: %w", rt.syncBranch, err)
	}

	syncChanged, err := diffPaths(base, rt.upstream)
	if err != nil {
		return err
	}
	if len(syncChanged) == 0 {
		fmt.Printf("No .pm changes to pull from %s\n", rt.syncBranch)
		return nil
	}

	currentChanged, err := diffPaths(base, "HEAD")
	if err != nil {
		return err
	}

	conflicts := intersectPaths(syncChanged, currentChanged)
	switch normalizeConflictStrategy(rt.project.Sync.ConflictStrategy) {
	case "theirs":
		// Apply all sync branch changes, including conflicts.
	case "ours":
		syncChanged = subtractPaths(syncChanged, conflicts)
	default:
		if len(conflicts) > 0 {
			return newSyncConflictError("pull", rt.syncBranch, conflicts)
		}
	}

	if err := applyRefChanges(rt.upstream, syncChanged); err != nil {
		return err
	}
	if err := cache.SyncCache(rt.pmPath); err != nil {
		return fmt.Errorf("failed to update cache after pull: %w", err)
	}

	fmt.Printf("✓ Pulled .pm changes from %s\n", rt.syncBranch)
	return nil
}

func runSyncPush(cmd *cobra.Command, args []string) error {
	rt, err := prepareSyncRuntime()
	if err != nil {
		return err
	}

	if rt.currentRef == rt.syncBranch {
		fmt.Printf("Already on sync branch %q\n", rt.syncBranch)
		return nil
	}

	if err := fetchSyncBranch(rt); err != nil {
		return err
	}

	base, err := gitOutput("", "merge-base", "HEAD", rt.upstream)
	if err != nil {
		return fmt.Errorf("failed to compute merge base with %s: %w", rt.syncBranch, err)
	}

	currentChanged, err := currentWorkspacePathsChanged(rt.upstream)
	if err != nil {
		return err
	}
	if len(currentChanged) == 0 {
		fmt.Printf("No .pm changes to push to %s\n", rt.syncBranch)
		return nil
	}

	syncChanged, err := diffPaths(base, rt.upstream)
	if err != nil {
		return err
	}
	if conflicts := intersectPaths(currentChanged, syncChanged); len(conflicts) > 0 {
		return newSyncConflictError("push", rt.syncBranch, conflicts)
	}

	if err := pushPMViaWorktree(rt); err != nil {
		return err
	}

	fmt.Printf("✓ Pushed .pm changes to %s\n", rt.syncBranch)
	return nil
}

func prepareSyncRuntime() (*syncRuntime, error) {
	repoRoot, err := gitOutput("", "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not in a git repository. pm sync requires git for version control.\nInitialize git with: git init")
	}

	gitDir, err := gitOutput("", "rev-parse", "--git-dir")
	if err != nil {
		return nil, fmt.Errorf("failed to locate git directory: %w", err)
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(repoRoot, gitDir)
	}

	if err := ensureNoInProgressGitOperation(gitDir); err != nil {
		return nil, err
	}

	currentRef, err := gitOutput("", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("cannot sync from detached HEAD. Checkout a branch first")
	}

	project, err := config.LoadProject(".pm")
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}
	if err := cache.EnsureCacheReady(".pm"); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	syncBranch, err := resolveSyncBranch(project)
	if err != nil {
		return nil, err
	}

	upstream, remote, err := resolveRemoteTrackingRef(syncBranch)
	if err != nil {
		return nil, err
	}

	return &syncRuntime{
		project:    project,
		repoRoot:   repoRoot,
		gitDir:     gitDir,
		pmPath:     ".pm",
		currentRef: currentRef,
		syncBranch: syncBranch,
		upstream:   upstream,
		remote:     remote,
	}, nil
}

func resolveSyncBranch(project *config.Project) (string, error) {
	if project.Sync.Branch != "" {
		if !branchExists(project.Sync.Branch) {
			return "", fmt.Errorf("sync branch %q does not exist. Configure with: pm config set sync.branch <branch>", project.Sync.Branch)
		}
		return project.Sync.Branch, nil
	}

	branch, err := detectDefaultBranch()
	if err != nil {
		return "", err
	}
	return branch, nil
}

func detectDefaultBranch() (string, error) {
	if ref, err := gitOutput("", "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
		return strings.TrimPrefix(ref, "origin/"), nil
	}

	for _, candidate := range []string{"main", "master"} {
		if branchExists(candidate) {
			return candidate, nil
		}
	}

	branches, err := gitLines("", "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return "", fmt.Errorf("failed to detect default branch: %w", err)
	}
	if len(branches) == 0 {
		return "", fmt.Errorf("failed to detect default branch: no local branches found")
	}
	return branches[0], nil
}

func resolveRemoteTrackingRef(branch string) (string, string, error) {
	if _, err := gitOutput("", "show-ref", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		if upstream, err := gitOutput("", "rev-parse", "--abbrev-ref", branch+"@{upstream}"); err == nil {
			parts := strings.SplitN(upstream, "/", 2)
			if len(parts) == 2 {
				return upstream, parts[0], nil
			}
		}
	}

	remoteRefs, err := gitLines("", "for-each-ref", "--format=%(refname:short)", "refs/remotes")
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect remote branches: %w", err)
	}
	for _, remoteRef := range remoteRefs {
		if strings.HasSuffix(remoteRef, "/"+branch) {
			parts := strings.SplitN(remoteRef, "/", 2)
			if len(parts) == 2 {
				return remoteRef, parts[0], nil
			}
		}
	}

	return "", "", fmt.Errorf("no remote tracking branch for %q. Push it first with: git push -u origin %s", branch, branch)
}

func branchExists(branch string) bool {
	if _, err := gitOutput("", "show-ref", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		return true
	}
	if _, err := gitOutput("", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch); err == nil {
		return true
	}
	return false
}

func ensureNoInProgressGitOperation(gitDir string) error {
	paths := []string{
		filepath.Join(gitDir, "MERGE_HEAD"),
		filepath.Join(gitDir, "rebase-apply"),
		filepath.Join(gitDir, "rebase-merge"),
		filepath.Join(gitDir, "CHERRY_PICK_HEAD"),
		filepath.Join(gitDir, "REVERT_HEAD"),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("repository is in the middle of a rebase/merge. Complete it before syncing")
		}
	}
	return nil
}

func ensureNoUncommittedPMChanges() error {
	status, err := gitOutput("", "status", "--porcelain", "--", ".pm")
	if err != nil {
		return fmt.Errorf("failed to inspect working tree: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("you have uncommitted changes in .pm/. Commit or stash them before syncing.\nRun: git add .pm && git commit -m 'chore(pm): update tickets'")
	}
	return nil
}

// fetchSyncBranch refreshes the local remote-tracking ref for the configured
// sync branch before we compare or push `.pm` changes. It goes through
// gitOutput(), which eventually delegates to gitBytes() where exec.Command
// actually shells out to the `git` CLI.
func fetchSyncBranch(rt *syncRuntime) error {
	remoteBranch := strings.TrimPrefix(rt.upstream, rt.remote+"/")
	if _, err := gitOutput("", "fetch", rt.remote, remoteBranch); err != nil {
		return fmt.Errorf("failed to fetch %s from %s: %w", remoteBranch, rt.remote, err)
	}
	return nil
}

func diffPaths(base, ref string) ([]string, error) {
	return gitLines("", "diff", "--name-only", base+".."+ref, "--", ".pm")
}

func currentWorkspacePathsChanged(ref string) ([]string, error) {
	localPaths, err := listLocalPMPaths(".pm")
	if err != nil {
		return nil, err
	}
	refPaths, err := listRefPMPaths(ref)
	if err != nil {
		return nil, err
	}

	all := map[string]struct{}{}
	for _, path := range localPaths {
		all[path] = struct{}{}
	}
	for _, path := range refPaths {
		all[path] = struct{}{}
	}

	var changed []string
	for path := range all {
		same, err := pathMatchesRef(ref, path)
		if err != nil {
			return nil, err
		}
		if !same {
			changed = append(changed, path)
		}
	}
	slices.Sort(changed)
	return changed, nil
}

func intersectPaths(left, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, path := range right {
		rightSet[path] = struct{}{}
	}
	var out []string
	for _, path := range left {
		if _, ok := rightSet[path]; ok {
			out = append(out, path)
		}
	}
	slices.Sort(out)
	return out
}

func subtractPaths(paths, remove []string) []string {
	removeSet := map[string]struct{}{}
	for _, path := range remove {
		removeSet[path] = struct{}{}
	}
	var out []string
	for _, path := range paths {
		if _, ok := removeSet[path]; !ok {
			out = append(out, path)
		}
	}
	return out
}

func applyRefChanges(ref string, paths []string) error {
	for _, path := range paths {
		exists, err := refPathExists(ref, path)
		if err != nil {
			return err
		}
		if !exists {
			if err := os.RemoveAll(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
			continue
		}

		content, err := gitBytes("", "show", ref+":"+path)
		if err != nil {
			return fmt.Errorf("failed to read %s from %s: %w", path, ref, err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}
	return nil
}

func refPathExists(ref, path string) (bool, error) {
	cmd := exec.Command("git", "cat-file", "-e", ref+":"+path)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, fmt.Errorf("git cat-file failed for %s:%s: %w", ref, path, err)
	}
	return true, nil
}

func pushPMViaWorktree(rt *syncRuntime) error {
	worktreeDir, err := os.MkdirTemp("", "pm-sync-worktree-*")
	if err != nil {
		return fmt.Errorf("failed to create sync worktree: %w", err)
	}
	defer os.RemoveAll(worktreeDir)

	if _, err := gitOutput("", "worktree", "add", "--detach", worktreeDir, rt.upstream); err != nil {
		return fmt.Errorf("failed to create worktree for %s: %w", rt.upstream, err)
	}
	defer gitOutput("", "worktree", "remove", "--force", worktreeDir)

	dstPM := filepath.Join(worktreeDir, ".pm")
	if err := os.RemoveAll(dstPM); err != nil {
		return fmt.Errorf("failed to clear sync worktree .pm directory: %w", err)
	}
	if err := copyDir(rt.pmPath, dstPM); err != nil {
		return err
	}

	if _, err := gitOutput(worktreeDir, "add", ".pm"); err != nil {
		return fmt.Errorf("failed to stage .pm changes on %s: %w", rt.syncBranch, err)
	}

	status, err := gitOutput(worktreeDir, "status", "--porcelain", "--", ".pm")
	if err != nil {
		return fmt.Errorf("failed to inspect staged .pm changes: %w", err)
	}
	if strings.TrimSpace(status) == "" {
		return nil
	}

	message := fmt.Sprintf("chore(pm): sync tickets from %s", rt.currentRef)
	if _, err := gitOutput(worktreeDir, "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit .pm changes on %s: %w", rt.syncBranch, err)
	}
	if _, err := gitOutput(worktreeDir, "push", rt.remote, "HEAD:refs/heads/"+rt.syncBranch); err != nil {
		return fmt.Errorf("failed to push .pm changes to %s: %w\nSuggestion: run `pm sync pull` first", rt.syncBranch, err)
	}

	return nil
}

func copyDir(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source path %s does not exist: %w", src, err)
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if filepath.Base(path) == ".cache.db" {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func pathMatchesRef(ref, path string) (bool, error) {
	exists, err := refPathExists(ref, path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	refContent, err := gitBytes("", "show", ref+":"+path)
	if err != nil {
		return false, fmt.Errorf("failed to read %s from %s: %w", path, ref, err)
	}
	localContent, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read local path %s: %w", path, err)
	}
	return bytes.Equal(refContent, localContent), nil
}

func listLocalPMPaths(root string) ([]string, error) {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == ".cache.db" {
			return nil
		}
		paths = append(paths, filepath.ToSlash(path))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list local .pm paths: %w", err)
	}
	return paths, nil
}

func listRefPMPaths(ref string) ([]string, error) {
	paths, err := gitLines("", "ls-tree", "-r", "--name-only", ref, "--", ".pm")
	if err != nil {
		return nil, fmt.Errorf("failed to list .pm paths from %s: %w", ref, err)
	}
	var filtered []string
	for _, path := range paths {
		if filepath.Base(path) == ".cache.db" {
			continue
		}
		filtered = append(filtered, filepath.ToSlash(path))
	}
	return filtered, nil
}

func backupPMDir(src string) (string, bool, error) {
	if _, err := os.Stat(src); errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	backupDir, err := os.MkdirTemp("", "pm-sync-backup-*")
	if err != nil {
		return "", false, fmt.Errorf("failed to create backup directory: %w", err)
	}
	dst := filepath.Join(backupDir, ".pm")
	if err := copyDir(src, dst); err != nil {
		return "", false, fmt.Errorf("failed to back up .pm directory: %w", err)
	}
	return dst, true, nil
}

func restorePMDir(dst, backup string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return copyDir(backup, dst)
}

func normalizeConflictStrategy(strategy string) string {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "theirs", "ours", "manual":
		return strings.ToLower(strings.TrimSpace(strategy))
	default:
		return "prompt"
	}
}

func newSyncConflictError(direction, branch string, conflicts []string) error {
	labels := conflictedLabels(conflicts)
	return fmt.Errorf(
		"merge conflicts detected in .pm/ while preparing %s with %q.\nConflicted items: %s\nResolution options: set sync.conflict_strategy to theirs or ours, or resolve manually and retry",
		direction, branch, strings.Join(labels, ", "),
	)
}

func conflictedLabels(paths []string) []string {
	seen := map[string]struct{}{}
	var labels []string
	for _, path := range paths {
		label := path
		if strings.HasPrefix(path, ".pm/tickets/") {
			rest := strings.TrimPrefix(path, ".pm/tickets/")
			parts := strings.Split(rest, string(filepath.Separator))
			if len(parts) > 0 && parts[0] != "" {
				label = strings.TrimSuffix(parts[0], ".md")
			}
		}
		if _, ok := seen[label]; !ok {
			seen[label] = struct{}{}
			labels = append(labels, label)
		}
	}
	slices.Sort(labels)
	return labels
}

func gitOutput(dir string, args ...string) (string, error) {
	out, err := gitBytes(dir, args...)
	return strings.TrimSpace(string(out)), err
}

// gitBytes is the low-level git runner for this file. Higher-level helpers like
// gitOutput(), gitLines(), and fetchSyncBranch() all route through here, and
// this is the single place where we directly invoke the `git` command with
// exec.Command instead of going through a Go git library.
func gitBytes(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("%s", strings.TrimSpace(stderr.String()))
		}
		return nil, err
	}
	return out, nil
}

func gitLines(dir string, args ...string) ([]string, error) {
	out, err := gitOutput(dir, args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	lines := strings.Split(out, "\n")
	var filtered []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	return filtered, nil
}
