-- NOTE: default table is obj_store because jsobs, while cute, will confuse
-- your DBA sooner or later.
CREATE TABLE obj_store (
	obj_path TEXT NOT NULL PRIMARY KEY,
	data JSONB NOT NULL,
	size INT NOT NULL,
	expiry TIMESTAMP WITH TIME ZONE NULL,
	modified TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX obj_store_expiry_idx ON obj_store USING btree (expiry) ;