package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

// contributeCmdRegister registers the contribute command.
func contributeCmdRegister(root *cobra.Command) {
	var prMode bool
	var prTitle string
	var prBody string

	var contributeCmd = &cobra.Command{
		Use:   "contribute <package-slug>",
		Short: "Contribute to a third-party package by automating fork and Pull Request creation",
		Long:  `Contribute to a package. It automatically forks the repository, configures upstream/origin remotes, checkouts a new branch, and opens a Pull Request on GitHub once you are done.`,
		Example: `  Start contributing to a package:
  boss contribute github.com/HashLoad/nidus

  Push changes and create a Pull Request on the upstream repository:
  boss contribute github.com/HashLoad/nidus --pr --title "Fix memory leak" --body "..."`,
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			packageSlug := args[0]
			config, err := LoadPubPascalConfig()
			if err != nil {
				msg.Die("❌ Failed to load PubPascal config: %s", err)
			}

			if config.AuthToken == "" {
				msg.Die("❌ Error: You must login first. Run 'boss login --token <token>'")
			}

			// Resolve local package folder
			dep := domain.ParseDependency(packageSlug, "")
			folderName := dep.Name()
			pkgDir := filepath.Join(env.GetModulesDir(), folderName)

			if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
				msg.Die("❌ Error: Package directory not found at %s. Ensure the package is installed.", pkgDir)
			}

			if _, err := os.Stat(filepath.Join(pkgDir, ".git")); os.IsNotExist(err) {
				msg.Die("❌ Error: Directory %s is not a git repository.", pkgDir)
			}

			if prMode {
				handlePullRequestFlow(packageSlug, pkgDir, config, prTitle, prBody)
			} else {
				handleForkSetupFlow(packageSlug, pkgDir, config)
			}
		},
	}

	contributeCmd.Flags().BoolVar(&prMode, "pr", false, "push changes and open a Pull Request")
	contributeCmd.Flags().StringVar(&prTitle, "title", "", "Pull Request title (defaults to last commit message)")
	contributeCmd.Flags().StringVar(&prBody, "body", "", "Pull Request description (defaults to last commit body)")
	root.AddCommand(contributeCmd)
}

func handleForkSetupFlow(packageSlug string, pkgDir string, config *PubPascalConfig) {
	msg.Info("🍴 Requesting Fork from portal for %s...", packageSlug)

	// Call Portal API to fork
	url := fmt.Sprintf("%s/api/packages/contribute/fork", config.PortalBaseUrl)
	requestBody, _ := json.Marshal(map[string]string{
		"packageSlug": packageSlug,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		msg.Die("❌ Failed to create request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		msg.Die("❌ Connection error: %s", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errRes map[string]string
		_ = json.Unmarshal(bodyBytes, &errRes)
		msg.Die("❌ Fork failed: %s", errRes["error"])
	}

	var res struct {
		Success   bool   `json:"success"`
		CloneUrl  string `json:"clone_url"`
		SshUrl    string `json:"ssh_url"`
		Username  string `json:"github_username"`
		UpstreamO string `json:"upstream_owner"`
		UpstreamR string `json:"upstream_repo"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		msg.Die("❌ Failed to parse API response: %s", err)
	}

	msg.Info("✅ Fork created on GitHub under account: @%s", res.Username)

	// Git Automation: setup remotes
	// 1. Rename origin to upstream (if upstream doesn't exist)
	if !remoteExists(pkgDir, "upstream") {
		msg.Info("⚙️ Renaming remote 'origin' to 'upstream'...")
		if _, err := runGitCmd(pkgDir, "remote", "rename", "origin", "upstream"); err != nil {
			msg.Die("❌ Failed to rename remote: %s", err)
		}
	} else {
		msg.Warn("⚠️ Remote 'upstream' already exists. Skipping remote renaming.")
	}

	// 2. Add fork as origin
	if remoteExists(pkgDir, "origin") {
		msg.Info("⚙️ Removing existing 'origin' remote...")
		if _, err := runGitCmd(pkgDir, "remote", "remove", "origin"); err != nil {
			msg.Die("❌ Failed to remove old origin: %s", err)
		}
	}
	
	// Choose clone URL format (prefer SSH if git config contains git@ or SSH)
	forkUrl := res.CloneUrl
	if auth := env.GlobalConfiguration().Auth[depPrefix(packageSlug)]; auth != nil && auth.UseSSH {
		forkUrl = res.SshUrl
	} else if strings.Contains(forkUrl, "git@") {
		forkUrl = res.SshUrl
	}

	msg.Info("⚙️ Adding Fork URL as 'origin'...")
	if _, err := runGitCmd(pkgDir, "remote", "add", "origin", forkUrl); err != nil {
		msg.Die("❌ Failed to add origin remote: %s", err)
	}

	// 3. Checkout contribution branch
	branchName := generateBranchName()
	msg.Info("⚙️ Creating and checking out branch: %s...", branchName)
	if _, err := runGitCmd(pkgDir, "checkout", "-b", branchName); err != nil {
		msg.Die("❌ Failed to checkout branch: %s", err)
	}

	msg.Info("🚀 Contribution environment successfully configured.")
	msg.Info("👉 You can now make your changes in the IDE, commit them, and run:")
	msg.Info("   boss contribute %s --pr", packageSlug)
}

func handlePullRequestFlow(packageSlug string, pkgDir string, config *PubPascalConfig, title string, body string) {
	// 1. Resolve current branch
	branch, err := runGitCmd(pkgDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		msg.Die("❌ Failed to get current git branch: %s", err)
	}

	if branch == "main" || branch == "master" || branch == "devel" {
		msg.Die("❌ Error: You are on the default branch '%s'. Please checkout your contribution branch first.", branch)
	}

	// 2. Resolve title and body from last commit if not provided
	if title == "" {
		lastCommitTitle, err := runGitCmd(pkgDir, "log", "-1", "--pretty=%s")
		if err == nil && lastCommitTitle != "" {
			title = lastCommitTitle
		} else {
			title = "Contribution from PubPascal Dev-Flow"
		}
	}

	if body == "" {
		lastCommitBody, err := runGitCmd(pkgDir, "log", "-1", "--pretty=%b")
		if err == nil {
			body = lastCommitBody
		}
	}

	msg.Info("🚀 Pushing branch '%s' to your fork (origin)...", branch)
	if _, err := runGitCmd(pkgDir, "push", "origin", branch, "--force"); err != nil {
		msg.Die("❌ Failed to push branch: %s", err)
	}

	msg.Info("📨 Submitting Pull Request to portal...")

	// Call Portal API to create PR
	url := fmt.Sprintf("%s/api/packages/contribute/pr", config.PortalBaseUrl)
	requestBody, _ := json.Marshal(map[string]string{
		"packageSlug": packageSlug,
		"branch":      branch,
		"title":       title,
		"body":        body,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		msg.Die("❌ Failed to create request: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		msg.Die("❌ Connection error: %s", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errRes map[string]string
		_ = json.Unmarshal(bodyBytes, &errRes)
		msg.Die("❌ Pull Request creation failed: %s", errRes["error"])
	}

	var res struct {
		Success bool   `json:"success"`
		PrUrl   string `json:"pr_url"`
		Head    string `json:"head"`
		Base    string `json:"base"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		msg.Die("❌ Failed to parse API response: %s", err)
	}

	msg.Info("🎉 Pull Request successfully created.")
	msg.Info("🔗 Access your PR here: %s", res.PrUrl)
}

// Helper to run git commands
func runGitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Helper to check if a git remote exists
func remoteExists(dir string, remoteName string) bool {
	out, err := runGitCmd(dir, "remote")
	if err != nil {
		return false
	}
	remotes := strings.Split(out, "\n")
	for _, r := range remotes {
		if strings.TrimSpace(r) == remoteName {
			return true
		}
	}
	return false
}

// Helper to get prefix provider
func depPrefix(repo string) string {
	dep := domain.Dependency{Repository: repo}
	return dep.GetURLPrefix()
}

func generateBranchName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("pubpascal/patch-%d", rand.Intn(10000))
}
