// sql.go -- SQL snippets go here.
//
// NOTE: using starts_with instead of LIKE for prefix matching, because it
// *seems like* that should be at least as fast, maybe faster -- certainly
// it is faster to scan based on that -- but in case there is some magic
// happening with LIKE, or anti-magic with starts_with (highly doubtful in
// both cases) it would behoove us to run some benchmarks.

package pgclient

import "fmt"

func (c *PgClient) saveSql() string {
	f := `INSERT INTO %s (obj_path,data,size,expiry,modified)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (obj_path)
DO UPDATE SET
	data = EXCLUDED.data,
	size = EXCLUDED.size,
	expiry = EXCLUDED.expiry,
	modified = EXCLUDED.modified;`
	return fmt.Sprintf(f, c.Table)
}

// This is a really annoying hack, but it saves us the trouble of making a
// pgxpool mock just for this package:
var DeleteTestingHackKeyWord = "DELETE"

func (c *PgClient) deleteSql() string {
	f := "%s FROM %s WHERE obj_path = $1;"
	return fmt.Sprintf(f, DeleteTestingHackKeyWord, c.Table)

}

func (c *PgClient) listSql() string {
	f := `SELECT obj_path
FROM %s
WHERE starts_with(obj_path,$1) = true AND (expiry IS NULL OR expiry > now())
ORDER BY obj_path;`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) listDetailSql() string {
	f := `SELECT obj_path,size,expiry,modified
FROM %s
WHERE starts_with(obj_path,$1) = true AND (expiry IS NULL OR expiry > now())
ORDER BY obj_path;`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) countSql() string {
	f := `SELECT COUNT(*)
FROM %s
WHERE starts_with(obj_path,$1) = true AND (expiry IS NULL OR expiry > now());`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) countAllSql() string {
	f := `SELECT COUNT(*)
FROM %s
WHERE expiry IS NULL OR expiry > now();`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) loadSql() string {
	f := `SELECT data
FROM %s
WHERE obj_path = $1 AND (expiry IS NULL or expiry > now());`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) loadDetailSql() string {
	f := `SELECT obj_path,size,expiry,modified
FROM %s
WHERE starts_with(obj_path,$1) = true AND (expiry IS NULL or expiry > now());`
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) purgeSql() string {
	f := "DELETE FROM %s WHERE expiry <= now();"
	return fmt.Sprintf(f, c.Table)

}

func (c *PgClient) schemaSql() string {

	f := `CREATE TABLE %s (
	obj_path TEXT NOT NULL PRIMARY KEY,
	data JSONB NOT NULL,
	size INT NOT NULL,
	expiry TIMESTAMP WITH TIME ZONE NULL,
	modified TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX %s_expiry_idx ON %s USING btree (expiry);`

	return fmt.Sprintf(f, c.Table, c.Table, c.Table)

}
