//nolint:testpackage // exercises unexported command plumbing
package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkspaceStatusPayloadShape pins every field name the PubPascal desktop
// graph reads. window.loadGraph defaults whatever it does not find, so a
// renamed field does not fail: it silently blanks part of the graph.
func TestWorkspaceStatusPayloadShape(t *testing.T) {
	out, err := json.Marshal(workspaceStatusPayload{Nodes: []workspaceStatusNode{{
		ID:       "janus",
		Name:     "Janus",
		Root:     true,
		Ref:      "v2.22.5",
		Branch:   "main",
		Ahead:    2,
		Behind:   1,
		Dirty:    true,
		Missing:  false,
		Writable: true,
	}}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	const want = `{"nodes":[{"id":"janus","name":"Janus","root":true,"ref":"v2.22.5",` +
		`"branch":"main","ahead":2,"behind":1,"dirty":true,"missing":false,"writable":true}]}`
	if string(out) != want {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

// TestWorkspaceStatusEmptyNodesMarshalAsArray checks that a workspace with no
// repository is an empty array and not null: the page calls .map on it.
func TestWorkspaceStatusEmptyNodesMarshalAsArray(t *testing.T) {
	out, err := json.Marshal(workspaceStatusPayload{Nodes: []workspaceStatusNode{}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"nodes":[]}` {
		t.Errorf("got %s, want %s", out, `{"nodes":[]}`)
	}
}

// TestParseAheadBehind covers the shapes git can hand back, including the
// failure of 'rev-list HEAD...@{u}' on a branch that tracks no remote.
func TestParseAheadBehind(t *testing.T) {
	cases := []struct {
		in     string
		ahead  int
		behind int
	}{
		{"2\t1\n", 2, 1},
		{"0\t0\n", 0, 0},
		{"", 0, 0},
		{"fatal: no upstream configured", 0, 0},
		{"7", 0, 0},
	}

	for _, tc := range cases {
		ahead, behind := parseAheadBehind(tc.in)
		if ahead != tc.ahead || behind != tc.behind {
			t.Errorf("%q gave (%d,%d), want (%d,%d)", tc.in, ahead, behind, tc.ahead, tc.behind)
		}
	}
}

// TestManifestRepoDirNameFallsBack proves a node still gets an identity when
// the portal sends no name -- an external repository seen by a non-owner has
// name and slug null.
func TestManifestRepoDirNameFallsBack(t *testing.T) {
	cases := []struct {
		name string
		repo ManifestRepo
		want string
	}{
		{"name wins", ManifestRepo{Name: "isaquepinheiro/janus", Slug: "janus",
			CloneURL: "https://github.com/isaquepinheiro/other.git"}, "janus"},
		{"slug when unnamed", ManifestRepo{Slug: "fluentquery",
			CloneURL: "https://github.com/acme/other.git"}, "fluentquery"},
		{"clone url last", ManifestRepo{CloneURL: "https://github.com/acme/dataengine.git"}, "dataengine"},
		{"nothing usable", ManifestRepo{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := manifestRepoDirName(tc.repo); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestOrderNodesRootFirst keeps the report independent of the order the
// directories happen to be listed in.
func TestOrderNodesRootFirst(t *testing.T) {
	got := orderNodesRootFirst([]workspaceStatusNode{
		{ID: "a"}, {ID: "root", Root: true}, {ID: "b"},
	})

	want := []string{"root", "a", "b"}
	for i, node := range got {
		if node.ID != want[i] {
			t.Fatalf("position %d is %q, want %q", i, node.ID, want[i])
		}
	}
}

// testMainBranch is the branch every test repository is created on. It is a
// constant so the literal is not repeated across the file.
const testMainBranch = "main"

// requireGit skips a test when git is not on PATH.
func requireGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
}

// gitEnv isolates the test repositories from the developer's own git
// configuration and identity, so the assertions do not depend on the machine
// running them.
func gitEnv() []string {
	return append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_AUTHOR_NAME=boss test",
		"GIT_AUTHOR_EMAIL=test@example.invalid",
		"GIT_COMMITTER_NAME=boss test",
		"GIT_COMMITTER_EMAIL=test@example.invalid",
		"GIT_TERMINAL_PROMPT=0",
	)
}

// runGit runs git in dir and fails the test when it does not succeed.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = gitEnv()

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// writeFile creates or overwrites a file inside a test repository.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// initRepo creates a repository with one commit on a branch named main.
func initRepo(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, dir, "-c", "init.defaultBranch="+testMainBranch, "init")
	writeFile(t, filepath.Join(dir, "README.md"), "first\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "first")
}

// TestReadRepoGitStateOnBranch covers the ordinary case: a branch checkout with
// no upstream is clean, at zero/zero, and names the branch in both ref and
// branch -- which is how the desktop hides the redundant tracking chip.
func TestReadRepoGitStateOnBranch(t *testing.T) {
	requireGit(t)

	dir := filepath.Join(t.TempDir(), "janus")
	initRepo(t, dir)

	state := readRepoGitState(t.Context(), dir)
	if state.branch != testMainBranch || state.ref != testMainBranch {
		t.Errorf("got ref %q branch %q, want both %q", state.ref, state.branch, testMainBranch)
	}
	if state.ahead != 0 || state.behind != 0 {
		t.Errorf("got ahead %d behind %d on a branch with no upstream, want 0/0", state.ahead, state.behind)
	}
	if state.dirty {
		t.Error("a fresh repository was reported dirty")
	}

	writeFile(t, filepath.Join(dir, "untracked.pas"), "unit untracked;\n")

	if !readRepoGitState(t.Context(), dir).dirty {
		t.Error("an untracked file did not make the repository dirty")
	}
}

// TestReadRepoGitStateDetachedTag is the pinned dependency: the checkout has no
// branch, and ref must name the tag rather than the literal "HEAD" that
// 'rev-parse --abbrev-ref' returns.
func TestReadRepoGitStateDetachedTag(t *testing.T) {
	requireGit(t)

	dir := filepath.Join(t.TempDir(), "fluentquery")
	initRepo(t, dir)
	runGit(t, dir, "tag", "v2.22.5")
	runGit(t, dir, "checkout", "v2.22.5")

	state := readRepoGitState(t.Context(), dir)
	if state.branch != "" {
		t.Errorf("a detached HEAD reported branch %q, want empty", state.branch)
	}
	if state.ref != "v2.22.5" {
		t.Errorf("got ref %q, want %q", state.ref, "v2.22.5")
	}
}

// TestReadRepoGitStateAheadAndBehind proves the counters against a real
// upstream, in the order the desktop displays them.
func TestReadRepoGitStateAheadAndBehind(t *testing.T) {
	requireGit(t)

	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	if err := os.MkdirAll(origin, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	runGit(t, origin, "-c", "init.defaultBranch="+testMainBranch, "init", "--bare")

	seed := filepath.Join(root, "seed")
	initRepo(t, seed)
	runGit(t, seed, "remote", "add", "origin", origin)
	runGit(t, seed, "push", "-u", "origin", testMainBranch)

	work := filepath.Join(root, "work")
	runGit(t, root, "clone", origin, work)

	// The upstream moves once, the clone moves once: one commit each way.
	writeFile(t, filepath.Join(seed, "upstream.pas"), "unit upstream;\n")
	runGit(t, seed, "add", ".")
	runGit(t, seed, "commit", "-m", "upstream")
	runGit(t, seed, "push")

	writeFile(t, filepath.Join(work, "local.pas"), "unit local;\n")
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "local")
	runGit(t, work, "fetch")

	state := readRepoGitState(t.Context(), work)
	if state.ahead != 1 || state.behind != 1 {
		t.Errorf("got ahead %d behind %d, want 1/1", state.ahead, state.behind)
	}
	if state.dirty {
		t.Error("a committed repository was reported dirty")
	}
}

// TestRootRepoIgnoresModulesWhenDirty is the regression that matters most for
// the graph: the dependency clones live under the root's modules/ directory and
// are untracked there, so counting them would paint the PAI dirty forever.
func TestRootRepoIgnoresModulesWhenDirty(t *testing.T) {
	requireGit(t)

	root := filepath.Join(t.TempDir(), "janus")
	initRepo(t, root)

	dep := filepath.Join(root, modulesDirName, "fluentquery")
	initRepo(t, dep)
	writeFile(t, filepath.Join(dep, "changed.pas"), "unit changed;\n")

	if readRepoGitState(t.Context(), root).dirty {
		t.Error("the root repository was reported dirty because of its modules/ clones")
	}
	if !readRepoGitState(t.Context(), dep).dirty {
		t.Error("the dependency with a new file was not reported dirty")
	}
}

// TestManifestRepoNodeCorrelatesWithDisk walks the three states a declared
// repository can be in: the root present on disk, a dependency present under
// modules/, and a dependency the portal declares but nobody cloned.
func TestManifestRepoNodeCorrelatesWithDisk(t *testing.T) {
	requireGit(t)

	cwd := t.TempDir()
	rootDir := "janus"
	initRepo(t, filepath.Join(cwd, rootDir))

	dep := filepath.Join(cwd, rootDir, modulesDirName, "fluentquery")
	initRepo(t, dep)
	runGit(t, dep, "tag", "v1.4.0")
	runGit(t, dep, "checkout", "v1.4.0")

	rootNode := manifestRepoNode(t.Context(), cwd, rootDir, ManifestRepo{
		NodeID: "11111111-1111-4111-8111-111111111111",
		Name:   "isaquepinheiro/janus", Slug: "janus", IsRoot: true, Writable: true,
		Ref: ManifestRef{Kind: refKindVersion, Value: "2.22.5"},
	})
	if rootNode.ID != "janus" || !rootNode.Root || rootNode.Missing {
		t.Errorf("root node wrong: %+v", rootNode)
	}
	if !rootNode.Writable {
		t.Error("the manifest said the root is writable and the node did not")
	}
	// The checkout wins over the declared pin: the graph reports what is on
	// disk, and the two differ exactly when the workspace is out of date.
	if rootNode.Ref != testMainBranch || rootNode.Branch != testMainBranch {
		t.Errorf("root ref %q branch %q, want both main", rootNode.Ref, rootNode.Branch)
	}

	depNode := manifestRepoNode(t.Context(), cwd, rootDir, ManifestRepo{
		NodeID: "22222222-2222-4222-8222-222222222222",
		Name:   "fluentquery", Slug: "fluentquery",
		Ref: ManifestRef{Kind: refKindTag, Value: "v1.4.0"},
	})
	if depNode.Missing || depNode.Root || depNode.Writable {
		t.Errorf("dependency node wrong: %+v", depNode)
	}
	if depNode.Ref != "v1.4.0" || depNode.Branch != "" {
		t.Errorf("pinned dependency ref %q branch %q, want v1.4.0 and no branch", depNode.Ref, depNode.Branch)
	}

	absent := manifestRepoNode(t.Context(), cwd, rootDir, ManifestRepo{
		NodeID: "33333333-3333-4333-8333-333333333333",
		Name:   "acme/dataengine", Slug: "dataengine",
		Ref: ManifestRef{Kind: refKindTag, Value: "v3.0.0"},
	})
	if !absent.Missing {
		t.Error("a repository that was never cloned was not reported missing")
	}
	if absent.ID != "dataengine" || absent.Ref != "v3.0.0" {
		t.Errorf("missing node lost its identity: %+v", absent)
	}
	if absent.Dirty || absent.Ahead != 0 || absent.Behind != 0 {
		t.Errorf("a missing repository reported git state: %+v", absent)
	}
}

// TestManifestRepoNodeKeepsUnnamedEntries makes sure a manifest entry this CLI
// cannot map to a directory still reaches the graph, flagged missing, instead
// of shrinking the workspace to the repositories with a usable name.
func TestManifestRepoNodeKeepsUnnamedEntries(t *testing.T) {
	node := manifestRepoNode(t.Context(), t.TempDir(), "janus", ManifestRepo{
		NodeID: "44444444-4444-4444-8444-444444444444",
	})

	if node.ID != "44444444-4444-4444-8444-444444444444" || !node.Missing {
		t.Errorf("unnamed entry wrong: %+v", node)
	}
}

// TestLocalWorkspaceStatusNodesFindsRootAndModules covers the argument-less
// run: no portal, no token, the workspace read straight off the disk.
func TestLocalWorkspaceStatusNodesFindsRootAndModules(t *testing.T) {
	requireGit(t)

	cwd := t.TempDir()
	root := filepath.Join(cwd, "janus")
	initRepo(t, root)
	initRepo(t, filepath.Join(root, modulesDirName, "fluentquery"))
	initRepo(t, filepath.Join(cwd, "unrelated"))

	nodes := localWorkspaceStatusNodes(t.Context(), cwd)
	if len(nodes) != 3 {
		t.Fatalf("found %d repositories, want 3: %+v", len(nodes), nodes)
	}
	if nodes[0].ID != "janus" || !nodes[0].Root {
		t.Errorf("the root did not come first: %+v", nodes)
	}

	for _, node := range nodes {
		if node.Missing {
			t.Errorf("%s was reported missing while sitting on disk", node.ID)
		}
		// Without the manifest there is no evidence of write access, and
		// claiming it would offer the desktop a push that cannot succeed.
		if node.Writable {
			t.Errorf("%s was reported writable without a manifest saying so", node.ID)
		}
	}
}

// TestLocalWorkspaceStatusNodesAcceptsRootAsCwd covers the desktop app being
// opened on the workspace root itself rather than on the folder holding it.
func TestLocalWorkspaceStatusNodesAcceptsRootAsCwd(t *testing.T) {
	requireGit(t)

	root := filepath.Join(t.TempDir(), "janus")
	initRepo(t, root)
	initRepo(t, filepath.Join(root, modulesDirName, "fluentquery"))

	nodes := localWorkspaceStatusNodes(t.Context(), root)
	if len(nodes) != 2 {
		t.Fatalf("found %d repositories, want 2: %+v", len(nodes), nodes)
	}
	if !nodes[0].Root || nodes[0].ID != "janus" {
		t.Errorf("the root was not recognised: %+v", nodes)
	}
	if nodes[1].ID != "fluentquery" || nodes[1].Root {
		t.Errorf("the dependency was not recognised: %+v", nodes)
	}
}
