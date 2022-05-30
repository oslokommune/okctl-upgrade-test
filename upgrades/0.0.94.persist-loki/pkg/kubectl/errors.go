package kubectl

import "regexp"

var reErrNotFound = regexp.MustCompile(`Error from server \(NotFound\): pods ".+" not found\W`)

func isErrNotFound(err error) bool {
	return reErrNotFound.MatchString(err.Error())
}
