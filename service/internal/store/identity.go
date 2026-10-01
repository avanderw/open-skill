package store

import (
	"context"
	"database/sql"
	"errors"

	"open-skill/service/internal/model"
)

func (s *Store) EnsureDemoIdentity(ctx context.Context, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO users (id, display) VALUES (?, ?)`, userID, userID); err != nil {
		return err
	}
	// Only create a player when the user has none, so existing handles are kept.
	if _, err := tx.ExecContext(ctx, `INSERT INTO players (id, user_id, handle)
		SELECT ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM players WHERE user_id = ?)`,
		userID, userID, userID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) PrimaryPlayerForUser(ctx context.Context, userID string) (*model.Player, error) {
	var player model.Player
	err := s.db.QueryRowContext(ctx, `SELECT id, user_id, handle FROM players WHERE user_id = ? ORDER BY created_at, id LIMIT 1`, userID).Scan(&player.ID, &player.UserID, &player.Handle)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &player, nil
}
