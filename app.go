package main

import (
	"github.com/hashload/boss/cmd"
	"github.com/hashload/boss/pkg/msg"
)

func main() {
	if err := cmd.Execute(); err != nil {
		msg.Die(err.Error())
	}
}
