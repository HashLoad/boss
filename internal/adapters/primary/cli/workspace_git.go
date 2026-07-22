package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

// Names of the workspace sub-commands that only drive git.
const (
	subCmdNameDiff   = "diff"
	subCmdNamePull   = "pull"
	subCmdNameCommit = "commit"
)

// excludeModulesPathspec keeps a git command inside the repository's own files.
//
// A workspace root holds its dependency repositories under modules/, and each
// of those is a git repository of its own that discoverWorkspaceRepos already
// reports separately. Without this pathspec, 'git status' in the root would
// flag the whole modules/ tree as an untracked change forever, and 'git add -A'
// would stage the nested clones as gitlinks -- committing a bogus submodule
// reference into the PAI repository.
const excludeModulesPathspec = ":(exclude)" + modulesDirName

// gitHeadRef is the symbolic name of the current checkout. It is also what
// 'git rev-parse --abbrev-ref HEAD' echoes back when the checkout is detached,
// which is how a repository pinned to a tag or a commit is recognised.
const gitHeadRef = "HEAD"

// workspaceDiffPayload is the contract consumed by the PubPascal desktop app,
// which renders one collapsible block per repository.
type workspaceDiffPayload struct {
	Repos []workspaceRepoDiff `json:"repos"`
}

// workspaceRepoDiff is the uncommitted diff of a single workspace repository.
type workspaceRepoDiff struct {
	Name string `json:"name"`
	Diff string `json:"diff"`
}

// pullOutcome is the result of fast-forwarding a single repository.
type pullOutcome int

const (
	pullUpdated pullOutcome = iota
	pullSkipped
	pullFailed
)

// pullTally counts the outcomes of a whole workspace pull.
type pullTally struct {
	updated int
	skipped int
	failed  int
}

// newWorkspaceDiffCmd builds 'boss workspace diff'.
func newWorkspaceDiffCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   subCmdNameDiff,
		Short: "Show the uncommitted changes of every repository in the workspace",
		Long: "Show the uncommitted changes of every repository in the workspace.\n\n" +
			"With --json the result is printed on standard output as an object holding " +
			"a \"repos\" array of name/diff entries. Only repositories that actually " +
			"have changes are listed; a repository whose only change is an untracked " +
			"file appears with an empty diff, because an untracked file has nothing to " +
			"compare against HEAD.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspaceDiff(cmd.Context(), asJSON)
		},
	}

	cmd.Flags().BoolVar(&asJSON, flagNameJSON, false, "print the diff as JSON on standard output")

	return cmd
}

// newWorkspacePullCmd builds 'boss workspace pull'.
func newWorkspacePullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   subCmdNamePull,
		Short: "Fast-forward every repository in the workspace",
		Long: "Fast-forward every repository in the workspace.\n\n" +
			"Repositories pinned to a tag or commit, and branches that track no " +
			"remote, are reported as skipped rather than failed: there is nothing to " +
			"fast-forward. Any other failure makes the command exit non-zero.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspacePull(cmd.Context())
		},
	}
}

// newWorkspaceCommitCmd builds 'boss workspace commit'.
func newWorkspaceCommitCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   subCmdNameCommit,
		Short: "Commit the pending changes of every modified repository in the workspace",
		Long: "Commit the pending changes of every modified repository in the workspace.\n\n" +
			"Each repository stages its own files -- the nested modules/ clones are " +
			"left alone -- and commits them with the given message. Nothing is pushed: " +
			"run 'boss workspace push' for that.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspaceCommit(cmd.Context(), message)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "commit message (required)")

	return cmd
}

// requireWorkspaceRepos returns the git repositories of the workspace rooted at
// the current directory, refusing to report success when there are none.
//
// discoverWorkspaceRepos returns an empty slice both for "the workspace is
// clean" and for "you are in the wrong folder", and the second case used to be
// indistinguishable from a no-op success.
func requireWorkspaceRepos() []string {
	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", flattenDetail(err.Error()))
	}

	repos := discoverWorkspaceRepos(cwd)
	if len(repos) == 0 {
		msg.Die("❌ No git repository found under %s. "+
			"Run this from the folder where the workspace was cloned.", cwd)
	}

	return repos
}

// workspaceRepoStatus returns the porcelain status of a repository's own files.
func workspaceRepoStatus(ctx context.Context, repoPath string) (string, bool) {
	out, ok := gitCapture(ctx, repoPath, "status", "--porcelain", "--", ".", excludeModulesPathspec)

	return strings.TrimSpace(out), ok
}

// runWorkspaceDiff collects the uncommitted changes of the whole workspace.
func runWorkspaceDiff(ctx context.Context, asJSON bool) {
	repos := requireWorkspaceRepos()

	payload := workspaceDiffPayload{Repos: make([]workspaceRepoDiff, 0, len(repos))}
	for _, repoPath := range repos {
		status, ok := workspaceRepoStatus(ctx, repoPath)
		if !ok {
			msg.Warn("  %s: could not be read as a git repository, skipping.", filepath.Base(repoPath))

			continue
		}
		if status == "" {
			continue
		}

		// A diff against HEAD covers staged and unstaged edits to tracked
		// files. Untracked files never appear in one, so a repository whose
		// only change is a new file is still listed, with an empty diff.
		diff, _ := gitCapture(ctx, repoPath, "diff", gitHeadRef, "--", ".", excludeModulesPathspec)
		payload.Repos = append(payload.Repos, workspaceRepoDiff{
			Name: filepath.Base(repoPath),
			Diff: diff,
		})
	}

	if asJSON {
		printJSONPayload(payload)

		return
	}

	printWorkspaceDiffText(payload)
}

// printWorkspaceDiffText prints the human-readable diff report.
func printWorkspaceDiffText(payload workspaceDiffPayload) {
	if len(payload.Repos) == 0 {
		msg.Info("Nothing to diff: every repository in the workspace is clean.")

		return
	}

	for _, repo := range payload.Repos {
		msg.Info("=== %s ===", repo.Name)
		if strings.TrimSpace(repo.Diff) == "" {
			msg.Info("  (untracked changes only, nothing to diff against HEAD)")

			continue
		}
		// Written straight to stdout: a diff is full of '%' and would be
		// mangled by a format-string logger.
		_, _ = fmt.Fprintln(os.Stdout, repo.Diff)
	}
}

// runWorkspacePull fast-forwards every repository and fails loudly if any real
// pull failed. Skipped repositories do not make the command fail.
func runWorkspacePull(ctx context.Context) {
	msg.Info("Pulling every repository in the workspace...")

	tally := pullWorkspaceRepos(ctx, requireWorkspaceRepos())

	msg.Info("Pull summary: %d updated, %d skipped, %d failed.",
		tally.updated, tally.skipped, tally.failed)

	if tally.failed > 0 {
		msg.Die("❌ %d repository(ies) could not be fast-forwarded.", tally.failed)
	}
}

// pullWorkspaceRepos fast-forwards each repository and tallies the outcomes.
func pullWorkspaceRepos(ctx context.Context, repos []string) pullTally {
	var tally pullTally

	for _, repoPath := range repos {
		switch pullWorkspaceRepo(ctx, repoPath) {
		case pullUpdated:
			tally.updated++
		case pullSkipped:
			tally.skipped++
		case pullFailed:
			tally.failed++
		}
	}

	return tally
}

// pullWorkspaceRepo fast-forwards a single repository.
//
// Two situations are skips rather than failures, because a workspace is
// expected to contain them: a repository checked out at a pinned tag or commit
// has a detached HEAD and no branch to advance, and a local-only branch has no
// upstream to pull from. Calling either one a failure would paint a correctly
// pinned workspace red on every run.
func pullWorkspaceRepo(ctx context.Context, repoPath string) pullOutcome {
	name := filepath.Base(repoPath)

	branchOut, ok := gitCapture(ctx, repoPath, "rev-parse", "--abbrev-ref", gitHeadRef)
	if !ok {
		msg.Err("  %s: could not be read as a git repository.", name)

		return pullFailed
	}

	branch := strings.TrimSpace(branchOut)
	if branch == gitHeadRef || branch == "" {
		msg.Info("  %s: skipped, detached HEAD (pinned to a fixed reference).", name)

		return pullSkipped
	}

	if _, hasUpstream := gitCapture(ctx, repoPath, "rev-parse", "--abbrev-ref",
		"--symbolic-full-name", "@{u}"); !hasUpstream {
		msg.Info("  %s: skipped, branch %s tracks no remote.", name, branch)

		return pullSkipped
	}

	msg.Info("  %s: fast-forwarding %s...", name, branch)
	if _, err := runGitCmd(ctx, repoPath, "pull", "--ff-only"); err != nil {
		msg.Err("  %s: git pull failed: %s", name, flattenDetail(err.Error()))

		return pullFailed
	}

	return pullUpdated
}

// runWorkspaceCommit commits the pending changes of every modified repository.
//
// Nothing is pushed: 'boss workspace push' is the command that publishes, and
// committing and pushing in one step would take the decision away from whoever
// is still reviewing the change.
func runWorkspaceCommit(ctx context.Context, message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		msg.Die("❌ A commit message is required. Run 'boss workspace commit -m \"your message\"'.")
	}

	repos := requireWorkspaceRepos()

	committed := 0
	pinned := 0
	failed := 0
	for _, repoPath := range repos {
		switch commitWorkspaceRepo(ctx, repoPath, message) {
		case commitDone:
			committed++
		case commitClean:
		case commitPinned:
			pinned++
		case commitFailed:
			failed++
		}
	}

	msg.Info("Commit summary: %d committed, %d pinned and left alone, %d failed.",
		committed, pinned, failed)

	if failed > 0 {
		msg.Die("❌ %d repository(ies) could not be committed.", failed)
	}
	if committed == 0 && pinned > 0 {
		msg.Die("❌ Nothing was committed: the repositories with changes are all pinned.")
	}
	if committed == 0 && pinned == 0 {
		msg.Info("Nothing to commit: every repository in the workspace is clean.")
	}
}

// commitOutcome is the result of committing a single repository.
type commitOutcome int

const (
	commitDone commitOutcome = iota
	commitClean
	commitPinned
	commitFailed
)

// commitWorkspaceRepo stages and commits one repository's own changes.
func commitWorkspaceRepo(ctx context.Context, repoPath string, message string) commitOutcome {
	name := filepath.Base(repoPath)

	status, ok := workspaceRepoStatus(ctx, repoPath)
	if !ok {
		msg.Err("  %s: could not be read as a git repository.", name)

		return commitFailed
	}
	if status == "" {
		return commitClean
	}

	// A repository pinned to a tag or a commit sits on a detached HEAD.
	// Committing there succeeds and then strands the commit: no branch points
	// at it, 'boss workspace push' has nothing to push, and the next checkout
	// drops it. Editing a pinned dependency is what 'boss contribute' is for.
	if branch, branchOK := gitCapture(ctx, repoPath, "rev-parse", "--abbrev-ref", gitHeadRef); branchOK {
		if trimmed := strings.TrimSpace(branch); trimmed == gitHeadRef || trimmed == "" {
			msg.Warn("  %s: has changes but is pinned to a fixed reference (detached HEAD), "+
				"so nothing was committed. Your edits are untouched in the working tree. "+
				"Run 'boss contribute' to turn them into a pull request.", name)

			return commitPinned
		}
	}

	msg.Info("  %s: staging and committing...", name)
	if _, err := runGitCmd(ctx, repoPath, "add", "-A", "--", ".", excludeModulesPathspec); err != nil {
		msg.Err("  %s: git add failed: %s", name, flattenDetail(err.Error()))

		return commitFailed
	}

	if _, err := runGitCmd(ctx, repoPath, "commit", "-m", message); err != nil {
		msg.Err("  %s: git commit failed: %s", name, flattenDetail(err.Error()))

		return commitFailed
	}

	return commitDone
}
