
# Dependency Manager for Delphi

Instalation: 
 * Download [Latest release](https://github.com/HashLoad/boss/releases/latest/) from boss.exe
 * Put in your path
 * Just type `boss` in cmd
```

Usage:
  boss [command]

Available Commands:
  help        Help about any command
  init        Initialize a new project
  install     Install a dependency
  login       Register login to repo
  publish     publish a dependency
  remove      Remove a dependency
  run         Run cmd script

Flags:
  -h, --help   help for boss

Use "boss [command] --help" for more information about a command.
```
+ Sample: 
	+ `boss install github.com/HashLoad/horse`
	+ `boss install github.com/HashLoad/horse:1.0.0`


### For use yor project in boss create a tag with [semantic version](https://semver.org/) 
