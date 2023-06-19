// pgclient_suite_test.go -- test suite rigging

package pgclient_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/biztos/jsobs/pgclient"

	"github.com/oklog/ulid/v2"

	"github.com/stretchr/testify/suite"
)

type PgClientTestSuite struct {
	suite.Suite
	Client *pgclient.PgClient
}

// Use one client per suite, so a connection failure stops the suite.
// Also use a new table per suite to avoid any overlap.
func (suite *PgClientTestSuite) SetupSuite() {

	require := suite.Require()

	client, err := pgclient.New()
	require.NoError(err, "client setup err")

	// Use a unique test table and drop/create it.
	client.Table = fmt.Sprintf("jsobs_test_%s", ulid.Make())

	suite.Client = client

	require.NoError(suite.DropTable(), "drop table")
	require.NoError(suite.Client.CreateTable(), "create table")

}

func (suite *PgClientTestSuite) DropTable() error {

	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s;", suite.Client.Table)
	_, err := suite.Client.Pool.Exec(context.Background(), sql)
	return err
}

// Drop the test database unless KEEP_DB is set.
func (suite *PgClientTestSuite) TearDownSuite() {

	require := suite.Require()

	val := os.Getenv("KEEP_DB")
	if val != "" && val != "0" && val != "false" {
		return
	}

	require.NoError(suite.DropTable(), "drop table")

}

// Zero out the table per test, so there are no fragments.
func (suite *PgClientTestSuite) SetupTest() {

	require := suite.Require()

	conn, err := suite.Client.Pool.Acquire(context.Background())
	require.NoError(err, "conn acquire")
	defer conn.Release()

	sql := fmt.Sprintf("TRUNCATE TABLE %s;", suite.Client.Table)
	_, err = conn.Exec(context.Background(), sql)
	require.NoError(err, "exec truncate")

}

func (suite *PgClientTestSuite) FullCount() int {

	require := suite.Require()

	conn, err := suite.Client.Pool.Acquire(context.Background())
	require.NoError(err, "conn acquire")
	defer conn.Release()

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %s;", suite.Client.Table)
	var count int
	row := conn.QueryRow(context.Background(), sql)
	err = row.Scan(&count)
	require.NoError(err, "query")
	return count

}

// see below...
type SaveSetResult struct {
	Paths    []string
	PathData map[string][]byte
}

// We need sets of data a lot, with expiry in future or not at all.
func (suite *PgClientTestSuite) SaveSet(count int, pfmt string, exp *time.Time) *SaveSetResult {

	require := suite.Require() // bail out on err

	paths := make([]string, count)
	path_data := make(map[string][]byte, count)
	for i := 0; i < count; i++ {
		path := fmt.Sprintf(pfmt, i)
		data := []byte(fmt.Sprintf(`{"n":%d}`, i))
		paths[i] = path
		path_data[path] = data
		var err error
		if exp == nil {
			err = suite.Client.SaveRaw(path, data)
		} else {
			err = suite.Client.SaveRawExpiry(path, data, *exp)
		}

		require.NoError(err, "save error")

	}
	return &SaveSetResult{
		Paths:    paths,
		PathData: path_data,
	}

}

// The actual runner func:
func TestPgClientTestSuite(t *testing.T) {
	suite.Run(t, new(PgClientTestSuite))
}
