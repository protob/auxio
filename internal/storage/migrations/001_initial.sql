PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS buckets (
    name       TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    region     TEXT NOT NULL DEFAULT 'us-east-1'
);

CREATE TABLE IF NOT EXISTS objects (
    bucket        TEXT NOT NULL,
    key           TEXT NOT NULL,
    etag          TEXT NOT NULL,
    size          INTEGER NOT NULL,
    content_type  TEXT,
    last_modified TIMESTAMP NOT NULL,
    PRIMARY KEY (bucket, key),
    FOREIGN KEY(bucket) REFERENCES buckets(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_objects_bucket_prefix ON objects(bucket, key);

CREATE TABLE IF NOT EXISTS multipart_uploads (
    upload_id    TEXT PRIMARY KEY,
    bucket       TEXT NOT NULL,
    key          TEXT NOT NULL,
    initiated    TIMESTAMP NOT NULL,
    content_type TEXT
);

CREATE TABLE IF NOT EXISTS multipart_parts (
    upload_id   TEXT NOT NULL,
    part_number INTEGER NOT NULL,
    etag        TEXT NOT NULL,
    size        INTEGER NOT NULL,
    PRIMARY KEY (upload_id, part_number),
    FOREIGN KEY(upload_id) REFERENCES multipart_uploads(upload_id) ON DELETE CASCADE
);
