package storage

import (
"context"
"database/sql"
"encoding/json"
"errors"
"time"

_ "modernc.org/sqlite"

"github.com/zebdo/utsusu/internal/core"
)

type sqliteStore struct {
	db *sql.DB
}

func NewSQLite(path string) (Storage, error) {
	dsn := "file:" + path + "?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL"
	db, err := sql.Open("sqlite", dsn)
	if err != nil { return nil, err }
	st := &sqliteStore{ db: db }
	if err := st.init(context.Background()); err != nil { return nil, err }
	return st, nil
}

func (s *sqliteStore) init(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS threads (
		id TEXT PRIMARY KEY,
		board TEXT NOT NULL,
		sticky INTEGER NOT NULL DEFAULT 0,
		closed INTEGER NOT NULL DEFAULT 0,
		metadata TEXT,
		updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
		id TEXT NOT NULL,
		thread_id TEXT NOT NULL,
		author TEXT,
		content TEXT,
		ts DATETIME,
		images TEXT,
		metadata TEXT,
		PRIMARY KEY (id, thread_id),
		FOREIGN KEY(thread_id) REFERENCES threads(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil { return err }
	}
	return nil
}

func (s *sqliteStore) SaveThread(t core.Thread) error {
	ctx := context.Background()
	meta, _ := json.Marshal(t.Metadata)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO threads (id, board, sticky, closed, metadata, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET board=excluded.board, sticky=excluded.sticky, closed=excluded.closed, metadata=excluded.metadata, updated_at=excluded.updated_at`,
		t.ID, t.Board, b2i(t.Sticky), b2i(t.Closed), string(meta), time.Now(),
	)
	if err != nil { return err }

	// Delta update strategy:
	// - get existing post IDs for this thread
	// - insert posts that don't exist yet
	exist := map[string]bool{}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM posts WHERE thread_id=?`, t.ID)
	if err != nil { return err }
	defer rows.Close()
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil { return err }
		exist[pid] = true
	}

	ins, err := s.db.PrepareContext(ctx, `INSERT OR IGNORE INTO posts (id, thread_id, author, content, ts, images, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil { return err }
	defer ins.Close()

	for _, p := range t.Posts {
		if exist[p.ID] { continue } // skip existing
		imgs, _ := json.Marshal(p.Images)
		pm, _ := json.Marshal(p.Metadata)
		if _, err := ins.ExecContext(ctx, p.ID, t.ID, p.Author, p.Content, p.Timestamp, string(imgs), string(pm)); err != nil { return err }
	}

	return nil
}

func (s *sqliteStore) GetThread(id string) (*core.Thread, error) {
	ctx := context.Background()
	row := s.db.QueryRowContext(ctx, `SELECT id, board, sticky, closed, metadata FROM threads WHERE id=?`, id)
	var th core.Thread
	var meta string
	var sticky, closed int
	if err := row.Scan(&th.ID, &th.Board, &sticky, &closed, &meta); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
		return nil, err
	}
	th.Sticky = i2b(sticky); th.Closed = i2b(closed)
	_ = json.Unmarshal([]byte(meta), &th.Metadata)

	rows2, err := s.db.QueryContext(ctx, `SELECT id, author, content, ts, images, metadata FROM posts WHERE thread_id=? ORDER BY ts ASC`, id)
	if err != nil { return nil, err }
	defer rows2.Close()
	for rows2.Next() {
		var p core.Post
		var imgs, pm string
		if err := rows2.Scan(&p.ID, &p.Author, &p.Content, &p.Timestamp, &imgs, &pm); err != nil { return nil, err }
		_ = json.Unmarshal([]byte(imgs), &p.Images)
		_ = json.Unmarshal([]byte(pm), &p.Metadata)
		th.Posts = append(th.Posts, p)
	}
	return &th, nil
}

func (s *sqliteStore) ListThreads(board string) ([]core.Thread, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx, `SELECT id, board, sticky, closed, metadata FROM threads WHERE board=? ORDER BY updated_at DESC`, board)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []core.Thread
	for rows.Next() {
		var th core.Thread
		var sticky, closed int
		var meta string
		if err := rows.Scan(&th.ID, &th.Board, &sticky, &closed, &meta); err != nil { return nil, err }
		th.Sticky = i2b(sticky); th.Closed = i2b(closed)
		_ = json.Unmarshal([]byte(meta), &th.Metadata)
		out = append(out, th)
	}
	return out, nil
}

func b2i(b bool) int { if b { return 1 }; return 0 }
func i2b(i int) bool { return i != 0 }
