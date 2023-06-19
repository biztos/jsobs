// interf/interfaces.go -- just the interfaces!
//
// Here to avoid circular dependencies between jsobs and the backends, both
// of which need to know about these interfaces.

// Package interf defines the independent interfaces used by jsobs.
package interf

import (
	"time"
)

// Detailer identifies an object in detail.  It is returned by ListDetail and
// LoadDetail.
type Detailer interface {
	Path() string
	Size() int
	Modified() time.Time
	Expires() bool
	Expiry() time.Time
}

// BackendClient is the client that talks to the back-end storage.
type BackendClient interface {
	String() string
	SaveRaw(path string, raw_obj []byte) error
	SaveRawExpiry(path string, raw_obj []byte, expiry time.Time) error
	LoadRaw(path string) ([]byte, error)
	LoadDetail(path string) (Detailer, error)
	Delete(path string) error
	List(prefix string) ([]string, error)
	ListDetail(prefix string) ([]Detailer, error)
	Count(prefix string) (int, error)
	CountAll() (int, error)
	Shutdown()
}
