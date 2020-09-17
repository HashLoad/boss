![Go](https://github.com/HashLoad/boss/workflows/Go/badge.svg)

# Dependency Manager for Delphi

### [Getting started](https://medium.com/@matheusarendthunsche/come%C3%A7ando-com-o-boss-72aad9bcc13) 

Installation: 
 * Download [setup](https://github.com/HashLoad/boss/releases)
 * Just type `boss` in cmd
 * (Optional) Install a [Boss Delphi IDE complement](https://github.com/HashLoad/boss-ide)

## Available Commands

### > Init
This command initialize a new project. Add `--q` to initialize the boss with default values.
```
boss init
boss init --q 
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
This command print all dependencies and your versions
```
boss dependencies
```
###### Aliases: dep

### > Version
This command show the client version
```
boss version
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
login            Register login to repo
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
