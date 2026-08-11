package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Index struct {
	db     *sql.DB
	dbPath string
}

type IndexOptions struct {
	DataDir string
}

func NewIndex(opts IndexOptions) (*Index, error) {
	dbPath := filepath.Join(opts.DataDir, "index.db")

	if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := openSQLite(dbPath)
	if err != nil {
		return nil, err
	}

	idx := &Index{db: db, dbPath: dbPath}

	if err := idx.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return idx, nil
}

func OpenDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	return openSQLite(dbPath)
}

// modernc.org/sqlite takes pragmas as _pragma=name(value) DSN params, not mattn's
// _busy_timeout= form. SetMaxOpenConns(1) serializes every query through one
// connection, so SQLITE_BUSY can only arrive from outside this process - which is
// what busy_timeout covers.
func openSQLite(dbPath string) (*sql.DB, error) {
	dsn := "file:" + dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func (i *Index) Close() error {
	return i.db.Close()
}

func (i *Index) runMigrations() error {
	if _, err := i.db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := i.db.Query("SELECT name FROM _migrations")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan migration name: %w", err)
		}
		applied[name] = true
	}

	migrations := []string{
		"001_initial.sql",
		"002_bucket_public.sql",
		"003_bucket_group.sql",
		"004_bucket_pinned.sql",
	}

	for _, name := range migrations {
		if applied[name] {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		for _, stmt := range splitStatements(string(content)) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := i.db.Exec(stmt); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", name, err)
			}
		}

		if _, err := i.db.Exec(
			"INSERT INTO _migrations (name, applied_at) VALUES (?, ?)",
			name, time.Now().UTC(),
		); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}
	}

	return nil
}

func splitStatements(content string) []string {
	var statements []string
	var current strings.Builder

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		statements = append(statements, current.String())
	}

	return statements
}

func (i *Index) InsertBucket(info *BucketInfo) error {
	_, err := i.db.Exec(
		"INSERT OR REPLACE INTO buckets (name, created_at, region, public, bucket_group, pinned) VALUES (?, ?, ?, ?, ?, ?)",
		info.Name, info.CreatedAt.UTC(), info.Region, info.Public, info.Group, info.Pinned,
	)
	return err
}

func (i *Index) DeleteBucket(name string) error {
	_, err := i.db.Exec("DELETE FROM buckets WHERE name = ?", name)
	return err
}

func (i *Index) ListBuckets() ([]BucketInfo, error) {
	rows, err := i.db.Query("SELECT name, created_at, region, public, bucket_group, pinned FROM buckets ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []BucketInfo
	for rows.Next() {
		var b BucketInfo
		var createdAtStr string
		if err := rows.Scan(&b.Name, &createdAtStr, &b.Region, &b.Public, &b.Group, &b.Pinned); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		b.CreatedAt = createdAt
		buckets = append(buckets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buckets, nil
}

func (i *Index) BucketHasObjects(name string) (bool, error) {
	var count int
	err := i.db.QueryRow("SELECT COUNT(*) FROM objects WHERE bucket = ?", name).Scan(&count)
	return count > 0, err
}

func (i *Index) InsertObject(info *ObjectInfo) error {
	_, err := i.db.Exec(`
		INSERT OR REPLACE INTO objects (bucket, key, etag, size, content_type, last_modified)
		VALUES (?, ?, ?, ?, ?, ?)
	`, info.Bucket, info.Key, info.ETag, info.Size, info.ContentType, info.LastModified.UTC())
	return err
}

func (i *Index) DeleteObject(bucket, key string) error {
	_, err := i.db.Exec("DELETE FROM objects WHERE bucket = ? AND key = ?", bucket, key)
	return err
}

func (i *Index) GetObject(bucket, key string) (*ObjectInfo, error) {
	row := i.db.QueryRow(`
		SELECT bucket, key, etag, size, content_type, last_modified
		FROM objects WHERE bucket = ? AND key = ?
	`, bucket, key)

	var info ObjectInfo
	var lastModifiedStr string
	var contentType sql.NullString
	if err := row.Scan(&info.Bucket, &info.Key, &info.ETag, &info.Size, &contentType, &lastModifiedStr); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	info.ContentType = contentType.String
	lastModified, err := time.Parse(time.RFC3339, lastModifiedStr)
	if err != nil {
		return nil, err
	}
	info.LastModified = lastModified

	return &info, nil
}

// escapeLike escapes LIKE wildcards so a user prefix matches literally.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (i *Index) ListObjects(bucket string, opts ListOpts) (*ListResult, error) {
	query := "SELECT key, etag, size, content_type, last_modified FROM objects WHERE bucket = ?"
	args := []any{bucket}

	if opts.Prefix != "" {
		query += ` AND key LIKE ? ESCAPE '\'`
		args = append(args, escapeLike(opts.Prefix)+"%")
	}

	if opts.StartAfter != "" {
		query += " AND key > ?"
		args = append(args, opts.StartAfter)
	}

	if opts.ContinuationToken != "" {
		query += " AND key > ?"
		args = append(args, opts.ContinuationToken)
	}

	query += " ORDER BY key ASC"

	limit := opts.MaxKeys
	if limit <= 0 {
		limit = 1000
	}
	query += " LIMIT ?"
	args = append(args, limit+1)

	rows, err := i.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []ObjectInfo
	for rows.Next() {
		var obj ObjectInfo
		var lastModifiedStr string
		var contentType sql.NullString
		if err := rows.Scan(&obj.Key, &obj.ETag, &obj.Size, &contentType, &lastModifiedStr); err != nil {
			return nil, err
		}
		obj.Bucket = bucket
		obj.ContentType = contentType.String
		lastModified, err := time.Parse(time.RFC3339, lastModifiedStr)
		if err != nil {
			return nil, err
		}
		obj.LastModified = lastModified
		objects = append(objects, obj)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := &ListResult{
		KeyCount: len(objects),
	}

	if len(objects) > limit {
		result.IsTruncated = true
		result.NextContinuationToken = objects[limit-1].Key
		objects = objects[:limit]
		result.KeyCount = limit
	}

	result.Contents = objects
	return result, nil
}

func (i *Index) InsertMultipartUpload(uploadID, bucket, key, contentType string) error {
	_, err := i.db.Exec(`
		INSERT INTO multipart_uploads (upload_id, bucket, key, initiated, content_type)
		VALUES (?, ?, ?, ?, ?)
	`, uploadID, bucket, key, time.Now().UTC(), nullString(contentType))
	return err
}

func (i *Index) GetMultipartUpload(uploadID string) (*UploadInfo, error) {
	row := i.db.QueryRow(`
		SELECT upload_id, bucket, key, initiated, content_type
		FROM multipart_uploads WHERE upload_id = ?
	`, uploadID)

	var info UploadInfo
	var initiatedStr string
	var contentType sql.NullString
	if err := row.Scan(&info.UploadID, &info.Bucket, &info.Key, &initiatedStr, &contentType); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUploadNotFound
		}
		return nil, err
	}

	info.ContentType = contentType.String
	initiated, err := time.Parse(time.RFC3339, initiatedStr)
	if err != nil {
		return nil, err
	}
	info.Initiated = initiated

	return &info, nil
}

func (i *Index) DeleteMultipartUpload(uploadID string) error {
	_, err := i.db.Exec("DELETE FROM multipart_uploads WHERE upload_id = ?", uploadID)
	return err
}

func (i *Index) InsertPart(uploadID string, partNumber int, etag string, size int64) error {
	_, err := i.db.Exec(`
		INSERT OR REPLACE INTO multipart_parts (upload_id, part_number, etag, size)
		VALUES (?, ?, ?, ?)
	`, uploadID, partNumber, etag, size)
	return err
}

func (i *Index) GetPart(uploadID string, partNumber int) (*PartInfo, error) {
	row := i.db.QueryRow(`
		SELECT part_number, etag, size
		FROM multipart_parts WHERE upload_id = ? AND part_number = ?
	`, uploadID, partNumber)

	var info PartInfo
	if err := row.Scan(&info.PartNumber, &info.ETag, &info.Size); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPartNotFound
		}
		return nil, err
	}

	return &info, nil
}

func (i *Index) ListParts(uploadID string) ([]PartInfo, error) {
	rows, err := i.db.Query(`
		SELECT part_number, etag, size
		FROM multipart_parts WHERE upload_id = ?
		ORDER BY part_number ASC
	`, uploadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []PartInfo
	for rows.Next() {
		var p PartInfo
		if err := rows.Scan(&p.PartNumber, &p.ETag, &p.Size); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

func (i *Index) DeleteParts(uploadID string) error {
	_, err := i.db.Exec("DELETE FROM multipart_parts WHERE upload_id = ?", uploadID)
	return err
}

func (i *Index) ListMultipartUploads(bucket string) ([]UploadInfo, error) {
	rows, err := i.db.Query(`
		SELECT upload_id, bucket, key, initiated, content_type
		FROM multipart_uploads WHERE bucket = ?
		ORDER BY initiated DESC
	`, bucket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []UploadInfo
	for rows.Next() {
		var u UploadInfo
		var initiatedStr string
		var contentType sql.NullString
		if err := rows.Scan(&u.UploadID, &u.Bucket, &u.Key, &initiatedStr, &contentType); err != nil {
			return nil, err
		}
		u.ContentType = contentType.String
		initiated, err := time.Parse(time.RFC3339, initiatedStr)
		if err != nil {
			return nil, err
		}
		u.Initiated = initiated
		uploads = append(uploads, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uploads, nil
}

func (i *Index) CountMultipartUploads() (int, error) {
	var n int
	err := i.db.QueryRow("SELECT COUNT(*) FROM multipart_uploads").Scan(&n)
	return n, err
}

func (i *Index) BucketStats(bucket string) (objectCount int64, totalSize int64, err error) {
	err = i.db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(size), 0) FROM objects WHERE bucket = ?",
		bucket,
	).Scan(&objectCount, &totalSize)
	return
}

func (i *Index) ServerStats() (buckets int, objects int64, totalSize int64, err error) {
	err = i.db.QueryRow("SELECT COUNT(*) FROM buckets").Scan(&buckets)
	if err != nil {
		return
	}
	err = i.db.QueryRow("SELECT COUNT(*), COALESCE(SUM(size), 0) FROM objects").Scan(&objects, &totalSize)
	return
}

type BucketStatsRow struct {
	BucketInfo
	ObjectCount int64
	TotalSize   int64
}

func (i *Index) ListBucketsWithStats() ([]BucketStatsRow, error) {
	query := `
		SELECT b.name, b.created_at, b.region, b.public, b.bucket_group, b.pinned,
		       COALESCE(o.object_count, 0) as object_count,
		       COALESCE(o.total_size, 0) as total_size
		FROM buckets b
		LEFT JOIN (
			SELECT bucket, COUNT(*) as object_count, SUM(size) as total_size
			FROM objects GROUP BY bucket
		) o ON b.name = o.bucket
		ORDER BY b.name
	`
	rows, err := i.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BucketStatsRow
	for rows.Next() {
		var r BucketStatsRow
		var createdAtStr string
		if err := rows.Scan(&r.Name, &createdAtStr, &r.Region, &r.Public, &r.Group, &r.Pinned, &r.ObjectCount, &r.TotalSize); err != nil {
			return nil, err
		}
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		r.CreatedAt = createdAt
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (i *Index) UpdateBucketPublic(name string, public bool) error {
	_, err := i.db.Exec(
		"UPDATE buckets SET public = ? WHERE name = ?",
		public, name,
	)
	return err
}

func (i *Index) UpdateBucketGroup(name, group string) error {
	_, err := i.db.Exec(
		"UPDATE buckets SET bucket_group = ? WHERE name = ?",
		group, name,
	)
	return err
}

func (i *Index) UpdateBucketPinned(name string, pinned bool) error {
	_, err := i.db.Exec("UPDATE buckets SET pinned = ? WHERE name = ?", pinned, name)
	return err
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

//go:embed migrations/*.sql
var migrationsFS embed.FS
