// Package main is the entry point for the Boss dependency manager CLI.
// Boss is a dependency manager for Delphi projects, similar to npm for JavaScript.
package main

import (
	"github.com/hashload/boss/cmd"
	"github.com/hashload/boss/pkg/msg"
)

// main is the entry point of the application.
func main() {
	if err := cmd.Execute(); err != nil {
		msg.Die(err.Error())
	}
}
