# Boss

![Boss][bossLogo]

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

This command initialize a new project. Add `-q` or `--quiet` to initialize the boss with default values.

```shell
boss init
boss init -q
boss init --quiet
```

### > Install

This command install a new dependency

```shell
boss install <dependency>
```

The dependency is case insensitive. For example, `boss install horse` is the same as the `boss install HORSE` command.

```pascal
boss install horse // By default, look for the Horse project within the GitHub Hashload organization.
boss install fake/horse // By default, look for the Horse project within the Fake GitHub organization.
boss install gitlab.com/fake/horse // By default, searches for the Horse project within the Fake GitLab organization.
boss install https://gitlab.com/fake/horse // You can also pass the full URL for installation
```

> Aliases: i, add

### > Uninstall

This command uninstall a dependency

```sh
boss uninstall <dependency>
```

> Aliases: remove, rm, r, un, unlink

### > Cache

This command removes the cache

```sh
 boss config cache rm
```

> Aliases: remove, rm, r

### > Dependencies

This command print all dependencies and your versions. To see versions, add aliases `-v`

```shell
boss dependencies
boss dependencies -v
```

> Aliases: dep, ls, list, ll, la

### > Version

This command show the client version

```shell
boss v
boss version
boss -v
boss --version
```

> Aliases: v

### > Update

This command update installed dependencies

```sh
boss update
```

> Aliases: up

### > Upgrade

This command upgrade the client latest version. Add `--dev` to upgrade to the latest pre-release.

```sh
boss upgrade
boss upgrade --dev
```

### > login

This command Register login to repo

```sh
boss login <repo>
boss adduser <repo>
boss add-user <repo>
boss login <repo> -u UserName -p Password
boss login <repo> -k PrivateKey -p PassPhrase
```

> Aliases: adduser, add-user

## Flags

### > Global

This flag defines a global environment

```sh
boss --global
```

> Aliases: -g

### > Help

This is a helper for boss. Use `boss <command> --help` for more information about a command.

```sh
boss --help
```

> Aliases: -h

## Another commands

```sh
delphi           Configure Delphi version
gc               Garbage collector
publish          Publish package to registry
run              Run cmd script
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

## 💻 Code Contributors

![GitHub Contributors Image](https://contrib.rocks/image?repo=Hashload/boss)

[githubContributorsBadge]: https://img.shields.io/github/contributors/hashload/boss
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
