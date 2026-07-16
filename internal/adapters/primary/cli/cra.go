package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
)

var securityEmail string

// craCmdRegister registers the cra commands under the boss CLI root
func craCmdRegister(root *cobra.Command) {
	var craCmd = &cobra.Command{
		Use:   "cra",
		Short: "Cyber Resilience Act (CRA) compliance checker and assistant",
		Long: `Diagnose and automate Cyber Resilience Act (CRA) compliance for your Delphi project.
Run without arguments to perform a local compliance check, or use 'cra init' to generate required files.`,
		Run: func(cmd *cobra.Command, _ []string) {
			runCraCheck()
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Start interactive wizard to make your project 100% CRA compliant",
		Long:  "Start the interactive wizard to generate required Cyber Resilience Act (CRA) compliance files, such as SECURITY.md and sbom.cdx.json.",
		Run: func(cmd *cobra.Command, _ []string) {
			runCraInit()
		},
	}

	initCmd.Flags().StringVar(&securityEmail, "email", "", "Security contact email for reporting vulnerabilities")

	craCmd.AddCommand(initCmd)
	root.AddCommand(craCmd)
}

// runCraCheck performs a local diagnostic of the project against CRA/Portal signals
func runCraCheck() {
	msg.Info("🔍 Diagnosing Cyber Resilience Act (CRA) Compliance...\n")

	hasSecurity := false
	securityCandidates := []string{"SECURITY.md", ".github/SECURITY.md", "docs/SECURITY.md"}
	for _, c := range securityCandidates {
		if _, err := os.Stat(c); err == nil {
			hasSecurity = true
			msg.Info("✅ Security Policy: Found security disclosure policy at '%s'", c)
			break
		}
	}
	if !hasSecurity {
		msg.Warn("❌ Security Policy: Missing 'SECURITY.md' in the repository.")
		msg.Info("   -> To fix: Run 'boss cra init' to generate one automatically.")
	}

	hasSbom := false
	sbomCandidates := []string{"sbom.cdx.json", "sbom.spdx.json", "bom.json", "sbom/sbom.cdx.json"}
	for _, c := range sbomCandidates {
		if _, err := os.Stat(c); err == nil {
			hasSbom = true
			msg.Info("✅ SBOM (Software Bill of Materials): Found SBOM file at '%s'", c)
			break
		}
	}
	if !hasSbom {
		msg.Warn("❌ SBOM (Software Bill of Materials): Missing 'sbom.cdx.json'.")
		msg.Info("   -> To fix: Run 'boss sbom' or 'boss cra init' to generate it.")
	}

	// Validate boss.json exists
	if _, err := os.Stat("boss.json"); err != nil {
		msg.Warn("⚠️ boss.json: No boss.json found in current directory.")
	}

	if hasSecurity && hasSbom {
		msg.Info("\n🎉 Your local project is 100% CRA compliant! Commit and push these files to GitHub to get the Gold badge in the portal.")
	} else {
		msg.Info("\n💡 Tips to get 100% CRA badge:")
		msg.Info("1. Run 'boss cra init' to let Boss generate the SECURITY.md and SBOM automatically.")
		msg.Info("2. Commit and push the files to your repository.")
	}
}

// runCraInit runs the interactive wizard to generate compliance files
func runCraInit() {
	msg.Info("🚀 Cyber Resilience Act (CRA) Compliance Wizard\n")

	// 1. Get security email
	email := securityEmail
	if email == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("📧 Enter the email address to report security vulnerabilities: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			msg.Die("❌ Failed to read email: %s", err)
		}
		email = strings.TrimSpace(input)
	}

	if email == "" {
		msg.Die("❌ An email address is required to generate the security policy.")
	}

	// 2. Generate SECURITY.md
	securityContent := fmt.Sprintf(`# Security Policy

## Reporting a Vulnerability

Please report security vulnerabilities directly to the maintainers at the following address:

📧 **%s**

Please do **NOT** open public issues or pull requests for security vulnerabilities until they have been reviewed and a fix is prepared. This keeps the report private and protects users of the package.

We aim to acknowledge reports within a few business days and keep you updated on progress.

## Supported Versions

Security fixes are applied to the latest active release. We recommend always running the latest version of this package.
`, email)

	err := os.WriteFile("SECURITY.md", []byte(securityContent), 0644)
	if err != nil {
		msg.Die("❌ Failed to write SECURITY.md: %s", err)
	}
	msg.Info("✅ Created 'SECURITY.md' with your security contact details.")

	// 3. Generate sbom.cdx.json if boss.json is present
	if _, err := os.Stat("boss.json"); err == nil {
		msg.Info("📦 boss.json detected. Generating Software Bill of Materials (SBOM)...")
		
		// Load boss.json
		data, err := os.ReadFile("boss.json")
		if err == nil {
			var bossManifest map[string]interface{}
			_ = json.Unmarshal(data, &bossManifest)
			
			projectName := "delphi-project"
			if name, ok := bossManifest["name"].(string); ok && name != "" {
				projectName = name
			}
			
			// We reuse the existing CycloneDX generator logic from pubpascal.go
			generateCycloneDxSbom(projectName, bossManifest, ".")
			
			// Move sbom.cdx.json to root if it was written elsewhere or check if it's in the root
			if _, err := os.Stat("sbom.cdx.json"); err == nil {
				msg.Info("✅ Created 'sbom.cdx.json' in the project root.")
			}
		}
	} else {
		msg.Warn("⚠️ No boss.json found. Skipping SBOM generation. Create a boss.json first.")
	}

	msg.Info("\n🎉 Compliance files generated successfully!")
	msg.Info("To get the 100%% CRA-ready badge in the PubPascal portal:")
	msg.Info("1. Commit the newly created 'SECURITY.md' and 'sbom.cdx.json' to Git.")
	msg.Info("2. Push them to your GitHub repository.")
}
