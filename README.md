<p align="center">
  <a href="https://github.com/HashLoad/boss/blob/master/img/png/sized/Boss%20Logo%20-%20128px.png">
    <img alt="Boss" src="https://github.com/HashLoad/boss/blob/master/img/png/sized/Boss%20Logo%20-%20128px.png">
  </a>  
</p><br>
<p align="center">
 <b>Boss</b> is an open source dependency manager inspired by <a href="https://www.npmjs.com/">npm</a><br>for projects developed in <b>Delphi</b> and <b>Lazarus</b>.
</p><br>
<p align="center">
  <a href="https://t.me/hashload">
    <img src="https://img.shields.io/badge/telegram-join%20channel-7289DA?style=flat-square">
  </a>
</p>

![Go](https://github.com/hashload/boss/workflows/Go/badge.svg)

### [Getting started](https://medium.com/@matheusarendthunsche/come%C3%A7ando-com-o-boss-72aad9bcc13)

Installation:
 * Download [setup](https://github.com/hashload/boss/releases)
 * Just type `boss` in cmd
 * (Optional) Install a [Boss Delphi IDE complement](https://github.com/hashload/boss-ide)

## Available Commands

### > Init
This command initialize a new project. Add `-q` or `--quiet` to initialize the boss with default values.
```
boss init
boss init -q
boss init --quiet
```

### > Install
This command install a new dependency
```
boss install <dependency>
```
###### Aliases: i, add

### > Uninstall
This command uninstall a dependency
```
boss uninstall <dependency>
```
###### Aliases: remove, rm, r, un, unlink

### > Cache
This command removes the cache
```
 boss config cache rm
```
###### Aliases: remove, rm, r

### > Dependencies
This command print all dependencies and your versions. To see versions, add aliases `-v`
```
boss dependencies
boss dependencies -v
```
###### Aliases: dep, ls, list, ll, la

### > Version
This command show the client version
```
boss v
boss version
boss -v
boss --version
```
###### Aliases: v

### > Update
This command update installed dependencies
```
boss update
```
###### Aliases: up

### > Upgrade
This command upgrade the client latest version. Add `--dev` to upgrade to the latest pre-release.
```
boss upgrade
boss upgrade --dev
```

### > login
This command Register login to repo
```
boss login <repo>
boss adduser <repo>
boss add-user <repo>
boss login <repo> -u UserName -p Password
boss login <repo> -k PrivateKey -p PassPhrase
```
###### Aliases: adduser, add-user

## Flags

### > Global
This flag defines a global environment
```
boss --global
```
###### Aliases: -g

### > Help
This is a helper for boss. Use `boss <command> --help` for more information about a command.
```
boss --help
```
###### Aliases: -h

## Another commands
```
delphi           Configure Delphi version
gc               Garbage collector  
publish          Publish package to registry
run              Run cmd script
```

## Samples
```
boss install horse
boss install horse:1.0.0
boss install -g delphi-docker
boss install -g boss-ide
```

## Using [semantic versioning](https://semver.org/) to specify update types your package can accept

You can specify which update types your package can accept from dependencies in your package’s boss.json file.

For example, to specify acceptable version ranges up to 1.0.4, use the following syntax:
 * Patch releases: 1.0 or 1.0.x or ~1.0.4
 * Minor releases: 1 or 1.x or ^1.0.4
 * Major releases: * or x

## 💻 Code Contributors

<a href="https://github.com/Hashload/boss/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Hashload/boss" />
</a>

