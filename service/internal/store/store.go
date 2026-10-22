// Package store provides SQLite persistence and query access for Open-Skill.
//
// Data-access is kept thin: rows map directly to the model types. Rating
// calculations live in the rating package and are applied atomically here.
package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// Store wraps the SQLite database handle.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the SQLite database at path and applies the
// schema.
func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return open(db)
}

// OpenRemote opens a remote libSQL database (e.g. Turso, libsql://...) and
// applies the schema.
func OpenRemote(url, authToken string) (*Store, error) {
	connector, err := libsql.NewConnector(url, libsql.WithAuthToken(authToken))
	if err != nil {
		return nil, fmt.Errorf("open libsql: %w", err)
	}
	db := sql.OpenDB(foreignKeysConnector{connector})
	// Turso drops idle HTTP connections; recycle before that happens.
	db.SetConnMaxIdleTime(20 * time.Second)
	return open(db)
}

func open(db *sql.DB) (*Store, error) {
	db.SetMaxOpenConns(1) // single writer for sqlite
	if err := applySchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// foreignKeysConnector enables foreign keys on each connection, since libSQL
// URLs cannot carry pragmas.
type foreignKeysConnector struct{ driver.Connector }

func (c foreignKeysConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var err error
	// Retry: a reused keep-alive connection closed by the server surfaces as EOF.
	for attempt := 0; attempt < 3; attempt++ {
		var conn driver.Conn
		if conn, err = c.connect(ctx); err == nil {
			return conn, nil
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil, err
}

func (c foreignKeysConnector) connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	execer, ok := conn.(driver.ExecerContext)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("libsql connection does not support exec")
	}
	if _, err := execer.ExecContext(ctx, "PRAGMA foreign_keys = ON", nil); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	return conn, nil
}

// Close releases the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the raw handle for advanced queries/tests.
func (s *Store) DB() *sql.DB { return s.db }

// applySchema executes the DDL. Because the statements use IF NOT EXISTS, it is
// also safe to run on an already-initialized database (idempotent migrations).
func applySchema(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return migratePlayerOwnership(db)
}

func migratePlayerOwnership(db *sql.DB) error {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(players)`)
	if err != nil {
		return fmt.Errorf("inspect player ownership schema: %w", err)
	}
	needsMigration := false
	for rows.Next() {
		var columnID, notNull, primaryKey int
		var name, columnType string
		var defaultValue sql.NullString
		if err := rows.Scan(&columnID, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return err
		}
		if name == "user_id" && notNull == 1 {
			needsMigration = true
		}
	}
	readErr := rows.Err()
	rows.Close()
	if readErr != nil {
		return readErr
	}
	if !needsMigration {
		return nil
	}
	if _, err := db.ExecContext(ctx, `
		PRAGMA foreign_keys = OFF;
		BEGIN IMMEDIATE;
		CREATE TABLE players_migrated (
			id TEXT PRIMARY KEY,
			user_id TEXT REFERENCES users(id),
			handle TEXT NOT NULL UNIQUE,
			community_id TEXT REFERENCES communities(id),
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
		INSERT INTO players_migrated (id, user_id, handle, created_at) SELECT id, user_id, handle, created_at FROM players;
		DROP TABLE players;
		ALTER TABLE players_migrated RENAME TO players;
		CREATE TEMP TABLE player_migration_integrity (violations INTEGER CHECK (violations = 0));
		INSERT INTO player_migration_integrity SELECT COUNT(*) FROM pragma_foreign_key_check;
		DROP TABLE player_migration_integrity;
		COMMIT;
		PRAGMA foreign_keys = ON;`); err != nil {
		db.ExecContext(ctx, `ROLLBACK`)
		db.ExecContext(ctx, `PRAGMA foreign_keys = ON`)
		return fmt.Errorf("migrate player ownership: %w", err)
	}
	return nil
}

// Ping verifies connectivity.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
