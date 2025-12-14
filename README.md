# Boss

![Boss][bossLogo]

[![Go Report Card][goReportBadge]][goReportLink]
[![GitHub release (latest by date)][latestReleaseBadge]](https://github.com/HashLoad/boss/releases/latest)
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

### > Init

Initialize a new project and create a `boss.json` file. Add `-q` or `--quiet` to skip interactive prompts and use default values.

```shell
boss init
boss init -q
boss init --quiet
```

### > Install

Install one or more dependencies with real-time progress tracking:

```shell
boss install <dependency>
```

**Progress Tracking:** Boss displays progress for each dependency being installed:

```
⏳ horse                          Waiting...
🧬 dataset-serialize              Cloning...
🔍 jhonson                        Checking...
🔥 redis-client                   Installing...
📦 boss-core                      Installed
```

The dependency name is case insensitive. For example, `boss install horse` is the same as `boss install HORSE`.

```shell
boss install horse                        # HashLoad organization on GitHub
boss install fake/horse                   # Fake organization on GitHub
boss install gitlab.com/fake/horse        # Fake organization on GitLab
boss install https://gitlab.com/fake/horse # Full URL
```

You can also specify the compiler version and platform:

```sh
boss install --compiler=37.0 --platform=Win64
```

> Aliases: i, add

### > Uninstall

Remove a dependency from the project:

```sh
boss uninstall <dependency>
```

> Aliases: remove, rm, r, un, unlink

### > Update

Update all installed dependencies to their latest compatible versions:

```sh
boss update
```

> Aliases: up

### > Upgrade

Upgrade the Boss CLI to the latest version. Add `--dev` to upgrade to the latest pre-release:

```sh
boss upgrade
boss upgrade --dev
```

### > Dependencies

List all project dependencies in a tree format. Add `-v` to show version information:

```shell
boss dependencies
boss dependencies -v
boss dependencies <package>
boss dependencies <package> -v
```

> Aliases: dep, ls, list, ll, la, dependency

### > Run

Execute a custom script defined in your `boss.json` file. Scripts are defined in the `scripts` section:

```json
{
  "name": "my-project",
  "scripts": {
    "build": "msbuild MyProject.dproj",
    "test": "MyProject.exe --test",
    "clean": "del /s *.dcu"
  }
}
```

```sh
boss run build
boss run test
boss run clean
```

### > Login

Register credentials for a repository. Useful for private repositories:

```sh
boss login <repo>
boss login <repo> -u UserName -p Password
boss login <repo> -s -k PrivateKey -p PassPhrase  # SSH authentication
```

> Aliases: adduser, add-user

### > Logout

Remove saved credentials for a repository:

```sh
boss logout <repo>
```

### > Version

Show the Boss CLI version:

```shell
boss version
boss v
boss -v
boss --version
```

> Aliases: v

## Global Flags

### > Global (-g)

Use global environment for installation. Packages installed globally are available system-wide:

```sh
boss install -g <dependency>
boss --global install <dependency>
```

### > Debug (-d)

Enable debug mode to see detailed output:

```sh
boss install --debug
boss -d install
```

### > Help (-h)

Show help for any command:

```sh
boss --help
boss <command> --help
```

## Configuration

### > Cache

Manage the Boss cache. Remove all cached modules to free up disk space:

```sh
boss config cache rm
```

> Aliases: purge, clean

### > Delphi Version

You can configure which Delphi version BOSS should use for compilation. This is useful when you have multiple Delphi versions installed.

#### List available versions

Lists all detected Delphi installations (32-bit and 64-bit) with their indexes.

```sh
boss config delphi list
```

#### Select a version

Selects a specific Delphi version to use globally. You can use the index from the list command, the version number, or the version with architecture.

```sh
boss config delphi use <index>
# or
boss config delphi use <version>
# or
boss config delphi use <version>-<arch>
```

Example:
```sh
boss config delphi use 0
boss config delphi use 37.0
boss config delphi use 37.0-Win64
```

### > Git Client

You can configure which Git client BOSS should use.

- `embedded`: Uses the built-in go-git client (default).
- `native`: Uses the system's installed git client (git.exe).

Using `native` is recommended on Windows if you need support for `core.autocrlf` (automatic line ending conversion).

```sh
boss config git mode native
# or
boss config git mode embedded
```

#### Shallow Clone

You can enable shallow cloning to significantly speed up dependency downloads. Shallow clones only fetch the latest commit without the full git history, reducing download size dramatically (e.g., from 127 MB to <1 MB for large repositories).

```sh
# Enable shallow clone (faster, recommended for CI/CD)
boss config git shallow true

# Disable shallow clone (full history)
boss config git shallow false
```

**Note:** Shallow clone is disabled by default to maintain compatibility. When enabled, you won't have access to the full git history of dependencies.

You can also temporarily enable shallow clone using an environment variable:

```sh
# Windows
set BOSS_GIT_SHALLOW=1
boss install

# Linux/macOS
BOSS_GIT_SHALLOW=1 boss install
```

### > Project Toolchain

You can also specify the required compiler version and platform in your project's `boss.json` file. This ensures that everyone working on the project uses the correct toolchain.

Add a `toolchain` section to your `boss.json`:

```json
{
  "name": "my-project",
  "version": "1.0.0",
  "toolchain": {
    "delphi": "37.0",
    "platform": "Win64"
  }
}
```

Supported fields in `toolchain`:
- `delphi`: The Delphi version (e.g., "37.0").
- `compiler`: The compiler version (e.g., "37.0").
- `platform`: The target platform ("Win32" or "Win64").
- `path`: Explicit path to the compiler (optional).
- `strict`: If true, fails if the exact version is not found (optional).

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
