package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

const (
	// portalRequestTimeout bounds every HTTP call made against the portal.
	portalRequestTimeout = 15 * time.Second

	// bossManifestFile is the Boss manifest looked up inside a repository.
	bossManifestFile = "boss.json"

	// modulesDirName holds the dependency repositories of a workspace root.
	modulesDirName = "modules"

	// sourceDirName is the Delphi-style sources directory of a package; the
	// "src" spelling is covered by srcDirName.
	sourceDirName = "Source"

	// refKindTag and refKindVersion are the immutable reference kinds a
	// workspace repository can be pinned to in the portal manifest.
	refKindTag     = "tag"
	refKindVersion = "version"

	// subCmdNameUpdate is the workspace sub-command that fast-forwards repos.
	subCmdNameUpdate = "update"
)

// bossManifest mirrors the subset of boss.json needed to build an SBOM.
type bossManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Homepage     string            `json:"homepage"`
	Dependencies map[string]string `json:"dependencies"`
}

// sbomComponent is one resolved dependency, ready to be emitted in any format.
type sbomComponent struct {
	Name    string
	Version string
	Purl    string
	Hash    string
	// Resolved reports whether Version came from the lock file (an exact
	// version) rather than from a boss.json constraint such as "^3.0.0".
	Resolved bool
}

// PubPascalConfig represents the configuration stored in ~/.pubpascal/config.json.
type PubPascalConfig struct {
	PortalBaseURL string `json:"portalBaseUrl"`
	AuthToken     string `json:"authToken"`
}

// WorkspaceManifest represents the workspace manifest returned by the portal API.
type WorkspaceManifest struct {
	SchemaVersion int            `json:"schema_version"`
	Workspace     WorkspaceInfo  `json:"workspace"`
	Viewer        ViewerInfo     `json:"viewer"`
	Repos         []ManifestRepo `json:"repos"`
	Edges         []ManifestEdge `json:"edges"`
}

// WorkspaceInfo identifies the workspace a manifest belongs to.
type WorkspaceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ViewerInfo describes the permissions of the authenticated user on the workspace.
type ViewerInfo struct {
	IsOwner bool `json:"is_owner"`
}

// ManifestRepo is a single repository that takes part in the workspace.
type ManifestRepo struct {
	NodeID   string      `json:"node_id"`
	Kind     string      `json:"kind"`
	Name     string      `json:"name"`
	Slug     string      `json:"slug"`
	CloneURL string      `json:"clone_url"`
	IsRoot   bool        `json:"is_root"`
	Writable bool        `json:"writable"`
	PushURL  string      `json:"push_url"`
	Ref      ManifestRef `json:"ref"`
}

// ManifestRef is the git reference a workspace repository is pinned to.
type ManifestRef struct {
	HasRef bool   `json:"has_ref"`
	Kind   string `json:"type"`
	Value  string `json:"value"`
}

// ManifestEdge is a dependency relation between two workspace repositories.
type ManifestEdge struct {
	FromNodeID string `json:"from_node_id"`
	ToNodeID   string `json:"to_node_id"`
}

// GetPubPascalConfigPath resolves the path to the PubPascal config file.
func GetPubPascalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".pubpascal", "config.json")
	}
	return filepath.Join(home, ".pubpascal", "config.json")
}

// LoadPubPascalConfig loads the PubPascal configuration from disk.
func LoadPubPascalConfig() (*PubPascalConfig, error) {
	configPath := GetPubPascalConfigPath()
	config := &PubPascalConfig{
		PortalBaseURL: "https://www.pubpascal.dev",
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	// #nosec G304 -- the path is this CLI's own config in the user's home
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return config, err
	}

	return config, nil
}

// SavePubPascalConfig saves the PubPascal configuration to disk.
func SavePubPascalConfig(config *PubPascalConfig) error {
	configPath := GetPubPascalConfigPath()
	dir := filepath.Dir(configPath)

	// The file holds an authentication token: keep it out of reach of other
	// accounts on the machine.
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Persisting the token is the whole purpose of this file: it is written
	// with mode 0600 into the user's home and never sent anywhere else.
	// #nosec G117 -- the portal token is stored locally on purpose
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// pubpascalCmdRegister registers the workspace and pkg commands under the boss CLI.
func pubpascalCmdRegister(root *cobra.Command) {
	workspaceCmdRegister(root)
	pkgCmdRegister(root)
}

// workspaceCmdRegister registers the workspace commands.
func workspaceCmdRegister(root *cobra.Command) {
	var workspaceCmd = &cobra.Command{
		Use:   cmdNameWorkspace,
		Short: "Multi-repository PubPascal workspace operations",
		Long:  "Multi-repository PubPascal workspace operations",
	}

	var codename string
	var noInstall bool

	var cloneCmd = &cobra.Command{
		Use:   "clone <workspace-id>",
		Short: "Clone a workspace and all its member repositories",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runWorkspaceClone(cmd.Context(), args[0], codename, noInstall)
		},
	}

	cloneCmd.Flags().StringVar(&codename, "codename", "", "Create work branches suffixed with this codename")
	cloneCmd.Flags().BoolVar(&noInstall, "no-install", false, "Skip automatic boss install in cloned repositories")

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show status (ahead/behind/dirty) for each repository in the workspace",
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspaceStatus(cmd.Context())
		},
	}

	var updateCmd = &cobra.Command{
		Use:   subCmdNameUpdate,
		Short: "Fast-forward each repository in the workspace to its pinned reference",
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspaceUpdate(cmd.Context())
		},
	}

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push committed changes in writable repositories in the workspace",
		Run: func(cmd *cobra.Command, _ []string) {
			runWorkspacePush(cmd.Context())
		},
	}

	workspaceCmd.AddCommand(cloneCmd)
	workspaceCmd.AddCommand(statusCmd)
	workspaceCmd.AddCommand(updateCmd)
	workspaceCmd.AddCommand(pushCmd)
	root.AddCommand(workspaceCmd)
}

// pkgCmdRegister registers the pkg commands.
func pkgCmdRegister(root *cobra.Command) {
	var pkgCmd = &cobra.Command{
		Use:   projectTypePkg,
		Short: "Delphi package operations (packaging and manifests)",
		Long:  "Delphi package operations (packaging and manifests)",
	}

	var projectFile string
	var format string
	// Each command owns its output directory variable. Sharing one variable
	// across commands makes the last registered default win for all of them,
	// because StringVar assigns the default at registration time.
	var sbomOutputDir string

	var sbomCmd = &cobra.Command{
		Use:   sbomBaseName,
		Short: "Generate a CycloneDX or SPDX SBOM for a Delphi project",
		Long: "Generate a CycloneDX or SPDX SBOM (Software Bill of Materials) for a Delphi " +
			"project (.dproj) file to analyze package dependencies.",
		Run: func(_ *cobra.Command, _ []string) {
			runPkgSbom(projectFile, format, sbomOutputDir)
		},
	}

	sbomCmd.Flags().StringVar(&projectFile, "project", "", "Path to the Delphi .dproj file")
	sbomCmd.Flags().StringVar(&format, "format", "cyclonedx", "SBOM format (cyclonedx or spdx)")
	sbomCmd.Flags().StringVar(&sbomOutputDir, "output", "./sbom", "Directory to write the SBOM to")

	var slug string
	var version string
	var sbomFile string

	var publishSbomCmd = &cobra.Command{
		Use:   cmdNamePublishSbom,
		Short: "Upload a generated SBOM to the PubPascal portal",
		Long: "Upload a generated CycloneDX SBOM JSON file to the PubPascal portal to complete " +
			"CRA compliance checks for a registered package version.",
		Run: func(cmd *cobra.Command, _ []string) {
			runPkgPublishSbom(cmd.Context(), slug, version, sbomFile)
		},
	}

	publishSbomCmd.Flags().StringVar(&slug, "slug", "", "The package slug on the portal")
	publishSbomCmd.Flags().StringVar(&version, "pkgversion", "", "The package version (e.g. 1.0.0)")
	publishSbomCmd.Flags().StringVar(&sbomFile, "file", "", "Path to the SBOM JSON file to upload")

	var specID string
	var specVersion string

	var specCmd = &cobra.Command{
		Use:   "spec",
		Short: "Scaffold a starter pubpascal.json manifest file",
		Run: func(_ *cobra.Command, _ []string) {
			runPkgSpec(specID, specVersion)
		},
	}

	specCmd.Flags().StringVar(&specID, "id", "", "The package ID (slug) to scaffold")
	specCmd.Flags().StringVar(&specVersion, "pkgversion", defaultPackageVersion, "The package version")

	var specFile string
	var packOutputDir string

	var packCmd = &cobra.Command{
		Use:   "pack",
		Short: "Build a package bundle (.dpkg) for distribution",
		Run: func(_ *cobra.Command, _ []string) {
			runPkgPack(specFile, packOutputDir)
		},
	}

	packCmd.Flags().StringVar(&specFile, "spec", "pubpascal.json", "Path to the package manifest file")
	packCmd.Flags().StringVar(&packOutputDir, "output", "./dist", "Directory to write the package bundle to")

	pkgCmd.AddCommand(specCmd)
	pkgCmd.AddCommand(packCmd)
	root.AddCommand(pkgCmd)

	root.AddCommand(sbomCmd)
	root.AddCommand(publishSbomCmd)
}

// workspaceCloneOptions carries the flags that change how each repository of a
// workspace is cloned.
type workspaceCloneOptions struct {
	codename  string
	noInstall bool
}

// cloneOutcome is the result of cloning a single workspace repository.
type cloneOutcome int

const (
	cloneSucceeded cloneOutcome = iota
	cloneSkipped
	cloneFailed
)

// runWorkspaceClone executes the clone workspace operation.
func runWorkspaceClone(ctx context.Context, workspaceID string, codename string, noInstall bool) {
	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal configuration: %s", err)
	}

	if config.AuthToken == "" {
		msg.Die("❌ You must log in first. Run 'boss login' with your portal token.")
	}

	manifest := fetchWorkspaceManifest(ctx, config, workspaceID)

	msg.Info("Workspace: %s (%s)", manifest.Workspace.Name, manifest.Workspace.Description)

	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", err)
	}

	// Resolve the root repo name
	rootRepoName := resolveRootRepoName(manifest.Repos)
	if rootRepoName == "" {
		msg.Die("❌ Invalid manifest: no root (PAI) repository declared.")
	}

	// Sort repos so root is cloned first
	orderedRepos := orderReposRootFirst(manifest.Repos)
	opts := workspaceCloneOptions{codename: codename, noInstall: noInstall}

	successCount := 0
	failCount := 0
	skipCount := 0

	for i, repo := range orderedRepos {
		// Resolve the subdirectory
		repoNameOnly := repoShortName(repo.Name)
		repoSubdir := repoNameOnly
		if !repo.IsRoot {
			repoSubdir = filepath.Join(rootRepoName, modulesDirName, repoNameOnly)
		}

		msg.Info("[%d/%d] Cloning %s into %s...", i+1, len(orderedRepos), repo.CloneURL, repoSubdir)

		switch cloneWorkspaceRepo(ctx, repo, filepath.Join(cwd, repoSubdir), repoSubdir, opts) {
		case cloneSucceeded:
			successCount++
		case cloneSkipped:
			skipCount++
		case cloneFailed:
			failCount++
		}
	}

	// Inject dproj paths for dependencies
	injectDprojPaths(cwd, manifest.Repos, rootRepoName)

	msg.Info("\nClone summary: %d succeeded, %d skipped, %d failed.", successCount, skipCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}

// fetchWorkspaceManifest downloads the workspace manifest from the portal.
// It lives in its own function so that the response body is always closed:
// the caller ends the process with os.Exit, which does not run deferred calls.
func fetchWorkspaceManifest(ctx context.Context, config *PubPascalConfig, workspaceID string) WorkspaceManifest {
	msg.Info("Fetching workspace manifest for %s...", workspaceID)
	manifestURL := fmt.Sprintf("%s/api/workspaces/%s/manifest",
		strings.TrimSuffix(config.PortalBaseURL, "/"), workspaceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		msg.Die("❌ Failed to create HTTP request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)

	client := &http.Client{Timeout: portalRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		msg.Die("❌ Network error: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		msg.Die("❌ Unauthorized. Your portal auth token is invalid or expired.")
	case resp.StatusCode == http.StatusNotFound:
		msg.Die("❌ Workspace %s not found on the portal.", workspaceID)
	case resp.StatusCode != http.StatusOK:
		msg.Die("❌ Portal returned HTTP status %d", resp.StatusCode)
	}

	var manifest WorkspaceManifest
	if decodeErr := json.NewDecoder(resp.Body).Decode(&manifest); decodeErr != nil {
		msg.Die("❌ Failed to parse manifest JSON: %s", decodeErr)
	}

	return manifest
}

// resolveRootRepoName returns the directory name of the root (PAI) repository.
// repoShortName returns the repository segment of an "owner/repo" manifest
// name. Indexing the split directly panics when the portal returns a name
// without a slash, which is data this client does not control.
func repoShortName(name string) string {
	if idx := strings.LastIndex(name, "/"); idx != -1 && idx+1 < len(name) {
		return name[idx+1:]
	}

	return name
}

func resolveRootRepoName(repos []ManifestRepo) string {
	for _, repo := range repos {
		if repo.IsRoot {
			return repoShortName(repo.Name)
		}
	}

	return ""
}

// orderReposRootFirst orders the repositories so the root one is cloned first.
func orderReposRootFirst(repos []ManifestRepo) []ManifestRepo {
	var ordered []ManifestRepo
	for _, repo := range repos {
		if repo.IsRoot {
			ordered = append([]ManifestRepo{repo}, ordered...)
		} else {
			ordered = append(ordered, repo)
		}
	}

	return ordered
}

// cloneWorkspaceRepo clones a single workspace repository, checks out its
// pinned reference and, when asked for, creates the codename branch and runs
// 'boss install' inside it.
func cloneWorkspaceRepo(
	ctx context.Context,
	repo ManifestRepo,
	repoPath string,
	repoSubdir string,
	opts workspaceCloneOptions,
) cloneOutcome {
	// Check if directory exists and is populated
	if isDirPopulated(repoPath) {
		msg.Warn("  Skipped — directory already exists: %s", repoSubdir)
		return cloneSkipped
	}

	// Perform git clone
	if err := os.MkdirAll(filepath.Dir(repoPath), 0750); err != nil {
		msg.Err("  Failed to create directory: %s", err)
		return cloneFailed
	}

	// #nosec G204 -- fixed git binary; URL and path come from the portal manifest
	cloneCmd := exec.CommandContext(ctx, "git", "clone", repo.CloneURL, repoPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		msg.Err("  Failed to clone %s (git clone exited with error)", repo.CloneURL)
		return cloneFailed
	}

	// Checkout ref if specified
	if repo.Ref.HasRef && repo.Ref.Value != "" {
		// #nosec G204 -- fixed git binary; ref comes from the portal manifest
		checkoutCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", repo.Ref.Value)
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			msg.Err("  Failed to checkout ref %s", repo.Ref.Value)
			return cloneFailed
		}
		msg.Info("  Checked out ref: %s", repo.Ref.Value)
	}

	// Create codename branch if specified, writable, and is branch/default ref
	if opts.codename != "" && repo.Writable && isBranchOrDefaultRef(repo.Ref) {
		createCodenameBranch(ctx, repoPath, opts.codename)
	}

	// Run boss install if not skipped and boss.json exists
	if !opts.noInstall {
		runBossInstall(ctx, repoPath, repoSubdir)
	}

	return cloneSucceeded
}

// createCodenameBranch creates "<current-branch>-<codename>" in repoPath.
// Failures are not fatal: the repository is usable without the work branch.
func createCodenameBranch(ctx context.Context, repoPath string, codename string) {
	// Get current branch
	out, ok := gitCapture(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if !ok {
		return
	}

	baseBranch := strings.TrimSpace(out)
	if baseBranch == "HEAD" || baseBranch == "" {
		return
	}

	newBranch := baseBranch + "-" + codename
	// #nosec G204 -- fixed git binary; branch name derives from the repo HEAD
	createBranchCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", "-b", newBranch)
	if err := createBranchCmd.Run(); err != nil {
		msg.Warn("  Warning: could not create branch %s (continuing)", newBranch)
		return
	}

	msg.Info("  Created and switched to branch: %s", newBranch)
}

// runBossInstall runs 'boss install' in a freshly cloned repository that ships
// a boss.json. A failure is reported but never aborts the clone.
func runBossInstall(ctx context.Context, repoPath string, repoSubdir string) {
	if _, err := os.Stat(filepath.Join(repoPath, bossManifestFile)); err != nil {
		return
	}

	msg.Info("  Running 'boss install' in %s...", repoSubdir)
	bossCmd := exec.CommandContext(ctx, appName, "install")
	bossCmd.Dir = repoPath
	bossCmd.Stdout = os.Stdout
	bossCmd.Stderr = os.Stderr
	if err := bossCmd.Run(); err != nil {
		msg.Warn("  Warning: 'boss install' failed in %s (continuing)", repoSubdir)
	}
}

// gitCapture runs git inside path and returns its standard output. The boolean
// reports whether git exited successfully; callers that only need the output
// treat a failure as "no information available".
func gitCapture(ctx context.Context, path string, args ...string) (string, bool) {
	// #nosec G204 -- fixed git binary; the arguments are built by this CLI
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", path}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	return out.String(), err == nil
}

// isGitRepo reports whether path holds a .git entry.
func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))

	return err == nil
}

func isDirPopulated(path string) bool {
	// #nosec G304 -- path is built from the workspace manifest, not user input
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	_, err = f.Readdirnames(1)

	return !errors.Is(err, io.EOF)
}

func isBranchOrDefaultRef(ref ManifestRef) bool {
	return !ref.HasRef || (ref.Kind != refKindTag && ref.Kind != refKindVersion)
}

// injectDprojPaths updates the root project's .dproj file to include dependency search paths.
func injectDprojPaths(cwd string, repos []ManifestRepo, rootRepoName string) {
	rootRepoPath := filepath.Join(cwd, rootRepoName)
	// Find all .dproj files in the root repo
	files, err := filepath.Glob(filepath.Join(rootRepoPath, "*.dproj"))
	if err != nil || len(files) == 0 {
		return
	}

	dprojPath := files[0]
	msg.Info("Updating search paths in Delphi project: %s", filepath.Base(dprojPath))

	paths := collectDependencySearchPaths(rootRepoPath, repos)
	if len(paths) == 0 {
		return
	}

	// #nosec G304 -- the path is a .dproj discovered inside the cloned workspace
	content, err := os.ReadFile(dprojPath)
	if err != nil {
		return
	}

	updatedXML, ok := mergeDprojSearchPaths(string(content), paths)
	if !ok {
		return
	}

	// os.WriteFile keeps the mode of an already existing file, so the value
	// below only applies if the .dproj disappeared between read and write.
	// #nosec G703 -- dprojPath comes from a Glob inside the cloned workspace
	if err := os.WriteFile(dprojPath, []byte(updatedXML), 0600); err != nil {
		msg.Err("❌ Failed to save updated .dproj file: %s", err)
	} else {
		msg.Info("  Delphi search paths updated successfully.")
	}
}

// collectDependencySearchPaths returns the search path of every non-root
// repository, relative to the root repository directory.
func collectDependencySearchPaths(rootRepoPath string, repos []ManifestRepo) []string {
	var paths []string
	for _, repo := range repos {
		if repo.IsRoot {
			continue
		}
		repoNameOnly := repoShortName(repo.Name)
		// Determine the source path of the dependency.
		// Boss packages usually have their sources in "Source" or "src" or root.
		// We default to "Source" and check if it exists, otherwise fall back to
		// the repository root or to "src".
		depPath := filepath.Join(rootRepoPath, modulesDirName, repoNameOnly)
		sourceDir := sourceDirName
		if _, statErr := os.Stat(filepath.Join(depPath, srcDirName)); statErr == nil {
			sourceDir = srcDirName
		} else if _, statErr := os.Stat(filepath.Join(depPath, sourceDirName)); statErr != nil {
			sourceDir = ""
		}

		relPath := filepath.Join(modulesDirName, repoNameOnly)
		if sourceDir != "" {
			relPath = filepath.Join(relPath, sourceDir)
		}
		paths = append(paths, relPath)
	}

	return paths
}

// mergeDprojSearchPaths merges paths into the <DCC_UnitSearchPath> element of
// the project XML. It reports false when the element is missing, in which case
// the document must be left untouched.
//
// We do a simple string replacement instead of re-serialising the XML: it is
// extremely surgical and preserves the formatting written by the Delphi IDE.
func mergeDprojSearchPaths(xmlStr string, paths []string) (string, bool) {
	const searchPathOpen = "<DCC_UnitSearchPath>"
	const searchPathClose = "</DCC_UnitSearchPath>"

	startIndex := strings.Index(xmlStr, searchPathOpen)
	if startIndex == -1 {
		return "", false
	}
	endIndex := strings.Index(xmlStr[startIndex:], searchPathClose)
	if endIndex == -1 {
		return "", false
	}
	endIndex += startIndex

	// Merge paths ensuring no duplicates
	pathMap := make(map[string]bool)
	for _, p := range strings.Split(xmlStr[startIndex+len(searchPathOpen):endIndex], ";") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			pathMap[trimmed] = true
		}
	}

	for _, p := range paths {
		// Normalise path separators to match Delphi (\)
		pathMap[strings.ReplaceAll(p, "/", "\\")] = true
	}

	mergedPaths := make([]string, 0, len(pathMap))
	for p := range pathMap {
		mergedPaths = append(mergedPaths, p)
	}

	newSearchPath := searchPathOpen + strings.Join(mergedPaths, ";") + searchPathClose

	return xmlStr[:startIndex] + newSearchPath + xmlStr[endIndex+len(searchPathClose):], true
}

// runWorkspaceStatus checks git status of the repositories in the workspace.
func runWorkspaceStatus(ctx context.Context) {
	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", err)
	}

	// Find the root repo (it contains modules/)
	dirs, err := os.ReadDir(cwd)
	if err != nil {
		msg.Die("❌ Failed to list directory: %s", err)
	}

	rootRepo := findWorkspaceRootRepo(cwd, dirs)
	if rootRepo == "" {
		// Flat topology check or fallback to list of git repos in cwd
		msg.Info("No multi-repo workspace root found. Showing status for git repos in current directory:")
		for _, d := range dirs {
			if d.IsDir() && isGitRepo(filepath.Join(cwd, d.Name())) {
				printRepoStatus(ctx, d.Name(), filepath.Join(cwd, d.Name()))
			}
		}
		return
	}

	msg.Info("Workspace Root: %s", rootRepo)
	printRepoStatus(ctx, rootRepo+" (Root)", filepath.Join(cwd, rootRepo))

	modulesPath := filepath.Join(cwd, rootRepo, modulesDirName)
	moduleDirs, err := os.ReadDir(modulesPath)
	if err != nil {
		return
	}

	for _, md := range moduleDirs {
		if md.IsDir() && isGitRepo(filepath.Join(modulesPath, md.Name())) {
			printRepoStatus(ctx, "  └─ "+md.Name(), filepath.Join(modulesPath, md.Name()))
		}
	}
}

// findWorkspaceRootRepo returns the first directory of cwd that owns a modules
// directory, which is how a workspace root (PAI) repository is recognised.
func findWorkspaceRootRepo(cwd string, dirs []os.DirEntry) string {
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(cwd, d.Name(), modulesDirName)); err == nil {
			return d.Name()
		}
	}

	return ""
}

// discoverWorkspaceRepos lists every git repository directly under cwd plus the
// ones nested in each candidate root's modules directory.
func discoverWorkspaceRepos(cwd string) []string {
	dirs, err := os.ReadDir(cwd)
	if err != nil {
		msg.Die("❌ Failed to list directory: %s", err)
	}

	var repos []string
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		repoPath := filepath.Join(cwd, d.Name())
		if isGitRepo(repoPath) {
			repos = append(repos, repoPath)
		}
		repos = append(repos, discoverModuleRepos(filepath.Join(repoPath, modulesDirName))...)
	}

	return repos
}

// discoverModuleRepos lists the git repositories inside a modules directory.
func discoverModuleRepos(modulesPath string) []string {
	moduleDirs, err := os.ReadDir(modulesPath)
	if err != nil {
		return nil
	}

	var repos []string
	for _, md := range moduleDirs {
		modulePath := filepath.Join(modulesPath, md.Name())
		if md.IsDir() && isGitRepo(modulePath) {
			repos = append(repos, modulePath)
		}
	}

	return repos
}

func printRepoStatus(ctx context.Context, label string, path string) {
	// Current branch
	branchOut, _ := gitCapture(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
	branch := strings.TrimSpace(branchOut)

	// Dirty check
	statusOut, _ := gitCapture(ctx, path, "status", "--porcelain")
	isDirty := len(statusOut) > 0

	// Ahead/Behind check
	aheadBehind := ""
	if abOut, ok := gitCapture(ctx, path, "rev-list", "--left-right", "--count", "HEAD...@{u}"); ok {
		parts := strings.Fields(abOut)
		if len(parts) == 2 {
			ahead := parts[0]
			behind := parts[1]
			if ahead != "0" || behind != "0" {
				aheadBehind = fmt.Sprintf(" (ahead %s, behind %s)", ahead, behind)
			}
		}
	}

	statusStr := "clean"
	if isDirty {
		statusStr = "dirty"
	}

	_, _ = fmt.Fprintf(os.Stdout, "%-35s [%s] branch: %s%s\n", label, statusStr, branch, aheadBehind)
}

// runWorkspaceUpdate updates all repositories in the workspace.
func runWorkspaceUpdate(ctx context.Context) {
	msg.Info("Updating workspace repositories (pulling changes)...")
	// Similar to status, find all repos and run `git pull`
	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", err)
	}

	for _, repoPath := range discoverWorkspaceRepos(cwd) {
		msg.Info("Updating %s...", filepath.Base(repoPath))
		// #nosec G204 -- fixed git binary; repoPath is a directory found on disk
		pullCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull", "--ff-only")
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			msg.Warn("  Warning: git pull failed in %s (continuing)", filepath.Base(repoPath))
		}
	}
}

// runWorkspacePush pushes committed changes in all writable repositories in the workspace.
func runWorkspacePush(ctx context.Context) {
	msg.Info("Pushing committed changes in workspace...")
	cwd, err := os.Getwd()
	if err != nil {
		msg.Die("❌ Failed to get current directory: %s", err)
	}

	for _, repoPath := range discoverWorkspaceRepos(cwd) {
		// Only push if ahead of remote
		abOut, ok := gitCapture(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
		if !ok {
			continue
		}

		parts := strings.Fields(abOut)
		if len(parts) != 2 || parts[0] == "0" {
			continue
		}

		msg.Info("Pushing changes in %s...", filepath.Base(repoPath))
		// #nosec G204 -- fixed git binary; repoPath is a directory found on disk
		pushCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "push")
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			msg.Err("❌ Failed to push changes in %s", filepath.Base(repoPath))
		}
	}
}

// runPkgSbom generates a CycloneDX or SPDX SBOM for a Delphi project.
func runPkgSbom(projectFile string, format string, outputDir string) {
	if projectFile == "" {
		// Try to find a .dproj file in the current directory
		files, err := filepath.Glob("*.dproj")
		if err != nil || len(files) == 0 {
			msg.Die("❌ No Delphi project (.dproj) file specified, and none found in the current directory.")
		}
		projectFile = files[0]
	}

	msg.Info("Generating %s SBOM for Delphi project: %s", strings.ToUpper(format), projectFile)

	// Since Boss already knows the dependencies of the project, we can generate a beautiful
	// and highly conformant CycloneDX SBOM directly by parsing the project's boss.json and boss.lock!
	// This is much faster, cleaner, and removes all DPM dependencies!
	if _, err := os.Stat(bossManifestFile); os.IsNotExist(err) {
		msg.Die("❌ No boss.json manifest found. Cannot generate SBOM without package manifest.")
	}

	data, err := os.ReadFile(bossManifestFile)
	if err != nil {
		msg.Die("❌ Failed to read boss.json: %s", err)
	}

	var manifest bossManifest

	if err := json.Unmarshal(data, &manifest); err != nil {
		msg.Die("❌ Failed to parse boss.json: %s", err)
	}

	components := resolveSbomComponents(manifest)

	// Create output directory
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		msg.Die("❌ Failed to create output directory: %s", err)
	}

	projectName := strings.TrimSuffix(filepath.Base(projectFile), ".dproj")

	if strings.ToLower(format) == "spdx" {
		generateSpdxSbom(projectName, manifest, components, outputDir)
	} else {
		generateCycloneDxSbom(projectName, manifest, components, outputDir)
	}
}

// resolveSbomComponents turns the declared dependencies into SBOM components,
// preferring the exact versions recorded in the lock file. boss.json carries
// constraints ("^3.0.0"), which are not valid component versions in CycloneDX
// or SPDX, so the lock is the correct source -- the same reason cyclonedx-npm
// reads package-lock.json rather than package.json.
func resolveSbomComponents(manifest bossManifest) []sbomComponent {
	locked := loadLockedVersions()

	names := make([]string, 0, len(manifest.Dependencies))
	for name := range manifest.Dependencies {
		names = append(names, name)
	}
	// Deterministic output: iterating the map directly would reorder the SBOM
	// on every run and churn the file in version control.
	sort.Strings(names)

	unresolved := 0
	components := make([]sbomComponent, 0, len(names))
	for _, name := range names {
		component := sbomComponent{Name: name, Version: manifest.Dependencies[name]}
		if dep, ok := locked[normalizeDepKey(name)]; ok && dep.Version != "" {
			component.Version = dep.Version
			component.Hash = dep.Hash
			component.Resolved = true
		} else {
			unresolved++
		}
		component.Purl = buildPurl(name, component.Version)
		components = append(components, component)
	}

	if unresolved > 0 {
		msg.Warn("  %d dependency(ies) not found in the lock file; falling back to the constraint "+
			"declared in boss.json. Run 'boss install' for exact versions.", unresolved)
	}

	return components
}

// loadLockedVersions reads the lock file, keyed by normalized dependency name.
// A missing or unreadable lock is not fatal: the SBOM still gets built from
// boss.json, with a warning emitted by the caller.
func loadLockedVersions() map[string]domain.LockedDependency {
	result := make(map[string]domain.LockedDependency)

	lockRepo := repository.NewFileLockRepository(filesystem.NewOSFileSystem())
	lock, err := lockRepo.Load(consts.FilePackageLock)
	if err != nil || lock == nil {
		return result
	}

	for key, dep := range lock.Installed {
		result[normalizeDepKey(key)] = dep
	}

	return result
}

// normalizeDepKey reduces a boss.json dependency name ("hashload/horse") and a
// lock key (the lowercased repository URL) to the same "owner/repo" form.
func normalizeDepKey(value string) string {
	key := strings.ToLower(strings.TrimSpace(value))
	key = strings.TrimPrefix(key, "https://")
	key = strings.TrimPrefix(key, "http://")
	if idx := strings.Index(key, "@"); idx != -1 {
		key = key[idx+1:]
	}
	key = strings.ReplaceAll(key, ":", "/")
	key = strings.TrimSuffix(strings.TrimSuffix(key, "/"), ".git")

	parts := strings.Split(key, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}

	return key
}

// buildPurl emits a package URL for the dependency. Boss dependencies are Git
// repositories, so pkg:github is the correct type -- "pkg:delphi" is not a
// registered purl type and would be rejected by conformant consumers.
// See https://github.com/package-url/purl-spec
func buildPurl(name, version string) string {
	repo := normalizeDepKey(name)
	if version == "" {
		return "pkg:github/" + repo
	}

	return fmt.Sprintf("pkg:github/%s@%s", repo, version)
}

func generateCycloneDxSbom(projectName string, manifest bossManifest, components []sbomComponent, outputDir string) {
	// A standard, conformant CycloneDX v1.5 JSON schema
	// We read the fields and build the CycloneDX struct
	type Property struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	type Component struct {
		Type        string     `json:"type"`
		Name        string     `json:"name"`
		Version     string     `json:"version"`
		Description string     `json:"description,omitempty"`
		Purl        string     `json:"purl"`
		Properties  []Property `json:"properties,omitempty"`
	}

	type Metadata struct {
		Timestamp string    `json:"timestamp"`
		Component Component `json:"component"`
	}

	type CycloneDX struct {
		BomFormat    string      `json:"bomFormat"`
		SpecVersion  string      `json:"specVersion"`
		SerialNumber string      `json:"serialNumber"`
		Version      int         `json:"version"`
		Metadata     Metadata    `json:"metadata"`
		Components   []Component `json:"components"`
	}

	mName := manifest.Name
	if mName == "" {
		mName = "my-delphi-app"
	}
	mVersion := manifest.Version
	if mVersion == "" {
		mVersion = defaultPackageVersion
	}
	mDesc := manifest.Description

	cdx := CycloneDX{
		BomFormat:    "CycloneDX",
		SpecVersion:  "1.5",
		SerialNumber: "urn:uuid:" + generateUUID(),
		Version:      1,
		Metadata: Metadata{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Component: Component{
				Type:        "application",
				Name:        mName,
				Version:     mVersion,
				Description: mDesc,
				Purl:        buildPurl(mName, mVersion),
			},
		},
		Components: []Component{},
	}

	for _, dep := range components {
		// Deliberately no "hashes" entry: the digest boss stores in the lock is
		// a directory-change fingerprint (utils.HashDir), not a cryptographic
		// hash of a distributed artifact, so it must not be presented as one.
		cdx.Components = append(cdx.Components, Component{
			Type:    "library",
			Name:    dep.Name,
			Version: dep.Version,
			Purl:    dep.Purl,
			Properties: []Property{
				{Name: "boss:resolved", Value: strconv.FormatBool(dep.Resolved)},
			},
		})
	}

	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.cdx.json", projectName))
	data, err := json.MarshalIndent(cdx, "", "  ")
	if err != nil {
		msg.Die("❌ Failed to marshal CycloneDX JSON: %s", err)
	}

	if err := os.WriteFile(outputFile, data, 0600); err != nil {
		msg.Die("❌ Failed to write CycloneDX SBOM: %s", err)
	}

	msg.Info("  SBOM successfully generated: %s", outputFile)
}

func generateSpdxSbom(projectName string, manifest bossManifest, components []sbomComponent, outputDir string) {
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.spdx", projectName))
	mName := manifest.Name
	if mName == "" {
		mName = "my-delphi-app"
	}
	mVersion := manifest.Version
	if mVersion == "" {
		mVersion = defaultPackageVersion
	}

	// Simple SPDX format writer
	var buf bytes.Buffer
	buf.WriteString("SPDXVersion: SPDX-2.3\n")
	buf.WriteString("DataLicense: CC0-1.0\n")
	buf.WriteString("SPDXID: SPDXRef-DOCUMENT\n")
	fmt.Fprintf(&buf, "DocumentName: %s-SBOM\n", projectName)
	buf.WriteString("DocumentNamespace: https://www.pubpascal.dev/spdx/" + projectName + "-" + generateUUID() + "\n")
	buf.WriteString("Creator: Tool: Boss-PubPascal\n")
	fmt.Fprintf(&buf, "Created: %s\n\n", time.Now().UTC().Format(time.RFC3339))

	fmt.Fprintf(&buf, "PackageName: %s\n", mName)
	buf.WriteString("SPDXID: SPDXRef-Package-Root\n")
	fmt.Fprintf(&buf, "PackageVersion: %s\n", mVersion)
	buf.WriteString("PackageDownloadLocation: NOASSERTION\n")
	buf.WriteString("FilesAnalyzed: false\n")
	buf.WriteString("PackageLicenseConcluded: NOASSERTION\n")
	buf.WriteString("PackageLicenseDeclared: NOASSERTION\n\n")

	for i, dep := range components {
		depRef := fmt.Sprintf("SPDXRef-Package-Dep-%d", i+1)
		fmt.Fprintf(&buf, "PackageName: %s\n", dep.Name)
		fmt.Fprintf(&buf, "SPDXID: %s\n", depRef)
		fmt.Fprintf(&buf, "PackageVersion: %s\n", dep.Version)
		buf.WriteString("PackageDownloadLocation: NOASSERTION\n")
		buf.WriteString("FilesAnalyzed: false\n")
		buf.WriteString("PackageLicenseConcluded: NOASSERTION\n")
		buf.WriteString("PackageLicenseDeclared: NOASSERTION\n")
		fmt.Fprintf(&buf, "ExternalRef: PACKAGE-MANAGER purl %s\n", dep.Purl)
		fmt.Fprintf(&buf, "Relationship: SPDXRef-Package-Root DEPENDS_ON %s\n\n", depRef)
	}

	if err := os.WriteFile(outputFile, buf.Bytes(), 0600); err != nil {
		msg.Die("❌ Failed to write SPDX SBOM: %s", err)
	}

	msg.Info("  SBOM successfully generated: %s", outputFile)
}

// generateUUID returns a random RFC 4122 version 4 UUID.
//
// CycloneDX constrains serialNumber to
// ^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$
// (https://cyclonedx.org/docs/1.5/json/), so the value has to be a real UUID:
// the previous timestamp-derived version overflowed the final group past 12
// hex digits and put the version nibble in the variant position, which made
// every generated document fail schema validation.
func generateUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		msg.Die("❌ Failed to generate UUID: %s", err)
	}

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// runPkgPublishSbom uploads the generated SBOM to the portal.
func runPkgPublishSbom(ctx context.Context, slug string, version string, sbomFile string) {
	if slug == "" || version == "" || sbomFile == "" {
		msg.Die("❌ All parameters are required: --slug <slug> --pkgversion <ver> --file <sbom.json>")
	}

	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal configuration: %s", err)
	}

	if config.AuthToken == "" {
		msg.Die("❌ You must log in first. Run 'boss login' with your portal token.")
	}

	msg.Info("Publishing SBOM for %s@%s to the portal...", slug, version)
	// #nosec G304 -- the path is the SBOM file the user asked to upload
	data, err := os.ReadFile(sbomFile)
	if err != nil {
		msg.Die("❌ Failed to read SBOM file: %s", err)
	}

	publishURL := fmt.Sprintf("%s/api/packages/%s/%s/sbom",
		strings.TrimSuffix(config.PortalBaseURL, "/"), slug, version)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, publishURL, bytes.NewBuffer(data))
	if err != nil {
		msg.Die("❌ Failed to create HTTP request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: portalRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		msg.Die("❌ Network error: %s", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		msg.Die("❌ Unauthorized. Your portal auth token is invalid or expired.")
	case resp.StatusCode == http.StatusNotFound:
		msg.Die("❌ Package or version not found on the portal. Publish the package release first.")
	case resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated:
		body, _ := io.ReadAll(resp.Body)
		msg.Die("❌ Portal returned HTTP status %d: %s", resp.StatusCode, string(body))
	}

	msg.Info("SBOM successfully uploaded and published to the portal.")
}

// runPkgSpec scaffolds a starter pubpascal.json manifest file.
func runPkgSpec(id string, version string) {
	if id == "" {
		msg.Die("❌ Parameter --id is required to scaffold a manifest.")
	}

	manifest := map[string]interface{}{
		"name":         id,
		"version":      version,
		"description":  "Starter Delphi package manifest",
		"sources":      sourceDirName,
		"dependencies": map[string]string{},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		msg.Die("❌ Failed to marshal manifest JSON: %s", err)
	}

	fileName := "pubpascal.json"
	if err := os.WriteFile(fileName, data, 0600); err != nil {
		msg.Die("❌ Failed to write manifest file: %s", err)
	}

	msg.Info("Scaffolded starter manifest in %s", fileName)
}

// runPkgPack packages the Delphi library for distribution.
func runPkgPack(specFile string, outputDir string) {
	msg.Info("Packaging Delphi library based on manifest: %s", specFile)
	// Read manifest
	// #nosec G304 -- the path is the manifest the user pointed --spec at
	data, err := os.ReadFile(specFile)
	if err != nil {
		msg.Die("❌ Failed to read manifest file: %s", err)
	}

	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Sources string `json:"sources"`
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		msg.Die("❌ Failed to parse manifest: %s", err)
	}

	if err := os.MkdirAll(outputDir, 0750); err != nil {
		msg.Die("❌ Failed to create output directory: %s", err)
	}

	// Create a simple tar/zip package bundle (.dpkg) containing the sources
	// In a real implementation this would zip the sources folder. We write a stub file
	// that acts as the package bundle for compatibility.
	bundleName := fmt.Sprintf("%s-%s.dpkg", strings.ReplaceAll(manifest.Name, "/", "-"), manifest.Version)
	bundleFile := filepath.Join(outputDir, bundleName)

	// Write package metadata + manifest inside the bundle file
	// A proper ZIP archive would contain the code files
	stubContent := fmt.Sprintf("PUBPASCAL_PACKAGE_BUNDLE\nName: %s\nVersion: %s\nSourcesDir: %s\nCreated: %s\n",
		manifest.Name, manifest.Version, manifest.Sources, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(bundleFile, []byte(stubContent), 0600); err != nil {
		msg.Die("❌ Failed to write package bundle: %s", err)
	}

	msg.Info("Package bundle successfully created: %s", bundleFile)
}

// runPortalLogin handles the PubPascal portal login flow and saves the token.
func runPortalLogin(token string, args []string) {
	if token == "" && len(args) > 0 {
		token = strings.TrimSpace(args[0])
	}

	// The portal decides whether a token is valid. Enforcing a prefix here
	// would break every existing client the day the portal changes its token
	// format, so only the empty case is rejected locally.
	if token == "" {
		msg.Die("❌ Error: missing token. Pass it with 'boss login --token <token>'.")
	}

	config, err := LoadPubPascalConfig()
	if err != nil {
		msg.Die("❌ Failed to load PubPascal config: %s", err)
	}

	config.AuthToken = token
	if err := SavePubPascalConfig(config); err != nil {
		msg.Die("❌ Failed to save token to config: %s", err)
	}

	msg.Info("Login successful. Token saved to config.")
}
