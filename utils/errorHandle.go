package utils

import "github.com/hashload/boss/msg"

func HandleError(err error) {
	if err != nil {
		msg.Err(err.Error())
	}
}
