// pgclient.go - postgres backend client
//
// TODO: make the client delete expired stuff... but when?
// 1. on demand through a DeleteExpired call
// 2. On a ticker calling the same every N
//
// tricky part seems to be wait for execution when at end of program, want
// to always wait on the single exec
//
// easy to prove?
package pgclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/biztos/jsobs/backend"
)

// Override if using multiple databases in your process.
var DatabaseUrlEnvVar = "DATABASE_URL"

var DefaultTable = "obj_store"

var ErrNotFound = pgx.ErrNoRows

// PgDetailer implements backend.Detailer to describe an object.
type PgDetailer struct {
	path     string
	size     int
	expiry   *time.Time
	modified time.Time
}

// Scan scans a database row.
// TODO: figure out why this works inversely from the documentation, is it a
// bug here, or there, or what?
func (d *PgDetailer) Scan(row pgx.Row) error {
	return row.Scan(
		&d.path,
		&d.size,
		&d.expiry,
		&d.modified,
	)
}

// Path implements backend.Detailer.
func (d *PgDetailer) Path() string {
	return d.path
}

// Size implements backend.Detailer.
func (d *PgDetailer) Size() int {
	return d.size
}

// Expires implements backend.Detailer.
func (d *PgDetailer) Expires() bool {
	return d.expiry != nil
}

// Expiry implements backend.Detailer. If Expires returns false then Expiry must
// be ignored.
func (d *PgDetailer) Expiry() time.Time {
	if d.expiry == nil {
		return time.Time{} // "zero time"
	}
	return *d.expiry
}

// Modified implements backend.Detailer.
func (d *PgDetailer) Modified() time.Time {
	return d.modified
}

// PgClient is a BackendClient for PostgreSQL databases.
type PgClient struct {
	Pool            *pgxpool.Pool
	Table           string
	PurgeOnShutdown bool
}

// String returns an identifying string.
func (c *PgClient) String() string {
	return fmt.Sprintf("pgclient (table=%s)", c.Table)
}

// New returns a new PgClient using DefaultTable and a pool from
// DatabaseUrlEnvVar, with PurgeOnShutdown true.
func New() (*PgClient, error) {
	dburl := os.Getenv(DatabaseUrlEnvVar)
	if dburl == "" {
		return nil, errors.New(DatabaseUrlEnvVar + " not defined in env")
	}

	pool, err := pgxpool.New(context.Background(), dburl)
	if err != nil {
		return nil, err
	}
	client := &PgClient{
		Pool:            pool,
		Table:           DefaultTable,
		PurgeOnShutdown: true,
	}
	return client, nil
}

// SaveRaw saves the raw_obj to the database with no expiry.
func (c *PgClient) SaveRaw(path string, raw_obj []byte) error {

	_, err := c.Pool.Exec(context.Background(), c.saveSql(),
		path, raw_obj, len(raw_obj), nil, time.Now())
	return err

}

// SaveRawExpiry saves the raw bytes to the database for availability until
// expiry.
func (c *PgClient) SaveRawExpiry(path string, raw_obj []byte, expiry time.Time) error {

	_, err := c.Pool.Exec(context.Background(), c.saveSql(),
		path, raw_obj, len(raw_obj), expiry, time.Now())
	return err
}

// LoadRaw retrieves the object at path and returns its raw value.
// If the object does not exist, the error returned will be ErrNotFound.
func (c *PgClient) LoadRaw(path string) ([]byte, error) {

	row := c.Pool.QueryRow(context.Background(), c.loadSql(), path)
	var data []byte
	err := row.Scan(&data)
	return data, err

}

// LoadDetail retrieves the details of the object at path and returns its
// a jsobs.Detailer.
// If the object does not exist, the error returned will be ErrNotFound.
func (c *PgClient) LoadDetail(path string) (backend.Detailer, error) {

	row := c.Pool.QueryRow(context.Background(), c.loadDetailSql(), path)
	detail := &PgDetailer{}
	err := detail.Scan(row)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

// Delete deletes the object at path.
//
// For security reasons, if the object is already expired it will still be
// deleted.  (It would be possible to DELETE RETURNING and thus check the
// expiry and return a different error if the caller deletes an expired
// object, but that use-case seems silly.  You want it gone, we make it gone.)
//
// If the object does not exist, the error returned will be ErrNotFound.
func (c *PgClient) Delete(path string) error {

	tag, err := c.Pool.Exec(context.Background(), c.deleteSql(), path)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// List returns an array of all objects beginning with prefix.  An empty array
// is not considered an error.
func (c *PgClient) List(prefix string) ([]string, error) {

	// NOTE: using starts_with so don't need to %-ify prefix.

	rows, _ := c.Pool.Query(context.Background(), c.listSql(), prefix)
	paths, err := pgx.CollectRows(rows, pgx.RowTo[string])
	return paths, err

}

// ListDetail returns an array of all Detailers describing all objects
// beginning with prefix.  An empty array is not considered an error.
//
// For S3, this operation may be slow if paged results are returned!
func (c *PgClient) ListDetail(prefix string) ([]backend.Detailer, error) {

	// NOTE: using starts_with so don't need to %-ify prefix.

	// Annoyingly, we can't use []*PgDetailer as a []Detailer because the Go
	// folks won't let os call an O(n) function, instead we have to roll our
	// own O(n) operation, yay:
	//
	// https://dusted.codes/using-go-generics-to-pass-struct-slices-for-backendace-slices
	//
	// However, lucky us, CollectRows takes care of it!
	rows, _ := c.Pool.Query(context.Background(), c.listDetailSql(), prefix)
	detailers, err := pgx.CollectRows(rows,
		func(row pgx.CollectableRow) (backend.Detailer, error) {
			d := &PgDetailer{}
			err := d.Scan(row)
			return d, err
		})
	return detailers, err

}

// Count returns the number of non-expired objects beginning with prefix.
// If none are found, zero is returned.
func (c *PgClient) Count(prefix string) (int, error) {

	// NOTE: using starts_with so don't need to %-ify prefix.
	count := -1
	row := c.Pool.QueryRow(context.Background(), c.countSql(),
		prefix)
	err := row.Scan(&count)
	return count, err

}

// CountAll returns the total number of non-expired objects in the database.
func (c *PgClient) CountAll() (int, error) {

	count := -1
	row := c.Pool.QueryRow(context.Background(), c.countAllSql())
	err := row.Scan(&count)
	return count, err
}

// Purge deletes expired items from the database.  Returns the number of rows
// deleted.
func (c *PgClient) Purge() (int, error) {

	tag, err := c.Pool.Exec(context.Background(), c.purgeSql())

	// NOTE: if you are purging more than two billion rows on a 32-bit system
	// you are insane!
	return int(tag.RowsAffected()), err

}

// Schema returns the SQL required to create this client's Table.
func (c *PgClient) Schema() string {
	return c.schemaSql()
}

// CreateTable executes the SQL returned from Schema on the current database.
// If the table exists an error is returned.  For obvious reasons there is no
// corresponding DropTable function.
func (c *PgClient) CreateTable() error {

	_, err := c.Pool.Exec(context.Background(), c.Schema())
	return err
}

// Shutdown calls Purge if PurgeOnShutdown is true.
//
// Note that it *should* be safe to kill the client in mid-purge, but you
// presumably want to purge at least once per run.
func (c *PgClient) Shutdown() error {
	if c.PurgeOnShutdown == true {
		// TODO (maybe) -- log results.
		_, err := c.Purge()
		return err
	}
	return nil

}
