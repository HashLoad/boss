

# Dependency Manager for Delphi

### [Getting started](https://medium.com/@matheusarendthunsche/come%C3%A7ando-com-o-boss-72aad9bcc13) 

Installation: 
 * Download [setup](https://github.com/HashLoad/boss/releases)
 * Just type `boss` in cmd
 * (Optional) Install a [Boss Delphi IDE complement](https://github.com/HashLoad/boss-ide)
 
```
Usage:
  boss [command]

Available Commands:
  delphi      Configure Delphi version
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

Flags:
  -g, --global   global environment
  -h, --help     help for boss

Use "boss [command] --help" for more information about a command.

```
+ Sample: 
	+ `boss install horse`
	+ `boss install horse:1.0.0`
	+ `boss install -g delphi-docker`
	+ `boss install -g boss-ide`


### For use yor project in boss create a tag with [semantic version](https://semver.org/) 
