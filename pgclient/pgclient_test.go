// pgclient_test.go
//
// suite rigging is in setup_test.go, actual tests are here.
//
// NOTE: using require everywhere because database screwups are the most
// common errors.
//
// Problem: mocking for error tests...
//
// First, mr jack has a very very complex mocker that is a Real Database.
//
// Second, there is a pgxpoolmock packate that is out of date and does not
// support pgx v5, making it useless to me.
//
// Ergo, idea: do a Very Simple Stupid Mock that only does what I need.
// (Also a good excuse to learn mocking vs interface based testing.)
// Ah but in order to do that, I need to make Pool an interface... fuck.
// (OK maybe worth it if I can make all my calls go to the Pool.)

package pgclient_test

import (
	"fmt"
	"os"
	"time"

	"github.com/biztos/jsobs/interf"
	"github.com/biztos/jsobs/pgclient"
)

// New *success* is tested in the setup, no need to test here.
func (suite *PgClientTestSuite) TestNewFailsBadUrl() {

	require := suite.Require()

	orig := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", orig)

	os.Setenv("DATABASE_URL", "very bogus url")
	_, err := pgclient.New()
	require.ErrorContains(err, "invalid dsn")

}

func (suite *PgClientTestSuite) TestNewFailsNoUrl() {

	require := suite.Require()

	pgclient.DatabaseUrlEnvVar = "OTHER_DATABASE_URL"
	os.Setenv("OTHER_DATABASE_URL", "")
	defer func() { pgclient.DatabaseUrlEnvVar = "DATABASE_URL" }()

	_, err := pgclient.New()
	require.ErrorContains(err, "OTHER_DATABASE_URL not defined in env")

}

func (suite *PgClientTestSuite) TestNewUsesAltEnvOK() {

	require := suite.Require()

	// ah, we test New after all...
	orig := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", orig)
	pgclient.DatabaseUrlEnvVar = "YET_ANOTHER_DATABASE_URL"
	os.Setenv("YET_ANOTHER_DATABASE_URL", orig)
	os.Setenv("DATABASE_URL", "ignore this")

	c, err := pgclient.New()
	require.NoError(err, "connect error")
	c.Table = suite.Client.Table
	_, err = c.CountAll()
	require.NoError(err, "query error")

}

func (suite *PgClientTestSuite) TestStringOK() {

	require := suite.Require()

	exp := fmt.Sprintf("pgclient (table=%s)", suite.Client.Table)
	require.Equal(exp, suite.Client.String(), "String")
}

func (suite *PgClientTestSuite) TestSaveRawFailsBadJson() {

	require := suite.Require()

	data := []byte("not json")
	err := suite.Client.SaveRaw("/any", data)
	require.ErrorContains(err, "SQLSTATE 22P02")

}

func (suite *PgClientTestSuite) TestSaveRawExpiryFailsBadJson() {

	require := suite.Require()

	data := []byte("not json")
	err := suite.Client.SaveRawExpiry("/any", data, time.Now())
	require.ErrorContains(err, "SQLSTATE 22P02")

}

func (suite *PgClientTestSuite) TestLoadRawFailsNotFound() {

	require := suite.Require()

	data, err := suite.Client.LoadRaw("/not/here")
	require.ErrorIs(err, pgclient.ErrNotFound)
	require.Nil(data)

}

func (suite *PgClientTestSuite) TestSaveRawLoadRawOK() {

	require := suite.Require()

	data := []byte(`{"json": true}`)
	err := suite.Client.SaveRaw("/any/thing.json", data)
	require.NoError(err)

	fetched, err := suite.Client.LoadRaw("/any/thing.json")
	require.NoError(err)
	require.EqualValues(data, fetched)

}

func (suite *PgClientTestSuite) TestSaveRawExpiryLoadRawOK() {

	require := suite.Require()

	data := []byte(`{"json": true}`)
	err := suite.Client.SaveRawExpiry("/any/thing.json", data,
		time.Now().Add(time.Hour))
	require.NoError(err)

	fetched, err := suite.Client.LoadRaw("/any/thing.json")
	require.NoError(err)
	require.EqualValues(data, fetched)

}

func (suite *PgClientTestSuite) TestLoadDetailFailsNotFound() {

	require := suite.Require()
	detail, err := suite.Client.LoadDetail("/nopers.json")
	require.ErrorIs(err, pgclient.ErrNotFound, "not found")
	require.Nil(detail, "detail returned")
}

func (suite *PgClientTestSuite) TestLoadDetailWithExpiryOK() {

	require := suite.Require()

	future := time.Now().Add(time.Hour)

	suite.SaveSet(1, "/detail/%d", &future)

	detail, err := suite.Client.LoadDetail("/detail/0")
	require.NoError(err, "load detail")

	require.Equal("/detail/0", detail.Path(), "path")
	require.Equal(7, detail.Size(), "size")
	require.True(detail.Expires(), "expires")
	// This is a bit wonky but the monotonic clock makes comparison hard.
	// Am I just missing an EqualTimes assertion somewhere?
	require.WithinDuration(future, detail.Expiry(), time.Second, "expiry")

	// Modified is controlled by the package and *might* vary from our input.
	require.Greater(time.Since(detail.Modified()), time.Duration(0),
		"modified not in future")
	require.Less(time.Since(detail.Modified()), time.Second,
		"modified not too old")

}

func (suite *PgClientTestSuite) TestLoadDetailWithoutExpiryOK() {

	require := suite.Require()

	suite.SaveSet(1, "/detail/%d", nil)

	detail, err := suite.Client.LoadDetail("/detail/0")
	require.NoError(err, "load detail")

	require.Equal("/detail/0", detail.Path(), "path")
	require.Equal(7, detail.Size(), "size")
	require.False(detail.Expires(), "expires")
	require.True(detail.Expiry().IsZero(), "expiry is zero time")

	// Modified is controlled by the package and *might* vary from our input.
	require.Greater(time.Since(detail.Modified()), time.Duration(0),
		"modified not in future")
	require.Less(time.Since(detail.Modified()), time.Second,
		"modified not too old")

}

func (suite *PgClientTestSuite) TestCountingOK() {

	require := suite.Require()

	// We should always start with an empty table!
	precount, err := suite.Client.CountAll()
	require.NoError(err)
	require.Equal(0, precount, "nothing there yet")

	suite.SaveSet(15, "/foo/%d", nil)

	postcount, err := suite.Client.Count("/foo")
	require.NoError(err)
	require.Equal(15, postcount, "saved number")

}

func (suite *PgClientTestSuite) TestListOK() {

	require := suite.Require()

	saved := suite.SaveSet(15, "/any/thing/%02d.json", nil)

	paths, err := suite.Client.List("/any/thing")
	require.NoError(err)
	require.EqualValues(saved.Paths, paths, "paths")

}

func (suite *PgClientTestSuite) TestListEmptyOK() {

	require := suite.Require()
	paths, err := suite.Client.List("/no/such/stuff")
	require.NoError(err)
	require.EqualValues([]string{}, paths, "empty path list")

}

func (suite *PgClientTestSuite) TestListDetailEmptyOK() {

	require := suite.Require()
	detailers, err := suite.Client.ListDetail("/")
	require.NoError(err)
	require.EqualValues([]interf.Detailer{}, detailers, "empty list")
}

func (suite *PgClientTestSuite) TestListDetailsOK() {

	require := suite.Require()

	// Create stuff we want to check, including expiring and non-expiring.
	future := time.Now().Add(time.Hour)
	suite.SaveSet(3, "/detail/exp/%d", &future)
	suite.SaveSet(3, "/detail/noexp/%d", nil)

	detailers, err := suite.Client.ListDetail("/detail/")
	require.NoError(err)
	require.Equal(6, len(detailers), "detailers returned")

	// They will have come back in path order.
	for i, d := range detailers {
		if i < 3 {
			require.Equal(fmt.Sprintf("/detail/exp/%d", i), d.Path(), "path %d", i)
			require.True(d.Expires(), "expires %d", i)
			require.WithinDuration(future, d.Expiry(), time.Second, "expiry %d", i)

		} else {
			require.Equal(fmt.Sprintf("/detail/noexp/%d", i-3), d.Path(), "path %d", i)
			require.False(d.Expires(), "expires %d", i)
			require.True(d.Expiry().IsZero(), "expiry is zero time %d", i)

		}

		// Modified is controlled by the package and *might* vary from our input.
		require.Greater(time.Since(d.Modified()), time.Duration(0),
			"modified not in future %d", i)
		require.Less(time.Since(d.Modified()), time.Second,
			"modified not too old %d", i)

	}

}

func (suite *PgClientTestSuite) TestDeleteFailsNotFound() {

	require := suite.Require()

	err := suite.Client.Delete("/not/there")
	require.ErrorIs(err, pgclient.ErrNotFound)
}

func (suite *PgClientTestSuite) TestDeleteFailsStupidSqlHack() {

	require := suite.Require()

	// It's 1:05am, and no, I would not really expect this to pass code review
	// but I need to move on to other things here, and I don't have time to do
	// a whole pgxpool mock (or update the existing ones, which would be nice
	// of me).
	pgclient.DeleteTestingHackKeyWord = "Embarrassing hack I know but "
	defer func() { pgclient.DeleteTestingHackKeyWord = "DELETE" }()

	err := suite.Client.Delete("/anything")
	require.ErrorContains(err, "SQLSTATE 42601")
}

func (suite *PgClientTestSuite) TestDeleteOK() {

	require := suite.Require()
	suite.SaveSet(1, "/any/thing/%d.json", nil)

	err := suite.Client.Delete("/any/thing/0.json")
	require.NoError(err)
	require.Equal(0, suite.FullCount(), "full count")

}

func (suite *PgClientTestSuite) TestPurgeOK() {

	require := suite.Require()

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	keep_a := suite.SaveSet(5, "/keep/a/%02d.json", &future)
	suite.SaveSet(5, "/purge/%02d.json", &past)
	keep_b := suite.SaveSet(5, "/keep/b/%02d.json", nil)

	require.Equal(15, suite.FullCount(), "full count")

	purged, err := suite.Client.Purge()
	require.NoError(err, "purge")
	require.Equal(5, purged, "purge count")

	require.Equal(10, suite.FullCount(), "full count")

	count, err := suite.Client.CountAll()
	require.NoError(err, "countall")
	require.Equal(10, count)

	// Now we should have the keeper paths.
	keep_paths := make([]string, 0, len(keep_a.Paths)+len(keep_b.Paths))
	keep_paths = append(keep_paths, keep_a.Paths...)
	keep_paths = append(keep_paths, keep_b.Paths...)

	paths, err := suite.Client.List("/keep/")
	require.NoError(err)

	require.EqualValues(keep_paths, paths, "path list")

}

func (suite *PgClientTestSuite) TestPurgeOnlyOnceOK() {

	require := suite.Require()

	// This is rank hackery here... because in a Very Fast System it might
	// still fail! Gonna need the mock...
	past := time.Now().Add(-1 * time.Hour)
	suite.SaveSet(100, "/keep/a/%02d.json", &past)
	go suite.Client.Purge()
	time.Sleep(time.Microsecond)   // annoying but necessary!
	_, err := suite.Client.Purge() // other one still purging!
	require.ErrorContains(err, "already purging")

}

func (suite *PgClientTestSuite) TestShutdownOK() {

	require := suite.Require()

	// Assumptions:
	// 1) a purge that deletes stuff is slower than a purge that doesn't.
	// 2) ...by enough that it's still working in the goroutine when we call
	//    Shutdown.
	require.True(suite.Client.PurgeOnShutdown, "PurgeOnShutdown") // by default
	start := time.Now()
	suite.Client.Shutdown()
	shorter_time := time.Since(start)

	// Now give it something to delete, and don't call an extra shutdown.
	suite.Client.PurgeOnShutdown = false
	past := time.Now().Add(-1 * time.Hour)
	suite.SaveSet(50, "/purge/%02d.json", &past)
	start = time.Now()
	go suite.Client.Purge()
	time.Sleep(time.Microsecond) // annoying but necessary!
	require.True(suite.Client.Purging(), "Purging")
	suite.Client.Shutdown()
	longer_time := time.Since(start)
	require.Equal(0, suite.FullCount(), "full count")

	require.Greater(longer_time, shorter_time, "real purge longer")

}
