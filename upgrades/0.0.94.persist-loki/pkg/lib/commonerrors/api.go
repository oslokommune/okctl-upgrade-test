// Package commonerrors defines commonly used errors across upgrades
package commonerrors

import "errors"

var (
	// ErrUserAborted indicates operations were canceled by the user
	ErrUserAborted = errors.New("aborted by user")
	// ErrNoActionRequired indicates the relevant operations has already been done
	ErrNoActionRequired = errors.New("no action required")
)
