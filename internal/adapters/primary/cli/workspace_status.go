package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

// subCmdNameStatus reports the state of every repository of a workspace.
const subCmdNameStatus = "status"

// aheadBehindFieldCount is how many counters 'git rev-list --left-right
// --count HEAD...@{u}' prints: the commits only HEAD has, then the ones only
// the upstream has.
const aheadBehindFieldCount = 2

// workspaceStatusPayload is the graph the PubPascal desktop app renders.
//
// The field names below are the contract: window.loadGraph reads nodes[] and
// each node's id, name, root, ref, branch, ahead, behind, dirty, missing and
// writable. Renaming any of them silently blanks a part of the graph, because
// the page defaults every field it does not find instead of failing.
type workspaceStatusPayload struct {
	Nodes []workspaceStatusNode `json:"nodes"`
}

// workspaceStatusNode is one repository of the workspace.
type workspaceStatusNode struct {
	// ID is the directory the repository occupies: the workspace folder for the
	// root, modules/<id> for every dependency. The desktop prints it as the
	// module path and matches it against the boss.json dependency keys, so it
	// has to be the name on disk and not the portal's node UUID.
	ID string `json:"id"`
	// Name is the label shown on the node.
	Name string `json:"name"`
	// Root marks the PAI repository, the one that owns modules/.
	Root bool `json:"root"`
	// Ref is what the checkout points at: the branch name, or the tag (or short
	// commit) of a detached HEAD. For a repository missing from disk it falls
	// back to the reference the portal manifest pins.
	Ref string `json:"ref"`
	// Branch is the checked-out branch, empty on a detached HEAD.
	Branch string `json:"branch"`
	// Ahead and Behind count the commits between HEAD and its upstream. Both
	// stay zero when the branch tracks no remote, which is not an error.
	Ahead  int `json:"ahead"`
	Behind int `json:"behind"`
	// Dirty reports uncommitted changes among the repository's own files. The
	// nested modules/ clones are excluded: they are separate nodes of this very
	// graph, and counting them would paint every root dirty forever.
	Dirty bool `json:"dirty"`
	// Missing means the portal declares the repository but it is not a git
	// repository on disk -- not cloned yet, or a clone that failed halfway.
	Missing bool `json:"missing"`
	// Writable means the portal grants a push target for it. It is only ever
	// true when the manifest says so: without the manifest there is no evidence
	// of write access, and claiming it would offer a push that cannot succeed.
	Writable bool `json:"writable"`
}

// repoGitState is everything this command reads out of a local repository.
type repoGitState struct {
	ref    string
	branch string
	ahead  int
	behind int
	dirty  bool
}

// newWorkspaceStatusCmd builds 'boss workspace status'.
func newWorkspaceStatusCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   subCmdNameStatus + " [workspace-id]",
		Short: "Show status (ahead/behind/dirty) for each repository in the workspace",
		Long: "Show status (ahead/behind/dirty) for each repository in the workspace.\n\n" +
			"Without an argument the workspace is detected from the current directory and " +
			"only git is consulted. Given a workspace id -- or a \"<package-slug>@<version>\" " +
			"reference -- the portal manifest is fetched first, so the report also knows which " +
			"repositories are declared but absent from disk, and which ones you may push to.\n\n" +
			"With --json the result is printed on standard output as an object holding a " +
			"\"nodes\" array of id/name/root/ref/branch/ahead/behind/dirty/missing/writable " +
			"entries: the graph the PubPascal desktop app renders.",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ref := ""
			if len(args) == 1 {
				ref = args[0]
			}
			runWorkspaceStatusCommand(cmd.Context(), ref, asJSON)
		},
	}

	cmd.Flags().BoolVar(&asJSON, flagNameJSON, false, "print the workspace graph as JSON on standard output")

	return cmd
}

// runWorkspaceStatusCommand dispatches the four shapes of this command.
func runWorkspaceStatusCommand(ctx context.Context, ref string, asJSON bool) {
	ref = strings.TrimSpace(ref)

	// The report that existed before this flag is left exactly as it was: no
	// argument and no --json still walks the current directory and prints the
	// same lines, so nothing that reads 'boss workspace status' today breaks.
	if ref == "" && !asJSON {
		runWorkspaceStatus(ctx)

		return
	}

	if asJSON {
		// Everything printed on the way to the payload lands in the same pipe
		// the host reads, and the host recovers the payload by slicing from the
		// first brace to the last one. The progress chatter carries no brace,
		// but it would still end up quoted in the host log, so it is muted and
		// the payload -- printed last -- is all that is left on standard output.
		msg.SetQuietMode(true)
	}

	nodes := collectWorkspaceStatusNodes(ctx, ref)

	if asJSON {
		printJSONPayload(workspaceStatusPayload{Nodes: nodes})

		return
	}

	printWorkspaceStatusText(nodes)
}

// collectWorkspaceStatusNodes builds the node list for the requested workspace.
func collectWorkspaceStatusNodes(ctx context.Context, ref string) []workspaceStatusNode {
	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", flattenDetail(err.Error()))
	}

	if ref == "" {
		return localWorkspaceStatusNodes(ctx, cwd)
	}

	return manifestWorkspaceStatusNodes(ctx, cwd, ref)
}

// manifestWorkspaceStatusNodes correlates the portal manifest of a workspace
// with the repositories found on disk.
//
// The manifest is the only source for three of the fields: which repository is
// the root, which ones are writable, and -- by declaring them at all -- which
// ones are missing. git cannot answer any of those.
func manifestWorkspaceStatusNodes(ctx context.Context, cwd string, ref string) []workspaceStatusNode {
	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal configuration: %s", flattenDetail(err.Error()))
	}

	if config.AuthToken == "" {
		msg.Die("❌ You must log in first. Run 'boss login --token <token>' with your portal token.")
	}

	// The reference is echoed back by the resolver error messages, and a brace
	// reaching the host output would be mistaken for the start of the payload.
	// Rejecting it here keeps every failure path of this command brace-free
	// without silently mangling what the user typed.
	if strings.ContainsAny(ref, "{}") {
		msg.Die("❌ The workspace reference contains a brace, which is valid neither in a " +
			"workspace id nor in a <package-slug>@<version> reference.")
	}

	manifest := fetchWorkspaceManifest(ctx, config, resolveWorkspaceRef(ctx, config, ref))

	rootDir := resolveRootRepoName(manifest.Repos)
	if rootDir == "" {
		msg.Die("❌ The workspace manifest declares no root (PAI) repository, " +
			"so its dependencies cannot be located on disk.")
	}

	repos := orderReposRootFirst(manifest.Repos)
	nodes := make([]workspaceStatusNode, 0, len(repos))
	for _, repo := range repos {
		nodes = append(nodes, manifestRepoNode(ctx, cwd, rootDir, repo))
	}

	return nodes
}

// manifestRepoNode reports one declared repository, present on disk or not.
func manifestRepoNode(ctx context.Context, cwd string, rootDir string, repo ManifestRepo) workspaceStatusNode {
	dir := manifestRepoDirName(repo)

	node := workspaceStatusNode{
		ID:       dir,
		Name:     manifestRepoDisplayName(repo, dir),
		Root:     repo.IsRoot,
		Ref:      repo.Ref.Value,
		Branch:   "",
		Ahead:    0,
		Behind:   0,
		Dirty:    false,
		Missing:  true,
		Writable: repo.Writable,
	}

	if dir == "" {
		// Nothing in the entry maps to a directory name, so there is no place
		// on disk to look at -- which is also why 'boss workspace clone' skips
		// it. The node still appears, identified by the portal own node id:
		// dropping it would understate the workspace.
		node.ID = repo.NodeID
		node.Name = manifestRepoDisplayName(repo, repo.NodeID)

		return node
	}

	repoPath := filepath.Join(cwd, dir)
	if !repo.IsRoot {
		repoPath = filepath.Join(cwd, rootDir, modulesDirName, dir)
	}

	if !isGitRepo(repoPath) {
		return node
	}

	node.Missing = false
	applyGitState(&node, readRepoGitState(ctx, repoPath))

	return node
}

// applyGitState copies the live git facts onto a node, keeping the reference
// declared by the manifest when the checkout cannot name one.
func applyGitState(node *workspaceStatusNode, state repoGitState) {
	if state.ref != "" {
		node.Ref = state.ref
	}
	node.Branch = state.branch
	node.Ahead = state.ahead
	node.Behind = state.behind
	node.Dirty = state.dirty
}

// manifestRepoDirName returns the directory a manifest entry occupies on disk.
//
// The first choice is the same expression 'boss workspace clone' uses, so both
// commands agree on where a repository lives. The fallbacks only matter for an
// entry clone would have skipped outright: reporting it as missing is then the
// honest answer, and it beats dropping the node from the graph.
func manifestRepoDirName(repo ManifestRepo) string {
	if dir := repoShortName(repo.Name); dir != "" {
		return dir
	}
	if dir := repoShortName(repo.Slug); dir != "" {
		return dir
	}

	return repoShortName(strings.TrimSuffix(repo.CloneURL, ".git"))
}

// manifestRepoDisplayName returns the label of a node, falling back to its
// directory name when the portal has no name for it -- an external repository
// listed for someone who does not own the workspace carries none.
func manifestRepoDisplayName(repo ManifestRepo, fallback string) string {
	if name := strings.TrimSpace(repo.Name); name != "" {
		return name
	}
	if slug := strings.TrimSpace(repo.Slug); slug != "" {
		return slug
	}

	return fallback
}

// localWorkspaceStatusNodes reports the workspace found in the current
// directory, without asking the portal anything.
//
// This is the offline answer: it fills in every field git can prove and leaves
// missing and writable false, because "declared" and "you may push here" are
// statements only the manifest can make.
func localWorkspaceStatusNodes(ctx context.Context, cwd string) []workspaceStatusNode {
	nodes := make([]workspaceStatusNode, 0)
	seen := make(map[string]bool)

	add := func(repoPath string) {
		if seen[repoPath] {
			return
		}
		seen[repoPath] = true
		nodes = append(nodes, localRepoNode(ctx, repoPath))
	}

	// The desktop app spawns this CLI inside the folder the user opened, which
	// is the workspace root itself as often as it is the folder holding it.
	if isGitRepo(cwd) {
		add(cwd)
		for _, modulePath := range discoverModuleRepos(filepath.Join(cwd, modulesDirName)) {
			add(modulePath)
		}
	}

	for _, repoPath := range discoverWorkspaceRepos(cwd) {
		add(repoPath)
	}

	return orderNodesRootFirst(nodes)
}

// localRepoNode reports a repository discovered on disk.
func localRepoNode(ctx context.Context, repoPath string) workspaceStatusNode {
	name := filepath.Base(repoPath)
	state := readRepoGitState(ctx, repoPath)

	return workspaceStatusNode{
		ID:     name,
		Name:   name,
		Root:   ownsModulesDir(repoPath),
		Ref:    state.ref,
		Branch: state.branch,
		Ahead:  state.ahead,
		Behind: state.behind,
		Dirty:  state.dirty,
		// Missing stays false: nothing was declared, so nothing can be absent.
		Missing: false,
		// Writable stays false on purpose. See workspaceStatusNode.Writable.
		Writable: false,
	}
}

// ownsModulesDir reports whether a repository holds the dependency clones of a
// workspace, which is how the root (PAI) repository is recognised on disk.
func ownsModulesDir(repoPath string) bool {
	info, err := os.Stat(filepath.Join(repoPath, modulesDirName))

	return err == nil && info.IsDir()
}

// orderNodesRootFirst moves the root nodes to the front, keeping the relative
// order of everything else, so the report does not depend on the order the
// directories happen to be listed in.
func orderNodesRootFirst(nodes []workspaceStatusNode) []workspaceStatusNode {
	ordered := make([]workspaceStatusNode, 0, len(nodes))
	for _, node := range nodes {
		if node.Root {
			ordered = append(ordered, node)
		}
	}
	for _, node := range nodes {
		if !node.Root {
			ordered = append(ordered, node)
		}
	}

	return ordered
}

// readRepoGitState reads the branch, the checked-out reference, the
// ahead/behind counters and the dirty flag of one repository.
//
// Every step degrades to its zero value instead of failing: a repository with
// no upstream, no tag or no commit at all is a normal member of a workspace,
// and refusing to report the rest of its state would blank the whole graph.
func readRepoGitState(ctx context.Context, repoPath string) repoGitState {
	var state repoGitState

	// symbolic-ref answers only on a real branch, which is what tells a branch
	// checkout apart from a detached HEAD: 'rev-parse --abbrev-ref HEAD' says
	// "HEAD" for the detached case, and that would be read as a branch name.
	if out, ok := gitCapture(ctx, repoPath, "symbolic-ref", "--quiet", "--short", gitHeadRef); ok {
		state.branch = strings.TrimSpace(out)
	}

	state.ref = state.branch
	if state.ref == "" {
		state.ref = detachedHeadRef(ctx, repoPath)
	}

	if out, ok := gitCapture(ctx, repoPath, "rev-list", "--left-right", "--count",
		gitHeadRef+"...@{u}"); ok {
		state.ahead, state.behind = parseAheadBehind(out)
	}

	if status, ok := workspaceRepoStatus(ctx, repoPath); ok {
		state.dirty = status != ""
	}

	return state
}

// detachedHeadRef names the commit a detached HEAD sits on: the exact tag when
// there is one -- which is how a repository pinned to a version reads -- and
// the short commit id otherwise.
func detachedHeadRef(ctx context.Context, repoPath string) string {
	if out, ok := gitCapture(ctx, repoPath, "describe", "--tags", "--exact-match"); ok {
		if tag := strings.TrimSpace(out); tag != "" {
			return tag
		}
	}

	if out, ok := gitCapture(ctx, repoPath, "rev-parse", "--short", gitHeadRef); ok {
		return strings.TrimSpace(out)
	}

	return ""
}

// parseAheadBehind reads the two counters git prints, returning zeros for any
// output that is not exactly two numbers.
func parseAheadBehind(out string) (int, int) {
	fields := strings.Fields(out)
	if len(fields) != aheadBehindFieldCount {
		return 0, 0
	}

	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0
	}

	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0
	}

	return ahead, behind
}

// printWorkspaceStatusText prints the human-readable report of a node list.
func printWorkspaceStatusText(nodes []workspaceStatusNode) {
	if len(nodes) == 0 {
		msg.Info("No repository found for this workspace in the current directory.")

		return
	}

	for _, node := range nodes {
		label := node.ID
		if node.Root {
			label += " (Root)"
		}

		// Written straight to standard output to keep the aligned columns of
		// the report that existed before this command learned about the portal.
		_, _ = fmt.Fprintf(os.Stdout, "%-35s [%s] %s\n", label, nodeStatusWord(node), nodeStatusDetail(node))
	}
}

// nodeStatusWord reduces a node to the single word the report shows first.
func nodeStatusWord(node workspaceStatusNode) string {
	switch {
	case node.Missing:
		return "missing"
	case node.Dirty:
		return "dirty"
	default:
		return "clean"
	}
}

// nodeStatusDetail spells out the reference, the branch and the divergence.
func nodeStatusDetail(node workspaceStatusNode) string {
	if node.Missing {
		if node.Ref == "" {
			return "not cloned in this directory"
		}

		return "not cloned in this directory, pinned to " + node.Ref
	}

	detail := "ref: " + node.Ref
	if node.Branch != "" && node.Branch != node.Ref {
		detail += ", branch: " + node.Branch
	}
	if node.Ahead != 0 || node.Behind != 0 {
		detail += fmt.Sprintf(" (ahead %d, behind %d)", node.Ahead, node.Behind)
	}
	if node.Writable {
		detail += ", writable"
	}

	return detail
}
