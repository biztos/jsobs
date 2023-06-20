# jsobs - Javascript Object Store

[![GoDoc][b1]][doc] [![Report Card][b2]][rpt] [![Coverage Status][b3]][cov]


[b1]: https://pkg.go.dev/badge/github.com/biztos/jsobs
[doc]: https://pkg.go.dev/github.com/biztos/jsobs
[b2]: https://goreportcard.com/badge/github.com/biztos/jsobs
[rpt]: https://goreportcard.com/report/github.com/biztos/jsobs
[b3]: https://coveralls.io/repos/github/biztos/jsobs/badge.svg
[cov]: https://coveralls.io/github/biztos/jsobs

JSOBS (pronounced "jay-sobs") is simple object storage for JSON-encodable
objects in a PostgreSQL database.

Future versions may also support SQLite, Amazon S3, and (maybe) the file
system.

## WARNING! ALPHA SOFTWARE!

This package is new (as of June 2023) and has not been tested much. Like all software, it probably contains bugs, and like all new software it probably contains a lot of them. 🪲🪲🪲

## Use Cases

### Debugging API Roundtrips

Generate a request/response pair and serialize it to JSON, then store it with
a ULID that matches the transaction in your logs.  If you find a problem, you
can reconstitute the full converstaion from the JSOBS data.

(This is the original use case that led to the JSOBS package.)

### Data Warehouse for Serialized Objects

Warehouse your objects in JSON and query them using the powerful features of
PostgreSQL's `JSONB` type.  Reconstitute the objects in code as needed. This
is a great match for ORM-centric systems.

### Short-Term Local, Long-Term Remote Storage

Want discoverability in the short term, and stable long-term storage for
things outside the window of attention?  Double-store your data with two
simple calls, each with its own TTL.

## Limitations

Besides the limitations of your database(s), please keep in mind:

### Minimal Metadata

There is no support for complex metadata at this time. Maybe later? Maybe not.

### Purge On Shutdown

Currently the only time data is purges is when Shutdown is called.  This may
be problematic for your use case!

One solution is to just call `Backend.Purge()` at your leisure.  Another,
which *may* be added here in the future, is to run the same on a timer so that
as long as the process is running, a purge happens ever (configurable) so
often.

The latter seems nicer, but it also has a big downside: in busy systems, you
might end up with resource contention.
