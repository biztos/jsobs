// jsobs_suite_test.go -- test suite rigging

package jsobs_test

import (
	"errors"
	"testing"
	"time"

	"github.com/biztos/jsobs"
	"github.com/biztos/jsobs/backend"

	"github.com/stretchr/testify/suite"
)

type DoesNotMarshal string

func (t DoesNotMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New(string(t))
}

type TestDetailer struct {
	path     string
	size     int
	expiry   *time.Time
	modified time.Time
}

func (d *TestDetailer) Path() string {
	return d.path
}
func (d *TestDetailer) Size() int {
	return d.size
}
func (d *TestDetailer) Expires() bool {
	return d.expiry != nil
}
func (d *TestDetailer) Expiry() time.Time {
	if d.expiry == nil {
		return time.Time{} // "zero time"
	}
	return *d.expiry
}
func (d *TestDetailer) Modified() time.Time {
	return d.modified
}

// This is a VERY simple mock because we don't really have any use case in
// which more than one backend call happens in a row before we can inspect
// it.
type TestBackend struct {
	// things we get:
	allCalls   []string
	lastCall   string // func name
	lastPath   string
	lastData   []byte
	lastExpiry time.Time
	lastPrefix string

	// things we return:
	nextError     error
	nextCount     int
	nextPaths     []string
	nextDetailer  backend.Detailer
	nextDetailers []backend.Detailer
	nextData      []byte
}

func (t *TestBackend) addCall(name string) {
	t.allCalls = append(t.allCalls, name)
	t.lastCall = name
}
func (t *TestBackend) String() string {
	t.addCall("String")
	return "test backend"
}
func (t *TestBackend) SaveRaw(path string, raw_obj []byte) error {
	t.addCall("SaveRaw")
	t.lastPath = path
	t.lastData = raw_obj
	return t.nextError
}

func (t *TestBackend) SaveRawExpiry(path string, raw_obj []byte, expiry time.Time) error {
	t.addCall("SaveRawExpiry")
	t.lastPath = path
	t.lastData = raw_obj
	t.lastExpiry = expiry
	return t.nextError
}
func (t *TestBackend) LoadRaw(path string) ([]byte, error) {
	t.addCall("LoadRaw")
	t.lastPath = path
	return t.nextData, t.nextError

}
func (t *TestBackend) LoadDetail(path string) (backend.Detailer, error) {
	t.addCall("LoadDetail")
	t.lastPath = path
	return t.nextDetailer, t.nextError

}
func (t *TestBackend) Delete(path string) error {
	t.addCall("Delete")
	t.lastPath = path
	return t.nextError

}
func (t *TestBackend) List(prefix string) ([]string, error) {
	t.addCall("List")
	t.lastPrefix = prefix
	return t.nextPaths, t.nextError

}
func (t *TestBackend) ListDetail(prefix string) ([]backend.Detailer, error) {
	t.addCall("ListDetail")
	t.lastPrefix = prefix
	return t.nextDetailers, t.nextError

}
func (t *TestBackend) Count(prefix string) (int, error) {
	t.addCall("Count")
	t.lastPrefix = prefix
	return t.nextCount, t.nextError

}
func (t *TestBackend) CountAll() (int, error) {
	t.addCall("CountAll")
	return t.nextCount, t.nextError

}
func (t *TestBackend) Shutdown() error {
	t.addCall("Shutdown")
	return t.nextError
}

type JsobsTestSuite struct {
	suite.Suite
	Client   *jsobs.Client
	Backend  *TestBackend
	Exited   bool
	ExitCode int
}

// Because we're not dealing with a real back-end here, we can set up the
// client for every test instead of once per suite.
func (suite *JsobsTestSuite) SetupTest() {

	// have to keep track of the backend separately so we can use the
	// quick-mock stuff without casting to type all the time
	suite.Backend = &TestBackend{}
	suite.Client = &jsobs.Client{Backend: suite.Backend}
	suite.Exited = false

	jsobs.ExitFunc = func(code int) {
		suite.Exited = true
		suite.ExitCode = code
	}

}

// The actual runner func:
func TestJsobsTestSuite(t *testing.T) {
	suite.Run(t, new(JsobsTestSuite))
}
