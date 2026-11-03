package storage

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var validIndexNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// DB wraps the SQLite database handle and provides thread-safe access.
type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

// Open initializes SQLite database at path with high-throughput PRAGMAs.
func Open(path string) (*DB, error) {
	connStr := path
	if !strings.Contains(path, "?") {
		connStr += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
	}

	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite single writer; serialized writes ensure WAL consistency
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	sdb := &DB{db: db}
	if err := sdb.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return sdb, nil
}

// Close closes the underlying database.
func (d *DB) Close() error {
	return d.db.Close()
}

// initSchema creates base tables if they don't exist.
func (d *DB) initSchema() error {
	queries := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA temp_store = MEMORY;`,
		`PRAGMA cache_size = -64000;`,

		`CREATE TABLE IF NOT EXISTS sc_indexes (
			name TEXT PRIMARY KEY,
			created_at DATETIME NOT NULL,
			field_boosts TEXT NOT NULL,
			default_mode TEXT NOT NULL,
			embedding_provider TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS sc_documents (
			index_name TEXT NOT NULL,
			doc_id TEXT NOT NULL,
			data TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (index_name, doc_id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_docs_updated ON sc_documents(index_name, updated_at DESC);`,

		`CREATE TABLE IF NOT EXISTS sc_vectors (
			index_name TEXT NOT NULL,
			doc_id TEXT NOT NULL,
			dim INTEGER NOT NULL,
			vector BLOB NOT NULL,
			PRIMARY KEY (index_name, doc_id)
		);`,

		`CREATE TABLE IF NOT EXISTS sc_search_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			index_name TEXT NOT NULL,
			query TEXT NOT NULL,
			mode TEXT NOT NULL,
			total_hits INTEGER NOT NULL,
			took_ms REAL NOT NULL,
			created_at DATETIME NOT NULL
		);`,

		`CREATE INDEX IF NOT EXISTS idx_logs_created ON sc_search_logs(created_at DESC);`,

		`CREATE TABLE IF NOT EXISTS sc_zero_results (
			index_name TEXT NOT NULL,
			query TEXT NOT NULL,
			count INTEGER NOT NULL DEFAULT 1,
			last_searched DATETIME NOT NULL,
			PRIMARY KEY (index_name, query)
		);`,
	}

	for _, q := range queries {
		if _, err := d.db.Exec(q); err != nil {
			return fmt.Errorf("exec query (%s): %w", q, err)
		}
	}

	return nil
}

// sanitizeIndexName ensures the index name is valid and safe for table identifiers.
func sanitizeIndexName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("index name cannot be empty")
	}
	if !validIndexNameRegex.MatchString(trimmed) {
		return "", fmt.Errorf("index name contains invalid characters: %s", trimmed)
	}
	return strings.ReplaceAll(trimmed, "-", "_"), nil
}

// ftsTableName returns the sanitized FTS5 table name for an index.
func ftsTableName(cleanName string) string {
	return fmt.Sprintf("sc_fts_%s", cleanName)
}

// CreateIndex registers a new index and sets up its FTS5 table.
func (d *DB) CreateIndex(name string, fieldBoosts map[string]float64, defaultMode, embeddingProvider string) error {
	cleanName, err := sanitizeIndexName(name)
	if err != nil {
		return err
	}

	if defaultMode == "" {
		defaultMode = "hybrid"
	}
	if embeddingProvider == "" {
		embeddingProvider = "fast_local"
	}
	if fieldBoosts == nil {
		fieldBoosts = map[string]float64{
			"title":       3.0,
			"content":     1.0,
			"description": 1.5,
			"tags":        2.0,
			"category":    1.5,
			"brand":       1.5,
		}
	}

	boostsJSON, err := json.Marshal(fieldBoosts)
	if err != nil {
		return fmt.Errorf("marshal field boosts: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	_, err = tx.Exec(`
		INSERT INTO sc_indexes (name, created_at, field_boosts, default_mode, embedding_provider)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			field_boosts = excluded.field_boosts,
			default_mode = excluded.default_mode,
			embedding_provider = excluded.embedding_provider
	`, cleanName, now, string(boostsJSON), defaultMode, embeddingProvider)
	if err != nil {
		return fmt.Errorf("insert sc_indexes: %w", err)
	}

	// Create FTS5 virtual table for full-text search with porter stemmer
	ftsTable := ftsTableName(cleanName)
	createFTS := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS %s USING fts5(
			doc_id UNINDEXED,
			title,
			content,
			description,
			tags,
			category,
			brand,
			tokenize = 'porter unicode61 remove_diacritics 2'
		);
	`, ftsTable)

	if _, err := tx.Exec(createFTS); err != nil {
		return fmt.Errorf("create fts5 table: %w", err)
	}

	return tx.Commit()
}

// GetIndex retrieves configuration and document count for an index.
func (d *DB) GetIndex(name string) (*IndexMeta, error) {
	cleanName, err := sanitizeIndexName(name)
	if err != nil {
		return nil, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	var meta IndexMeta
	var boostsRaw string
	err = d.db.QueryRow(`
		SELECT name, created_at, field_boosts, default_mode, embedding_provider
		FROM sc_indexes WHERE name = ?
	`, cleanName).Scan(&meta.Name, &meta.CreatedAt, &boostsRaw, &meta.DefaultMode, &meta.EmbeddingProvider)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("index not found: %s", cleanName)
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(boostsRaw), &meta.FieldBoosts)

	err = d.db.QueryRow(`SELECT COUNT(*) FROM sc_documents WHERE index_name = ?`, cleanName).Scan(&meta.DocumentCount)
	if err != nil {
		meta.DocumentCount = 0
	}

	return &meta, nil
}

// ListIndexes returns metadata for all search indexes.
func (d *DB) ListIndexes() ([]IndexMeta, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT i.name, i.created_at, i.field_boosts, i.default_mode, i.embedding_provider,
		       COALESCE(COUNT(d.doc_id), 0) AS doc_count
		FROM sc_indexes i
		LEFT JOIN sc_documents d ON i.name = d.index_name
		GROUP BY i.name
		ORDER BY i.created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []IndexMeta
	for rows.Next() {
		var meta IndexMeta
		var boostsRaw string
		if err := rows.Scan(&meta.Name, &meta.CreatedAt, &boostsRaw, &meta.DefaultMode, &meta.EmbeddingProvider, &meta.DocumentCount); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(boostsRaw), &meta.FieldBoosts)
		result = append(result, meta)
	}

	return result, nil
}

// DeleteIndex drops the index, its document rows, vectors, and FTS5 table.
func (d *DB) DeleteIndex(name string) error {
	cleanName, err := sanitizeIndexName(name)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.Exec(`DELETE FROM sc_indexes WHERE name = ?`, cleanName)
	_, _ = tx.Exec(`DELETE FROM sc_documents WHERE index_name = ?`, cleanName)
	_, _ = tx.Exec(`DELETE FROM sc_vectors WHERE index_name = ?`, cleanName)
	_, _ = tx.Exec(`DELETE FROM sc_zero_results WHERE index_name = ?`, cleanName)

	ftsTable := ftsTableName(cleanName)
	_, _ = tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, ftsTable))

	return tx.Commit()
}

// UpsertDocuments inserts or updates a slice of documents into both SQLite and FTS5.
func (d *DB) UpsertDocuments(indexName string, docs []Document) error {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ftsTable := ftsTableName(cleanName)
	now := time.Now().UTC()

	stmtDoc, err := tx.Prepare(`
		INSERT INTO sc_documents (index_name, doc_id, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(index_name, doc_id) DO UPDATE SET
			data = excluded.data,
			updated_at = excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("prepare doc stmt: %w", err)
	}
	defer stmtDoc.Close()

	deleteFTSStmt, err := tx.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE doc_id = ?`, ftsTable))
	if err != nil {
		return fmt.Errorf("prepare fts delete: %w", err)
	}
	defer deleteFTSStmt.Close()

	insertFTSStmt, err := tx.Prepare(fmt.Sprintf(`
		INSERT INTO %s (doc_id, title, content, description, tags, category, brand)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, ftsTable))
	if err != nil {
		return fmt.Errorf("prepare fts insert: %w", err)
	}
	defer insertFTSStmt.Close()

	for _, doc := range docs {
		rawJSON, err := json.Marshal(doc.Data)
		if err != nil {
			return fmt.Errorf("marshal doc %s: %w", doc.ID, err)
		}

		if _, err := stmtDoc.Exec(cleanName, doc.ID, string(rawJSON), now, now); err != nil {
			return fmt.Errorf("exec doc insert %s: %w", doc.ID, err)
		}

		// Delete existing FTS entry if updating
		_, _ = deleteFTSStmt.Exec(doc.ID)

		// Extract searchable textual fields from JSON data
		title := extractString(doc.Data, "title", "name")
		content := extractString(doc.Data, "content", "body", "text")
		description := extractString(doc.Data, "description", "summary")
		tags := extractTags(doc.Data, "tags", "keywords")
		category := extractString(doc.Data, "category", "type")
		brand := extractString(doc.Data, "brand", "vendor")

		if _, err := insertFTSStmt.Exec(doc.ID, title, content, description, tags, category, brand); err != nil {
			return fmt.Errorf("exec fts insert %s: %w", doc.ID, err)
		}
	}

	return tx.Commit()
}

// GetDocument returns a single document by ID.
func (d *DB) GetDocument(indexName, docID string) (*Document, error) {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return nil, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	var dataRaw string
	var createdAt, updatedAt time.Time

	err = d.db.QueryRow(`
		SELECT data, created_at, updated_at
		FROM sc_documents
		WHERE index_name = ? AND doc_id = ?
	`, cleanName, docID).Scan(&dataRaw, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", docID)
		}
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(dataRaw), &data); err != nil {
		return nil, fmt.Errorf("unmarshal doc json: %w", err)
	}

	return &Document{
		ID:        docID,
		Index:     cleanName,
		Data:      data,
		RawJSON:   dataRaw,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// ListDocuments returns paginated documents for an index.
func (d *DB) ListDocuments(indexName string, limit, offset int) ([]Document, int64, error) {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	var total int64
	_ = d.db.QueryRow(`SELECT COUNT(*) FROM sc_documents WHERE index_name = ?`, cleanName).Scan(&total)

	rows, err := d.db.Query(`
		SELECT doc_id, data, created_at, updated_at
		FROM sc_documents
		WHERE index_name = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, cleanName, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var docID, dataRaw string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&docID, &dataRaw, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}

		var data map[string]any
		_ = json.Unmarshal([]byte(dataRaw), &data)

		docs = append(docs, Document{
			ID:        docID,
			Index:     cleanName,
			Data:      data,
			RawJSON:   dataRaw,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return docs, total, nil
}

// DeleteDocument removes a document from SQLite, FTS5, and vector tables.
func (d *DB) DeleteDocument(indexName, docID string) error {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.Exec(`DELETE FROM sc_documents WHERE index_name = ? AND doc_id = ?`, cleanName, docID)
	_, _ = tx.Exec(`DELETE FROM sc_vectors WHERE index_name = ? AND doc_id = ?`, cleanName, docID)

	ftsTable := ftsTableName(cleanName)
	_, _ = tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE doc_id = ?`, ftsTable), docID)

	return tx.Commit()
}

// SaveVectors serializes and stores float32 vectors as binary BLOBs.
func (d *DB) SaveVectors(indexName string, vectors map[string][]float32) error {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return err
	}

	if len(vectors) == 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO sc_vectors (index_name, doc_id, dim, vector)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(index_name, doc_id) DO UPDATE SET
			dim = excluded.dim,
			vector = excluded.vector
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for docID, vec := range vectors {
		dim := len(vec)
		blob := packFloat32(vec)
		if _, err := stmt.Exec(cleanName, docID, dim, blob); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadVectors loads all stored float32 vectors for an index.
func (d *DB) LoadVectors(indexName string) (map[string][]float32, error) {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return nil, err
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT doc_id, dim, vector
		FROM sc_vectors
		WHERE index_name = ?
	`, cleanName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string][]float32)
	for rows.Next() {
		var docID string
		var dim int
		var blob []byte
		if err := rows.Scan(&docID, &dim, &blob); err != nil {
			return nil, err
		}
		vec := unpackFloat32(blob, dim)
		res[docID] = vec
	}

	return res, nil
}

// QueryFTS executes an FTS5 full-text query with BM25 ranking.
type FTSResult struct {
	DocID     string
	BM25Score float64 // normalized positive score (higher is better)
}

func (d *DB) QueryFTS(indexName, matchQuery string, limit int) ([]FTSResult, error) {
	cleanName, err := sanitizeIndexName(indexName)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	ftsTable := ftsTableName(cleanName)
	d.mu.RLock()
	defer d.mu.RUnlock()

	// SQLite bm25 returns negative score where more negative is better rank.
	// We negate it so higher positive value = higher relevance.
	query := fmt.Sprintf(`
		SELECT doc_id, -bm25(%s, 3.0, 1.0, 1.5, 2.0, 1.5, 1.5) AS score
		FROM %s
		WHERE %s MATCH ?
		ORDER BY score DESC
		LIMIT ?
	`, ftsTable, ftsTable, ftsTable)

	rows, err := d.db.Query(query, matchQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FTSResult
	for rows.Next() {
		var item FTSResult
		if err := rows.Scan(&item.DocID, &item.BM25Score); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, nil
}

// RecordSearch logs query execution stats and updates zero-result counts if hits == 0.
func (d *DB) RecordSearch(log SearchLog) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	_, _ = d.db.Exec(`
		INSERT INTO sc_search_logs (index_name, query, mode, total_hits, took_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, log.IndexName, log.Query, log.Mode, log.TotalHits, log.TookMs, now)

	if log.TotalHits == 0 && strings.TrimSpace(log.Query) != "" {
		_, _ = d.db.Exec(`
			INSERT INTO sc_zero_results (index_name, query, count, last_searched)
			VALUES (?, ?, 1, ?)
			ON CONFLICT(index_name, query) DO UPDATE SET
				count = count + 1,
				last_searched = excluded.last_searched
		`, log.IndexName, strings.ToLower(strings.TrimSpace(log.Query)), now)
	}
}

// GetAnalyticsOverview computes total queries, latency percentiles, and zero-result rates.
func (d *DB) GetAnalyticsOverview() (*AnalyticsOverview, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	overview := &AnalyticsOverview{}

	var totalSearches, zeroHits int64
	var avgLatency sql.NullFloat64

	_ = d.db.QueryRow(`SELECT COUNT(*), COALESCE(AVG(took_ms), 0) FROM sc_search_logs`).Scan(&totalSearches, &avgLatency)
	_ = d.db.QueryRow(`SELECT COUNT(*) FROM sc_search_logs WHERE total_hits = 0`).Scan(&zeroHits)

	overview.TotalSearches = totalSearches
	overview.ZeroResultSearches = zeroHits
	if totalSearches > 0 {
		overview.ZeroResultRate = float64(zeroHits) / float64(totalSearches)
		overview.AvgLatencyMs = math.Round(avgLatency.Float64*100) / 100
	}

	// Calculate percentiles
	rows, err := d.db.Query(`SELECT took_ms FROM sc_search_logs ORDER BY took_ms ASC`)
	if err == nil {
		defer rows.Close()
		var latencies []float64
		for rows.Next() {
			var l float64
			if err := rows.Scan(&l); err == nil {
				latencies = append(latencies, l)
			}
		}
		if len(latencies) > 0 {
			overview.P50LatencyMs = math.Round(percentile(latencies, 0.50)*100) / 100
			overview.P95LatencyMs = math.Round(percentile(latencies, 0.95)*100) / 100
			overview.P99LatencyMs = math.Round(percentile(latencies, 0.99)*100) / 100
		}
	}

	return overview, nil
}

// GetPopularQueries returns the top searched query terms.
func (d *DB) GetPopularQueries(limit int) ([]PopularQuery, error) {
	if limit <= 0 {
		limit = 10
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT query, COUNT(*) as cnt, AVG(total_hits) as avg_hits
		FROM sc_search_logs
		WHERE TRIM(query) != ''
		GROUP BY LOWER(TRIM(query))
		ORDER BY cnt DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PopularQuery
	for rows.Next() {
		var item PopularQuery
		if err := rows.Scan(&item.Query, &item.Count, &item.AvgHits); err != nil {
			return nil, err
		}
		item.AvgHits = math.Round(item.AvgHits*10) / 10
		list = append(list, item)
	}

	return list, nil
}

// GetZeroResults returns terms that produced zero hits, ordered by frequency.
func (d *DB) GetZeroResults(limit int) ([]ZeroResultQuery, error) {
	if limit <= 0 {
		limit = 20
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(`
		SELECT query, count, last_searched
		FROM sc_zero_results
		ORDER BY count DESC, last_searched DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ZeroResultQuery
	for rows.Next() {
		var item ZeroResultQuery
		if err := rows.Scan(&item.Query, &item.Count, &item.LastSearched); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, nil
}

// Helper functions for packing/unpacking float32 vectors.
func packFloat32(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, f := range vec {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], bits)
	}
	return buf
}

func unpackFloat32(buf []byte, dim int) []float32 {
	if len(buf) < dim*4 {
		return make([]float32, dim)
	}
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		bits := binary.LittleEndian.Uint32(buf[i*4 : (i+1)*4])
		vec[i] = math.Float32frombits(bits)
	}
	return vec
}

func percentile(sortedVals []float64, p float64) float64 {
	if len(sortedVals) == 0 {
		return 0
	}
	if !sort.Float64sAreSorted(sortedVals) {
		sort.Float64s(sortedVals)
	}
	idx := int(math.Ceil(p*float64(len(sortedVals)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sortedVals) {
		idx = len(sortedVals) - 1
	}
	return sortedVals[idx]
}

func extractString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if val, ok := m[k]; ok && val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}

func extractTags(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if val, ok := m[k]; ok && val != nil {
			switch v := val.(type) {
			case []string:
				return strings.Join(v, " ")
			case []any:
				var strList []string
				for _, item := range v {
					strList = append(strList, fmt.Sprintf("%v", item))
				}
				return strings.Join(strList, " ")
			case string:
				return v
			}
		}
	}
	return ""
}
