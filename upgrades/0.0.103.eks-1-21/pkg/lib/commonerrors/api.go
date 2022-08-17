package commonerrors

import "errors"

var (
	ErrUserAborted = errors.New("aborted by user")
	ErrNothingToDo = errors.New("nothing to do")
)
