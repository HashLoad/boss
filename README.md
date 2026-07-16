# Boss

![Boss][bossLogo]

[![Go Report Card][goReportBadge]][goReportLink]
[![GitHub release (latest by date)][latestReleaseBadge]](https://github.com/HashLoad/boss/releases/latest)
[![CRA Compliance][craBadge]][securityPolicyLink]
[![SBOM Badge][sbomBadge]](https://www.pubpascal.dev)
[![GitHub Release Date][releaseDateBadge]](https://github.com/HashLoad/boss/releases)
[![GitHub repo size][repoSizeBadge]](https://github.com/HashLoad/boss/archive/refs/heads/main.zip)
[![GitHub All Releases][totalDownloadsBadge]](https://github.com/HashLoad/boss/releases)
[![GitHub][githubLicenseBadge]](https://github.com/HashLoad/boss/blob/main/LICENSE)
[![GitHub issues][githubIssuesBadge]](https://github.com/HashLoad/boss/issues)
[![GitHub pull requests][githubPullRequestsBadge]](https://github.com/HashLoad/boss/pulls)
[![Ask DeepWiki][deepWikiBadge]](https://deepwiki.com/HashLoad/boss)
[![GitHub contributors][githubContributorsBadge]](https://github.com/HashLoad/boss?tab=readme-ov-file#-code-contributors)
![Github Stars][repoStarsBadge]

_Boss_ is an open source dependency manager inspired by [npm](https://www.npmjs.com/) for projects developed in _Delphi_ and _Lazarus_.

[![Boss][telegramBadge]][telegramLink]

<!-- getting start with emoji -->

## 🚀 Getting started

We have a [Getting Started](https://medium.com/@matheusarendthunsche/come%C3%A7ando-com-o-boss-72aad9bcc13) article to help you get started with Boss.

## 📦 Installation

- Download [setup](https://github.com/hashload/boss/releases)
- Just type `boss` in the terminal
- (Optional) Install a [Boss Delphi IDE complement](https://github.com/hashload/boss-ide)

Or you can use the following the steps below:

- Download the latest version of the [Boss](https://github.com/hashload/boss/releases)
- Extract the files to a folder
- Add the folder to the system path
- Run the command `boss` in the terminal

## 📚 Available Commands

This section documents all commands supported by the Boss CLI, grouped in the exact order they appear in `boss --help`.

---

### 1. Available Commands (Legacy Core)

These are the classic dependency management commands inherited from the original Boss engine.

#### > config
Manage configuration settings for Boss (e.g., Delphi paths, Git client settings, etc.):
```sh
# Set native Git client (recommended on Windows)
boss config git mode native

# Enable shallow cloning for faster dependency checkout
boss config git shallow true
```

#### > dependencies
List all project dependencies in a tree format. Add `-v` to show version information:
```sh
boss dependencies
boss dependencies -v
boss dependencies <package>
```
> Aliases: `dep`, `ls`, `list`, `ll`, `la`, `dependency`

#### > init
Initialize a new, minimal project configuration in the current directory and create a `boss.json` file:
```sh
boss init
boss init --quiet  # Skip interactive prompts
```

#### > install
Install one or more dependencies defined in the `boss.json` file or add a new dependency:
```sh
# Install dependencies from boss.json
boss install

# Add and install a new dependency
boss install github.com/HashLoad/horse
```
> Aliases: `i`, `add`

#### > logout
Remove saved credentials for a private repository or registry:
```sh
boss logout github.com/username
```

#### > uninstall
Remove a dependency from the current project:
```sh
boss uninstall <dependency>
```
> Aliases: `remove`, `rm`, `r`, `un`, `unlink`

#### > update
Update all installed dependencies to their latest compatible versions:
```sh
boss update
```
> Aliases: `up`

#### > upgrade
Upgrade the Boss CLI client to the latest version:
```sh
boss upgrade
boss upgrade --dev  # Upgrade to the latest pre-release
```

#### > version
Show the Boss CLI version:
```sh
boss version
boss --version
```
> Aliases: `v`

---

### 2. Available Commands (new)

These commands add modern project creation, compiling, and script running capabilities to Boss.

#### > new
Generate a fully structured Delphi or Lazarus project template (skeleton) in the current directory:
```sh
boss new delphi
boss new lazarus
```

#### > pkg
Perform Delphi package operations including packaging, signing, and verification.
* **`pkg spec`**: Scaffolds a starter `pubpascal.json` manifest file for the package:
  ```sh
  boss pkg spec --id my-package --pkgversion 1.0.0
  ```
* **`pkg pack`**: Build a redistributable package bundle (`.dpkg`):
  ```sh
  boss pkg pack --spec pubpascal.json --output ./dist
  ```
* **`pkg sign`**: Statically sign a package bundle using a PFX certificate:
  ```sh
  boss pkg sign --package mypkg.dpkg --pfx cert.pfx --pfx-password-env CERT_PASSWORD
  ```
* **`pkg verify`**: Verify a package bundle's signature and integrity:
  ```sh
  boss pkg verify --package mypkg.dpkg
  ```

#### > run
Execute a custom shell script defined in the `scripts` section of your `boss.json` file:
```json
{
  "scripts": {
    "build": "msbuild MyProject.dproj /p:Config=Release",
    "clean": "del /s *.dcu"
  }
}
```
```sh
boss run build
boss run clean
```

---

### 3. Available Commands (pubpascal)

These commands integrate your local development workflow with the PubPascal Portal.

#### > login
Authenticate your local environment with the PubPascal portal using a Personal Access Token (PAT):
```sh
# Authenticate using a personal access token
boss login --token <pat>

# Or start interactive login (prompts for the token)
boss login portal
```
#### > contribute
Contribute to a third-party package by automating repository forking and Pull Request creation.
* **Fork & Setup**: Automatically forks the upstream package and configures your local git remotes (`origin` for your fork and `upstream` for the original):
  ```sh
  boss contribute github.com/HashLoad/horse
  ```
* **Submit Pull Request**: Once your commits are ready, push changes and submit a Pull Request to the original repository with one command:
  ```sh
  boss contribute github.com/HashLoad/horse --pr --title "Fix memory leak" --body "..."
  ```

#### > workspace
Manage multi-repository PubPascal workspaces locally.
* **`workspace clone`**: Clones a workspace and all its member repositories, setting up writable forks:
  ```sh
  boss workspace clone <workspace-id>
  boss workspace clone <workspace-id> --codename my-branch
  ```
* **`workspace status`**: Show Git status (ahead/behind/dirty) for all repositories in the workspace:
  ```sh
  boss workspace status
  ```
* **`workspace update`**: Fast-forward all repositories to their pinned reference branch/commit:
  ```sh
  boss workspace update
  ```
* **`workspace push`**: Push committed changes across all writable repositories in the workspace:
  ```sh
  boss workspace push
  ```

---

### 4. Cyber Resilience Act (CRA) & SBOM

These native commands help you achieve 100% Cyber Resilience Act (CRA) compliance.

#### > cra
Check your project's CRA compliance status or initialize required files automatically.
* **`cra` (Diagnose)**: Scan the local project for required CRA signals (Security Policy, SBOM):
  ```sh
  boss cra
  ```
* **`cra init` (Wizard)**: Start the interactive wizard to generate the `SECURITY.md` policy and `sbom.cdx.json` SBOM:
  ```sh
  boss cra init
  boss cra init --email security@yourcompany.com  # Silent/CI mode
  ```

#### > sbom
Generate a standard CycloneDX or SPDX Software Bill of Materials (SBOM) for your Delphi project:
```sh
# Generate CycloneDX SBOM (outputs to ./sbom/sbom.cdx.json)
boss sbom

# Specify custom project file and output path
boss sbom --project ./src/MyProj.dproj --output ./custom-sbom-folder

# Generate in SPDX format
boss sbom --format spdx
```

#### > scan
Scan a generated SBOM against the OSV.dev database for known vulnerabilities:
```sh
boss scan
boss scan --sbom ./sbom/sbom.cdx.json
```

#### > publish-sbom
Upload a generated SBOM to the PubPascal portal for remote compliance badge rendering:
```sh
boss publish-sbom --slug my-pkg-slug --pkgversion 1.0.0 --file ./sbom/sbom.cdx.json
```

---

### 5. Additional Commands

#### > cache
Manage the Boss local cache to clear downloaded modules and free up disk space:
```sh
boss config cache rm
```
> Aliases: `purge`, `clean`

#### > completion
Generate the autocompletion script for the specified shell (bash, zsh, fish, or powershell):
```sh
boss completion powershell | Out-String | Invoke-Expression
```

---

## Global Flags

* **`-g, --global`**: Use global environment for installation (packages are available system-wide):
  ```sh
  boss install -g <dependency>
  ```
* **`-d, --debug`**: Enable debug mode to see detailed output:
  ```sh
  boss install -d
  ```
* **`-h, --help`**: Show help for any command:
  ```sh
  boss --help
  boss install --help
  ```
* **`-v, --version`**: Show CLI client version:
  ```sh
  boss --version
  ```

## Configuration

### > Delphi Version
Configure which Delphi version Boss should use for compiling packages:
```sh
# List all detected Delphi installations
boss config delphi list

# Select a Delphi version to use globally
boss config delphi use <index>
boss config delphi use 37.0
boss config delphi use 37.0-Win64
```



## Samples

```sh
boss install horse
boss install horse:1.0.0
boss install -g delphi-docker
boss install -g boss-ide
```

## Using [semantic versioning](https://semver.org/) to specify update types your package can accept

You can specify which update types your package can accept from dependencies in your package's boss.json file.

For example, to specify acceptable version ranges up to 1.0.4, use the following syntax:

- Patch releases: 1.0 or 1.0.x or ~1.0.4
- Minor releases: 1 or 1.x or ^1.0.4
- Major releases: \* or x

## boss.json File Format

The `boss.json` file is the manifest for your Delphi/Lazarus project. It contains metadata, dependencies, build configuration, and custom scripts.

### Complete Structure

Here's a comprehensive example showing all available fields:

```json
{
  "name": "my-project",
  "description": "A sample Delphi project using Boss",
  "version": "1.0.0",
  "homepage": "https://github.com/myuser/my-project",
  "mainsrc": "src/",
  "browsingpath": "src/;libs/",
  "projects": [
    "MyProject.dproj",
    "MyPackage.dproj"
  ],
  "dependencies": {
    "github.com/HashLoad/horse": "^3.0.0",
    "github.com/HashLoad/jhonson": "~2.1.0",
    "dataset-serialize": "*"
  },
  "scripts": {
    "build": "msbuild MyProject.dproj /p:Config=Release",
    "test": "MyProject.exe --test",
    "clean": "del /s *.dcu"
  },
  "engines": {
    "compiler": ">=35.0",
    "platforms": ["Win32", "Win64"]
  },
  "toolchain": {
    "compiler": "37.0",
    "platform": "Win64",
    "path": "C:\\Program Files\\Embarcadero\\Studio\\37.0",
    "strict": false
  }
}
```

### Field Descriptions

#### Core Fields

- **`name`** (required): Package name. Must be unique if publishing.
  ```json
  "name": "my-awesome-library"
  ```

- **`description`** (optional): A brief description of your project.
  ```json
  "description": "REST API framework for Delphi"
  ```

- **`version`** (required): Package version following [semantic versioning](https://semver.org/).
  ```json
  "version": "1.2.3"
  ```

- **`homepage`** (optional): Project website or repository URL.
  ```json
  "homepage": "https://github.com/myuser/my-project"
  ```

#### Source Configuration

- **`mainsrc`** (optional): Main source directory path.
  ```json
  "mainsrc": "src/"
  ```

- **`browsingpath`** (optional): Additional paths for IDE browsing (semicolon-separated).
  ```json
  "browsingpath": "src/;src/controllers/;src/models/"
  ```

#### Build Configuration

- **`projects`** (optional): List of Delphi project files (`.dproj`) to compile.
  ```json
  "projects": [
    "MyProject.dproj",
    "MyLibrary.dproj"
  ]
  ```

  **Note:** If not specified, Boss won't compile the package but will still manage dependencies.

#### Dependencies

- **`dependencies`** (optional): Map of package dependencies with version constraints.
  ```json
  "dependencies": {
    "github.com/HashLoad/horse": "^3.0.0",
    "dataset-serialize": "~2.1.0",
    "jhonson": "*"
  }
  ```

  Supported version formats:
  - Exact version: `"1.0.0"`
  - Caret (minor updates): `"^1.0.0"` (allows 1.x.x, but not 2.x.x)
  - Tilde (patch updates): `"~1.0.0"` (allows 1.0.x, but not 1.1.x)
  - Wildcard (any): `"*"` or `"x"`
  - Range: `">=1.0.0 <2.0.0"`

#### Custom Scripts

- **`scripts`** (optional): Custom commands you can run with `boss run <script-name>`.
  ```json
  "scripts": {
    "build": "msbuild MyProject.dproj /p:Config=Release",
    "test": "dunitx-console.exe MyProject.exe",
    "clean": "del /s *.dcu *.exe",
    "deploy": "xcopy /s /y bin\\*.exe deploy\\"
  }
  ```

  Execute with:
  ```sh
  boss run build
  boss run test
  ```

#### Engine Requirements

- **`engines`** (optional): Specify minimum compiler/platform requirements.
  ```json
  "engines": {
    "compiler": ">=35.0",
    "platforms": ["Win32", "Win64", "Linux64"]
  }
  ```

  - `compiler`: Minimum compiler version
  - `platforms`: Supported target platforms

#### Toolchain Configuration

- **`toolchain`** (optional): Specify the exact toolchain to use for this project.
  ```json
  "toolchain": {
    "compiler": "37.0",
    "platform": "Win64",
    "path": "C:\\Program Files\\Embarcadero\\Studio\\37.0",
    "strict": true
  }
  ```

  - `compiler`: Required compiler version
  - `platform`: Target platform ("Win32", "Win64", "Linux64", etc.)
  - `path`: Explicit path to the compiler (optional)
  - `strict`: If `true`, fails if the exact version is not found (default: `false`)

### Minimal boss.json (Classic Format)

A basic, classic `boss.json` showing that Boss remains fully backwards-compatible and works out of the box with just dependency definitions:

```json
{
  "name": "my-project",
  "version": "1.0.0",
  "dependencies": {
    "github.com/HashLoad/horse": "^3.0.0"
  }
}
```

### Creating a new boss.json

Use `boss init` to create a new `boss.json` interactively:

```sh
boss init
```

Or use quiet mode for defaults:

```sh
boss init -q
```

### Example: Library Package

```json
{
  "name": "my-delphi-library",
  "description": "Utilities for Delphi applications",
  "version": "2.1.0",
  "homepage": "https://github.com/myuser/my-library",
  "mainsrc": "src/",
  "projects": [
    "MyLibrary.dproj"
  ],
  "dependencies": {
    "github.com/HashLoad/horse": "^3.0.0"
  }
}
```

### Example: Application Package

```json
{
  "name": "my-app",
  "description": "My awesome Delphi application",
  "version": "1.0.0",
  "projects": [
    "MyApp.dproj"
  ],
  "dependencies": {
    "github.com/HashLoad/horse": "^3.0.0"
  },
  "scripts": {
    "build": "msbuild MyApp.dproj /p:Config=Release",
    "run": "bin\\MyApp.exe",
    "test": "dunitx-console.exe bin\\MyAppTests.exe"
  },
  "toolchain": {
    "compiler": "37.0",
    "platform": "Win32"
  }
}
```


## 💻 Code Contributors

![GitHub Contributors Image](https://contrib.rocks/image?repo=Hashload/boss)

[githubContributorsBadge]: https://img.shields.io/github/contributors/hashload/boss
[ciBadge]: https://github.com/hashload/boss/actions/workflows/ci.yml/badge.svg
[ciLink]: https://github.com/hashload/boss/actions/workflows/ci.yml
[codecovBadge]: https://codecov.io/gh/hashload/boss/branch/main/graph/badge.svg
[codecovLink]: https://codecov.io/gh/hashload/boss
[goReportBadge]: https://goreportcard.com/badge/github.com/hashload/boss
[goReportLink]: https://goreportcard.com/report/github.com/hashload/boss
[bossLogo]: ./assets/png/sized/boss-logo-128px.png
[latestReleaseBadge]: https://img.shields.io/github/v/release/hashload/boss
[releaseDateBadge]: https://img.shields.io/github/release-date/hashload/boss
[repoSizeBadge]: https://img.shields.io/github/repo-size/hashload/boss
[totalDownloadsBadge]: https://img.shields.io/github/downloads/hashload/boss/total
[githubLicenseBadge]: https://img.shields.io/github/license/hashload/boss
[githubIssuesBadge]: https://img.shields.io/github/issues/hashload/boss
[githubPullRequestsBadge]: https://img.shields.io/github/issues-pr/hashload/boss
[deepwikiBadge]: https://deepwiki.com/badge.svg
[telegramBadge]: https://img.shields.io/badge/telegram-join%20channel-7289DA?style=flat-square
[telegramLink]: https://t.me/hashload
[repoStarsBadge]: https://img.shields.io/github/stars/hashload/boss?style=social
[craBadge]: https://img.shields.io/badge/CRA-100%25%20compliant-brightgreen
[securityPolicyLink]: SECURITY.md
[sbomBadge]: https://img.shields.io/badge/SBOM-compliant-brightgreen
