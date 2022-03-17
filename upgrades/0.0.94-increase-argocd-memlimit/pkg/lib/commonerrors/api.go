// Package commonerrors contains common errors
package commonerrors

import "errors"

// ErrUserAborted indicates that the user aborted
var ErrUserAborted = errors.New("aborted by user")
