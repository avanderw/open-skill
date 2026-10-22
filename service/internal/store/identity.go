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

func (s *Store) CreateUnclaimedPlayer(ctx context.Context, actorID, communityID, handle string) (*model.Player, error) {
	handle = strings.Join(strings.Fields(handle), " ")
	if handle == "" || len([]rune(handle)) > 64 {
		return nil, fmt.Errorf("name must be between 1 and 64 characters")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var isMember bool
	if err := tx.QueryRowContext(ctx, `SELECT owner_id = ? OR EXISTS (
		SELECT 1 FROM community_members WHERE community_id = c.id AND user_id = ?
	) FROM communities c WHERE id = ?`, actorID, actorID, communityID).Scan(&isMember); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if actorID == "" || !isMember {
		return nil, ErrForbidden
	}
	player := &model.Player{ID: newID("p"), Handle: handle, Unclaimed: true, CommunityID: communityID, Avatar: initials(handle)}
	if _, err := tx.ExecContext(ctx, `INSERT INTO players (id, handle, community_id) VALUES (?, ?, ?)`, player.ID, handle, communityID); err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			return nil, ErrHandleTaken
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return player, nil
}

// RenamePlayer changes a player's handle; only the owning user may do so.
func (s *Store) RenamePlayer(ctx context.Context, actorID, playerID, handle string) (*model.Player, error) {
	var owner string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(user_id, '') FROM players WHERE id = ?`, playerID).Scan(&owner)
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

type PlayerClaimRequest struct {
	UserID    string `json:"userId"`
	Display   string `json:"display"`
	CreatedAt string `json:"createdAt"`
}

type CommunityPlayerClaim struct {
	PlayerClaimRequest
	PlayerID       string `json:"playerId"`
	Handle         string `json:"handle"`
	TargetPlayerID string `json:"targetPlayerId"`
	TargetHandle   string `json:"targetHandle"`
}

func (s *Store) CommunityPlayerClaims(ctx context.Context, actorID, communityID string) ([]CommunityPlayerClaim, error) {
	if err := s.CanAdmin(ctx, communityID, actorID); err != nil {
		return nil, ErrForbidden
	}
	rows, err := s.db.QueryContext(ctx, `SELECT cr.user_id, u.display, cr.created_at, p.id, p.handle, target.id, target.handle
		FROM player_claim_requests cr
		JOIN players p ON p.id = cr.player_id
		JOIN users u ON u.id = cr.user_id
		JOIN players target ON target.id = (SELECT id FROM players WHERE user_id = cr.user_id ORDER BY created_at, id LIMIT 1)
		WHERE p.community_id = ? AND p.user_id IS NULL
		ORDER BY cr.created_at, p.handle, cr.user_id`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	claims := []CommunityPlayerClaim{}
	for rows.Next() {
		var claim CommunityPlayerClaim
		if err := rows.Scan(&claim.UserID, &claim.Display, &claim.CreatedAt, &claim.PlayerID, &claim.Handle, &claim.TargetPlayerID, &claim.TargetHandle); err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	return claims, rows.Err()
}

func (s *Store) RequestPlayerClaim(ctx context.Context, actorID, playerID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var communityID string
	var unclaimed bool
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(community_id, ''), user_id IS NULL FROM players WHERE id = ?`, playerID).Scan(&communityID, &unclaimed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !unclaimed {
		return ErrAlreadyClaimed
	}
	var isMember bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM community_members WHERE community_id = ? AND user_id = ?)`, communityID, actorID).Scan(&isMember); err != nil {
		return err
	}
	if !isMember || actorID == "" {
		return ErrForbidden
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO player_claim_requests (player_id, user_id) VALUES (?, ?)`, playerID, actorID); err != nil {
		return err
	}
	return tx.Commit()
}

var ErrAlreadyClaimed = errors.New("player is already claimed")

var ErrPlayerMergeConflict = errors.New("both players appear in the same match; correct those matches before approving the claim")

func (s *Store) PlayerClaims(ctx context.Context, actorID, playerID string) ([]PlayerClaimRequest, error) {
	player, _, err := s.Profile(ctx, playerID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.CommunityRole(ctx, player.CommunityID, actorID); err != nil {
		return nil, ErrForbidden
	}
	query := `SELECT cr.user_id, u.display, cr.created_at FROM player_claim_requests cr JOIN users u ON u.id = cr.user_id WHERE cr.player_id = ?`
	args := []any{playerID}
	if err := s.CanAdmin(ctx, player.CommunityID, actorID); err != nil {
		query += ` AND cr.user_id = ?`
		args = append(args, actorID)
	}
	rows, err := s.db.QueryContext(ctx, query+` ORDER BY cr.created_at, cr.user_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	claims := []PlayerClaimRequest{}
	for rows.Next() {
		var claim PlayerClaimRequest
		if err := rows.Scan(&claim.UserID, &claim.Display, &claim.CreatedAt); err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	return claims, rows.Err()
}

func (s *Store) ApprovePlayerClaim(ctx context.Context, actorID, playerID, userID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var communityID string
	var unclaimed bool
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(community_id, ''), user_id IS NULL FROM players WHERE id = ?`, playerID).Scan(&communityID, &unclaimed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if err := requireCommunityAdmin(ctx, tx, communityID, actorID); err != nil {
		return "", err
	}
	if actorID == "" {
		return "", ErrForbidden
	}
	if !unclaimed {
		return "", ErrAlreadyClaimed
	}
	var targetID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM players WHERE user_id = ?
		AND EXISTS(SELECT 1 FROM player_claim_requests WHERE player_id = ? AND user_id = ?)
		AND EXISTS(SELECT 1 FROM community_members WHERE community_id = ? AND user_id = ?)
		ORDER BY created_at, id LIMIT 1`, userID, playerID, userID, communityID, userID).Scan(&targetID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	var conflict bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM match_participants source JOIN match_participants target ON target.match_id = source.match_id
		WHERE source.player_id = ? AND target.player_id = ?)`, playerID, targetID).Scan(&conflict); err != nil {
		return "", err
	}
	if conflict {
		return "", ErrPlayerMergeConflict
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT m.community_id FROM matches m
		JOIN match_participants mp ON mp.match_id = m.id WHERE mp.player_id = ?`, playerID)
	if err != nil {
		return "", err
	}
	communities := []string{}
	for rows.Next() {
		var affectedCommunity string
		if err := rows.Scan(&affectedCommunity); err != nil {
			rows.Close()
			return "", err
		}
		communities = append(communities, affectedCommunity)
	}
	readErr := rows.Err()
	rows.Close()
	if readErr != nil {
		return "", readErr
	}
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`UPDATE match_participants SET player_id = ? WHERE player_id = ?`, []any{targetID, playerID}},
		{`DELETE FROM ratings WHERE player_id = ?`, []any{playerID}},
		{`DELETE FROM player_claim_requests WHERE player_id = ?`, []any{playerID}},
		{`INSERT INTO player_redirects (old_player_id, player_id) VALUES (?, ?)`, []any{playerID, targetID}},
		{`DELETE FROM players WHERE id = ?`, []any{playerID}},
	} {
		if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
			return "", err
		}
	}
	for _, affectedCommunity := range communities {
		if err := recomputeCommunity(ctx, tx, affectedCommunity); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return targetID, nil
}

func (s *Store) RejectPlayerClaim(ctx context.Context, actorID, playerID, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var communityID string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(community_id, '') FROM players WHERE id = ?`, playerID).Scan(&communityID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if actorID != userID {
		if err := requireCommunityAdmin(ctx, tx, communityID, actorID); err != nil {
			return err
		}
	}
	if actorID == "" {
		return ErrForbidden
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM player_claim_requests WHERE player_id = ? AND user_id = ?`, playerID, userID); err != nil {
		return err
	}
	return tx.Commit()
}
