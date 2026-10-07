package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"open-skill/service/internal/model"
)

// EnsureUser creates the user and a primary player on first sign-in. Handles are
// unique, so a taken handle gets a random numeric suffix.
func (s *Store) EnsureUser(ctx context.Context, userID, playerID, handle string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO users (id, display) VALUES (?, ?)`, userID, handle); err != nil {
		return err
	}
	candidate := handle
	for attempt := 0; ; attempt++ {
		// Only create a player when the user has none, so existing handles are kept.
		result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO players (id, user_id, handle)
			SELECT ?, ?, ? WHERE NOT EXISTS (SELECT 1 FROM players WHERE user_id = ?)`,
			playerID, userID, candidate, userID)
		if err != nil {
			return err
		}
		inserted, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if inserted == 1 {
			break
		}
		var hasPlayer bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM players WHERE user_id = ?)`, userID).Scan(&hasPlayer); err != nil {
			return err
		}
		if hasPlayer {
			break
		}
		if attempt == 5 {
			return fmt.Errorf("could not allocate a unique handle for %q", handle)
		}
		candidate = fmt.Sprintf("%s-%d", handle, 1000+rand.IntN(9000))
	}
	return tx.Commit()
}

// ErrHandleTaken is returned when another player already uses the handle.
var ErrHandleTaken = errors.New("handle already in use")

// RenamePlayer changes a player's handle; only the owning user may do so.
func (s *Store) RenamePlayer(ctx context.Context, actorID, playerID, handle string) (*model.Player, error) {
	var owner string
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM players WHERE id = ?`, playerID).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if owner != actorID {
		return nil, ErrForbidden
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE players SET handle = ? WHERE id = ?`, handle, playerID); err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return nil, ErrHandleTaken
		}
		return nil, err
	}
	return &model.Player{ID: playerID, UserID: owner, Handle: handle}, nil
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
