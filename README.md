
# Dependency Manager for Delphi

### [Getting started](https://medium.com/@matheusarendthunsche/come%C3%A7ando-com-o-boss-72aad9bcc13) 

Instalation: 
 * Download [setup](https://github.com/HashLoad/boss/releases/download/v1.5.3/setup.exe)
 * Just type `boss` in cmd
```

Usage:
  boss [command]

Available Commands:
  gc          Garbage collector
  help        Help about any command
  init        Initialize a new project
  install     Install a dependency
  login       Register login to repo
  publish     Publish package to registry
  remove      Remove a dependency
  run         Run cmd script
  update      update dependencies
  upgrade     upgrade a cli
  version     show cli version

Use "boss [command] --help" for more information about a command.

```
+ Sample: 
	+ `boss install github.com/HashLoad/horse`
	+ `boss install github.com/HashLoad/horse:1.0.0`


### For use yor project in boss create a tag with [semantic version](https://semver.org/) 
