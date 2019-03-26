package main

import (
	"github.com/hashload/boss/cmd"
	"github.com/hashload/boss/setup"
)

func main() {
	setup.Initialize()
	cmd.Execute()
}
