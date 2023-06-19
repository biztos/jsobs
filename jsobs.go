// jsobs.go

// Package jsobs provides simple storage for objects serializable to JSON.
package jsobs

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/biztos/jsobs/backend"
	"github.com/biztos/jsobs/pgclient"
)

func IsNotFound(err error) bool {
	if err == pgclient.ErrNotFound {
		return true
	}
	// other known cases here...
	return false
}

var ExitFunc = os.Exit

// Client handles save, load, list and delete operations for its Backend.
type Client struct {
	Backend backend.BackendClient
}

// New returns a client with the provided backend.  Any error returned from
// the backend creation function is returned here.
func New(bc backend.BackendClient, err error) (*Client, error) {
	return &Client{Backend: bc}, err
}

// NewPgClient returns a client with a connection pool to the PostgreSQL
// database specified at DATABASE_URL in the environment.
//
// Configuration for pgx may be included in the URL.
func NewPgClient() (*Client, error) {
	return New(pgclient.New())
}

// Save marshals obj to json and stores it at path with no expiry.
// Any existing object at path will be overwritten.
func (c *Client) Save(path string, obj any) error {

	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}
	return c.SaveRaw(path, b)

}

// SaveExpiry marshals obj to json and stores it at path with expiry set.
// Any existing object at path will be overwritten.
func (c *Client) SaveExpiry(path string, obj any, expiry time.Time) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}
	return c.SaveRawExpiry(path, b, expiry)
}

// SaveRaw behaves like Save but sends raw_obj directly.
//
// Use with caution!
func (c *Client) SaveRaw(path string, raw_obj []byte) error {
	return c.Backend.SaveRaw(path, raw_obj)
}

// SaveRawExpiry behaves like SaveExpiry but sends raw_obj directly.
//
// Use with caution!
func (c *Client) SaveRawExpiry(path string, raw_obj []byte, expiry time.Time) error {
	return c.Backend.SaveRawExpiry(path, raw_obj, expiry)
}

// Load retrieves the object at path from storage and unmarshals it to obj.
func (c *Client) Load(path string, obj any) error {

	b, err := c.LoadRaw(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, obj); err != nil {
		return fmt.Errorf("Failed to marshal JSON: %w", err)
	}
	return nil

}

// LoadRaw retrieves the object at path and returns its raw value.
func (c *Client) LoadRaw(path string) ([]byte, error) {
	return c.Backend.LoadRaw(path)
}

// LoadDetail retrieves the object at path and returns its details, but not
// the actual data.
func (c *Client) LoadDetail(path string) (backend.Detailer, error) {
	return c.Backend.LoadDetail(path)
}

// Delete deletes the object at path.
// If the object does not exist, the error returned will be ErrNotFound.
func (c *Client) Delete(path string) error {
	return c.Backend.Delete(path)
}

// List returns an array of all objects beginning with prefix.  An empty array
// is not considered an error.
//
// For S3, this operation may be slow if paged results are returned!
func (c *Client) List(prefix string) ([]string, error) {
	return c.Backend.List(prefix)
}

// ListDetail returns an array of all Detailers describing all objects
// beginning with prefix.  An empty array is not considered an error.
//
// This operation may be slow for any backend returning paged results!
func (c *Client) ListDetail(prefix string) ([]backend.Detailer, error) {
	return c.Backend.ListDetail(prefix)
}

// Count returns the number of non-expired objects beginning with prefix.
// If none are found, zero is returned.
func (c *Client) Count(prefix string) (int, error) {
	return c.Backend.Count(prefix)
}

// CountAll returns the total number of non-expired objects.
func (c *Client) CountAll() (int, error) {

	return c.Backend.CountAll()

}

// Shutdown calls Backend.Shutdown, which should perform any shutdown
// operations such as purging expired items from the pool; and then calls
// ExitFunc with the provided exit code.
//
// If Backend.Shutdown returns an error, 99 is used instead of code.
func (c *Client) Shutdown(code int) {
	// TODO: logging...
	err := c.Backend.Shutdown()
	if err != nil {
		// TODO: log this bit for sure... and maybe package var for the 99.
		code = 99
	}
	ExitFunc(code)
}
