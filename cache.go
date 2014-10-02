package oxygen

import (
	"errors"
	"io"
)

var (
	ErrEntityDoesNotExist = errors.New("oxygen-go cache: Entity does not exist.")
)

// A caching interface to allow for different levels of caching and consistency to
// be implemented on top of the Oxygen service. A cache service MUST be safe to use
// concurrently.
type Cache interface {
	Get(id int64, offset int64, size int) (reader io.Reader, err error)
	Put(id int64, reader io.Reader) (err error)
	Evict(id int64) error
}
