package kubectl

import (
	"errors"
	"regexp"
)

// errNotFound indicates something is missing
var errNotFound = errors.New("not found")

var reErrNotFound = regexp.MustCompile(`Error from server \(NotFound\): [\w.]+ ".+" not found\W`)

func isErrNotFound(err error) bool {
	return reErrNotFound.MatchString(err.Error())
}
