package tool

import (
	"errors"
	"fmt"
	"runtime/debug"
)

// Recover 错误捕捉
func Recover(err *error, stack ...bool) {
	if er := recover(); er != nil {
		if err != nil {
			if len(stack) > 0 && stack[0] {
				*err = errors.New(fmt.Sprintln(er) + string(debug.Stack()))
			} else {
				*err = errors.New(fmt.Sprintln(er))
			}
		}
	}
}
