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
	return open(sql.OpenDB(foreignKeysConnector{connector}))
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
	return nil
}

// Ping verifies connectivity.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
