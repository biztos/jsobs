// jsobs_test.go

package jsobs_test

import (
	"errors"
	"os"
	"time"

	"github.com/biztos/jsobs"
	"github.com/biztos/jsobs/backend"
	"github.com/biztos/jsobs/pgclient"
)

func (suite *JsobsTestSuite) TestNewPgClientFailsBadUrl() {

	// trying to be very pass-through on this... so a basic failure case
	// should be all the testing required for coverage of NewXXClient for
	// any future XX, and we test its New separately.

	require := suite.Require()

	orig := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", orig)

	os.Setenv("DATABASE_URL", "very bogus url")
	_, err := jsobs.NewPgClient()
	require.ErrorContains(err, "invalid dsn")
}

func (suite *JsobsTestSuite) TestCountError() {

	require := suite.Require()

	exp_err := errors.New("boo count")
	suite.Backend.nextError = exp_err
	_, err := suite.Client.Count("/any")
	require.ErrorIs(err, exp_err)
	require.EqualValues([]string{"Count"}, suite.Backend.allCalls, "calls")
	require.Equal("/any", suite.Backend.lastPrefix, "prefix passed")
}

func (suite *JsobsTestSuite) TestCountOK() {

	require := suite.Require()

	suite.Backend.nextCount = 123456789
	count, err := suite.Client.Count("/any")
	require.NoError(err)
	require.Equal(count, 123456789, "count")
	require.EqualValues([]string{"Count"}, suite.Backend.allCalls, "calls")
	require.Equal("/any", suite.Backend.lastPrefix, "prefix passed")

}

func (suite *JsobsTestSuite) TestCountAllError() {

	require := suite.Require()

	exp_err := errors.New("boo count all")
	suite.Backend.nextError = exp_err
	_, err := suite.Client.CountAll()
	require.ErrorIs(err, exp_err)
	require.EqualValues([]string{"CountAll"}, suite.Backend.allCalls, "calls")
}

func (suite *JsobsTestSuite) TestCountAllOK() {

	require := suite.Require()

	suite.Backend.nextCount = 123456789
	count, err := suite.Client.CountAll()
	require.NoError(err)
	require.Equal(count, 123456789, "count")
	require.EqualValues([]string{"CountAll"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestShutdownOK() {

	require := suite.Require()

	code := 3
	suite.Client.Shutdown(code)
	require.EqualValues([]string{"Shutdown"}, suite.Backend.allCalls, "calls")
	require.True(suite.Exited)
	require.Equal(code, suite.ExitCode)

}

func (suite *JsobsTestSuite) TestShutdownWithBackendErrorOK() {

	require := suite.Require()

	suite.Backend.nextError = errors.New("backend fail")

	code := 3
	suite.Client.Shutdown(code)
	require.EqualValues([]string{"Shutdown"}, suite.Backend.allCalls, "calls")
	require.True(suite.Exited)
	require.Equal(99, suite.ExitCode)

}

func (suite *JsobsTestSuite) TestListDetailError() {

	require := suite.Require()

	exp_err := errors.New("detailerrrr")
	suite.Backend.nextError = exp_err
	detailers, err := suite.Client.ListDetail("/any")
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPrefix, "prefix")
	require.Equal(0, len(detailers), "detailers")
	require.EqualValues([]string{"ListDetail"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestListDetailOK() {

	require := suite.Require()

	exp_detailers := []backend.Detailer{
		&TestDetailer{path: "/foo"},
		&TestDetailer{path: "/bar"},
	}
	suite.Backend.nextDetailers = exp_detailers

	detailers, err := suite.Client.ListDetail("/any")
	require.NoError(err)
	require.EqualValues(exp_detailers, detailers, "detailers")
	require.Equal("/any", suite.Backend.lastPrefix, "prefix")
	require.EqualValues([]string{"ListDetail"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestListError() {

	require := suite.Require()

	exp_err := errors.New("listbomb")
	suite.Backend.nextError = exp_err
	paths, err := suite.Client.List("/any")
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPrefix, "prefix")
	require.Equal(0, len(paths), "paths")
	require.EqualValues([]string{"List"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestListOK() {

	require := suite.Require()

	exp_paths := []string{"/foo", "/bar", "/baz"}
	suite.Backend.nextPaths = exp_paths

	paths, err := suite.Client.List("/any")
	require.NoError(err)
	require.EqualValues(exp_paths, paths, "paths")
	require.Equal("/any", suite.Backend.lastPrefix, "prefix")
	require.EqualValues([]string{"List"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestDeleteError() {

	require := suite.Require()

	exp_err := errors.New("delete fail")
	suite.Backend.nextError = exp_err
	err := suite.Client.Delete("/any")
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"Delete"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestDeleteOK() {

	require := suite.Require()

	err := suite.Client.Delete("/any")
	require.NoError(err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"Delete"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadDetailError() {

	require := suite.Require()

	exp_err := errors.New("load fail")
	suite.Backend.nextError = exp_err

	detailer, err := suite.Client.LoadDetail("/any")
	require.ErrorIs(err, exp_err)
	require.Nil(detailer)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadDetail"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadDetailOK() {

	require := suite.Require()

	exp_detailer := &TestDetailer{path: "/any"}
	suite.Backend.nextDetailer = exp_detailer

	detailer, err := suite.Client.LoadDetail("/any")
	require.NoError(err)
	require.EqualValues(exp_detailer, detailer, "detailer")

	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadDetail"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadRawError() {

	require := suite.Require()

	exp_err := errors.New("oof bammo")
	suite.Backend.nextError = exp_err

	data, err := suite.Client.LoadRaw("/any")
	require.ErrorIs(err, exp_err)
	require.Nil(data)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadRawOK() {

	require := suite.Require()

	exp_data := []byte("anything here")
	suite.Backend.nextData = exp_data

	data, err := suite.Client.LoadRaw("/any")
	require.NoError(err)
	require.EqualValues(exp_data, data, "data")

	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadError() {

	require := suite.Require()

	exp_err := errors.New("loooooaaaaad buum")
	suite.Backend.nextError = exp_err

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{}

	err := suite.Client.Load("/any", obj)
	require.ErrorIs(err, exp_err)
	require.EqualValues(&Thing{}, obj)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadJsonError() {

	require := suite.Require()

	suite.Backend.nextData = []byte("not json")

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{}

	err := suite.Client.Load("/any", obj)
	require.ErrorContains(err, "Failed to marshal JSON")
	require.EqualValues(&Thing{}, obj)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestLoadOK() {

	require := suite.Require()

	exp_data := []byte(`{"foo":"hello","bar":42}`)
	suite.Backend.nextData = exp_data

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{}

	err := suite.Client.Load("/any", obj)
	require.NoError(err)
	require.EqualValues(&Thing{"hello", 42}, obj)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.EqualValues([]string{"LoadRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveRawExpiryError() {

	require := suite.Require()

	data := []byte("helo")
	expiry := time.Now()

	exp_err := errors.New("no save safe")
	suite.Backend.nextError = exp_err

	err := suite.Client.SaveRawExpiry("/any", data, expiry)
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(data, suite.Backend.lastData, "data")
	require.Equal(expiry, suite.Backend.lastExpiry, "expiry")
	require.EqualValues([]string{"SaveRawExpiry"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveRawExpiryOK() {

	require := suite.Require()

	data := []byte("helo")
	expiry := time.Now()

	err := suite.Client.SaveRawExpiry("/any", data, expiry)
	require.NoError(err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(data, suite.Backend.lastData, "data")
	require.Equal(expiry, suite.Backend.lastExpiry, "expiry")
	require.EqualValues([]string{"SaveRawExpiry"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveRawError() {

	require := suite.Require()

	data := []byte("helo again")

	exp_err := errors.New("no save safe")
	suite.Backend.nextError = exp_err

	err := suite.Client.SaveRaw("/any", data)
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(data, suite.Backend.lastData, "data")
	require.EqualValues([]string{"SaveRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveRawOK() {

	require := suite.Require()

	data := []byte("helo still")

	err := suite.Client.SaveRaw("/any", data)
	require.NoError(err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(data, suite.Backend.lastData, "data")
	require.EqualValues([]string{"SaveRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveExpiryError() {

	require := suite.Require()

	exp_err := errors.New("saaave buum")
	suite.Backend.nextError = exp_err

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{"here", 99}
	exp_data := []byte(`{"Foo":"here","Bar":99}`)
	expiry := time.Now()

	err := suite.Client.SaveExpiry("/any", obj, expiry)
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(exp_data, suite.Backend.lastData, "data")
	require.Equal(expiry, suite.Backend.lastExpiry, "expiry")
	require.EqualValues([]string{"SaveRawExpiry"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveExpiryJsonError() {

	require := suite.Require()

	obj := DoesNotMarshal("KABOOM")
	expiry := time.Now()

	err := suite.Client.SaveExpiry("/any", obj, expiry)
	require.ErrorContains(err, "KABOOM")
	// no calls!
	require.Nil(suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveExpiryOK() {

	require := suite.Require()

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{"here", 99}
	exp_data := []byte(`{"Foo":"here","Bar":99}`)
	expiry := time.Now()

	err := suite.Client.SaveExpiry("/any", obj, expiry)
	require.NoError(err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(exp_data, suite.Backend.lastData, "data")
	require.EqualValues([]string{"SaveRawExpiry"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveError() {

	require := suite.Require()

	exp_err := errors.New("saaave buum")
	suite.Backend.nextError = exp_err

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{"here", 99}
	exp_data := []byte(`{"Foo":"here","Bar":99}`)

	err := suite.Client.Save("/any", obj)
	require.ErrorIs(err, exp_err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(exp_data, suite.Backend.lastData, "data")
	require.EqualValues([]string{"SaveRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveJsonError() {

	require := suite.Require()

	obj := DoesNotMarshal("KABOOM")

	err := suite.Client.Save("/any", obj)
	require.ErrorContains(err, "KABOOM")
	// no calls!
	require.Nil(suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestSaveOK() {

	require := suite.Require()

	type Thing struct {
		Foo string
		Bar int
	}
	obj := &Thing{"here", 99}
	exp_data := []byte(`{"Foo":"here","Bar":99}`)

	err := suite.Client.Save("/any", obj)
	require.NoError(err)
	require.Equal("/any", suite.Backend.lastPath, "path")
	require.Equal(exp_data, suite.Backend.lastData, "data")
	require.EqualValues([]string{"SaveRaw"}, suite.Backend.allCalls, "calls")

}

func (suite *JsobsTestSuite) TestIsNotFound() {

	require := suite.Require()

	// right now it's just dealing with pg so... Ye Olde Filler Teste!
	require.True(jsobs.IsNotFound(pgclient.ErrNotFound), "not found")
	require.False(jsobs.IsNotFound(errors.New("X")), "not not found")

}
