package goal

import "errors"

var (
	ErrNilContext             = errors.New("context cannot be nil")
	ErrNilCache               = errors.New("cacher cannot be nil")
	ErrEmptyDBAddress         = errors.New("db address cannot be empty")
	ErrEmptyDBDriver          = errors.New("db driver cannot be empty")
	ErrLiveQueryUnsupportedDB = errors.New("livequeries only supports postgres database")
)
