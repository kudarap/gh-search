package ghsearch

import "errors"

var ErrUserSourceFailed = errors.New("user source failure")
var ErrUserSourceTimeout = errors.New("user source timed out")

type SourceError struct {
	ErrString string
}

func (e SourceError) Error() string {
	return e.ErrString
}
