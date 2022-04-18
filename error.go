package ghsearch

import "errors"

var (
	ErrUserSourceFailed  = errors.New("user source failure")
	ErrUserSourceTimeout = errors.New("user source timed out")
)

// SourceError represents an error from a source.
type SourceError struct {
	Err error
}

func (e SourceError) Error() string {
	return e.Error()
}

func NewSourceError(e error) error {
	if e == nil {
		return nil
	}
	return &SourceError{e}
}
