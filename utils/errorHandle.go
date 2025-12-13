package utils

import "github.com/hashload/boss/pkg/msg"

// HandleError prints an error message if err is not nil
func HandleError(err error) {
	if err != nil {
		msg.Err(err.Error())
	}
}

// HandleErrorFatal prints an error message and exits if err is not nil
func HandleErrorFatal(err error) {
	if err != nil {
		msg.Die(err.Error())
	}
}
