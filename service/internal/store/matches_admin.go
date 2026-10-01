package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"open-skill/service/internal/model"
	"open-skill/service/internal/rating"
)

// Match returns one match with its participants.
func (s *Store) Match(ctx context.Context, matchID string) (*MatchHistory, error) {
	var out MatchHistory
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, community_id, mode, played_at FROM matches WHERE id = ?`, matchID).
		Scan(&out.MatchID, &out.CommunityID, &out.Mode, &out.PlayedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("match %q: %w", matchID, ErrNotFound)
		}
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT mp.player_id, COALESCE(p.handle, mp.player_id), COALESCE(mp.team, ''), mp.score,
		       COALESCE(mp.finish, -1), mp.rating_delta
		FROM match_participants mp
		LEFT JOIN players p ON p.id = mp.player_id
		WHERE mp.match_id = ?
		ORDER BY mp.finish, mp.score DESC, mp.player_id`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p ParticipantHistory
		if err := rows.Scan(&p.PlayerID, &p.Handle, &p.Team, &p.Score, &p.Finish, &p.RatingDelta); err != nil {
			return nil, err
		}
		out.Participants = append(out.Participants, p)
	}
	return &out, rows.Err()
}

// DeleteMatch removes a match and rebuilds the community's ratings from the
// remaining match log. Only community owners and admins may do this.
func (s *Store) DeleteMatch(ctx context.Context, actorID, matchID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	communityID, err := matchCommunity(ctx, tx, matchID)
	if err != nil {
		return "", err
	}
	if err := requireCommunityAdmin(ctx, tx, communityID, actorID); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM match_participants WHERE match_id = ?`, matchID); err != nil {
		return "", fmt.Errorf("delete match participants: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM matches WHERE id = ?`, matchID); err != nil {
		return "", fmt.Errorf("delete match: %w", err)
	}
	if err := recomputeCommunity(ctx, tx, communityID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return communityID, nil
}

// UpdateMatch replaces a match's participants and result, then rebuilds the
// community's ratings so later matches account for the correction.
func (s *Store) UpdateMatch(ctx context.Context, actorID, matchID string, participants []Participant, notes string) (*ReportOutcome, error) {
	mode, _, err := normalizeReport(participants)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	communityID, err := matchCommunity(ctx, tx, matchID)
	if err != nil {
		return nil, err
	}
	if err := requireCommunityAdmin(ctx, tx, communityID, actorID); err != nil {
		return nil, err
	}
	if err := validateCommunityParticipants(ctx, tx, communityID, participants); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE matches SET mode = ?, notes = ? WHERE id = ?`, string(mode), notes, matchID); err != nil {
		return nil, fmt.Errorf("update match: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM match_participants WHERE match_id = ?`, matchID); err != nil {
		return nil, fmt.Errorf("clear match participants: %w", err)
	}
	for _, p := range participants {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO match_participants (match_id, player_id, team, score, finish, rating_delta)
			VALUES (?, ?, ?, ?, ?, 0)`, matchID, p.PlayerID, p.Team, p.Score, nullableFinish(p)); err != nil {
			return nil, fmt.Errorf("insert match participant: %w", err)
		}
	}
	if err := recomputeCommunity(ctx, tx, communityID); err != nil {
		return nil, err
	}
	deltas, err := matchDeltas(ctx, tx, communityID, matchID, mode)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &ReportOutcome{MatchID: matchID, Mode: mode, Deltas: deltas}, nil
}

func matchCommunity(ctx context.Context, tx *sql.Tx, matchID string) (string, error) {
	var communityID string
	if err := tx.QueryRowContext(ctx, `SELECT community_id FROM matches WHERE id = ?`, matchID).Scan(&communityID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("match %q: %w", matchID, ErrNotFound)
		}
		return "", err
	}
	return communityID, nil
}

func requireCommunityAdmin(ctx context.Context, tx *sql.Tx, communityID, actorID string) error {
	if actorID == "" {
		return nil
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM communities c WHERE c.id = ? AND c.owner_id = ?
			UNION ALL
			SELECT 1 FROM community_members cm
			 WHERE cm.community_id = ? AND cm.user_id = ? AND cm.role IN ('owner', 'admin')
		)`, communityID, actorID, communityID, actorID).Scan(&allowed); err != nil {
		return fmt.Errorf("check community admin: %w", err)
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

// recomputeCommunity replays every match in the community in chronological
// order. Ratings are a projection of the match log, so edits and deletions are
// applied by rebuilding rather than by trying to invert a single result.
func recomputeCommunity(ctx context.Context, tx *sql.Tx, communityID string) error {
	defaults := rating.NewOpenSkillPlayer("")
	if _, err := tx.ExecContext(ctx, `
		UPDATE ratings SET mu = ?, sigma = ?, games = 0, wins = 0, losses = 0, draws = 0, last_played = NULL
		WHERE community_id = ?`, defaults.Mu, defaults.Sigma, communityID); err != nil {
		return fmt.Errorf("reset ratings: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, mode, played_at FROM matches WHERE community_id = ? ORDER BY played_at, id`, communityID)
	if err != nil {
		return fmt.Errorf("list community matches: %w", err)
	}
	type replayMatch struct {
		id       string
		mode     model.Mode
		playedAt string
	}
	var replays []replayMatch
	for rows.Next() {
		var m replayMatch
		if err := rows.Scan(&m.id, &m.mode, &m.playedAt); err != nil {
			rows.Close()
			return err
		}
		replays = append(replays, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, replay := range replays {
		participants, err := matchParticipants(ctx, tx, replay.id)
		if err != nil {
			return err
		}
		_, normalized, err := normalizeReport(participants)
		if err != nil {
			return fmt.Errorf("replay match %s: %w", replay.id, err)
		}
		if _, err := applyMatchRatings(ctx, tx, communityID, replay.id, replay.mode, participants, normalized, replay.playedAt); err != nil {
			return fmt.Errorf("replay match %s: %w", replay.id, err)
		}
	}
	return nil
}

func matchParticipants(ctx context.Context, tx *sql.Tx, matchID string) ([]Participant, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT player_id, COALESCE(team, ''), score, COALESCE(finish, -1)
		FROM match_participants WHERE match_id = ? ORDER BY player_id`, matchID)
	if err != nil {
		return nil, fmt.Errorf("load match participants: %w", err)
	}
	defer rows.Close()
	out := []Participant{}
	for rows.Next() {
		var p Participant
		if err := rows.Scan(&p.PlayerID, &p.Team, &p.Score, &p.Finish); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func matchDeltas(ctx context.Context, tx *sql.Tx, communityID, matchID string, mode model.Mode) ([]model.RatingDelta, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT mp.player_id, mp.rating_delta, r.mu, r.sigma
		FROM match_participants mp
		LEFT JOIN ratings r ON r.community_id = ? AND r.player_id = mp.player_id AND r.mode = ?
		WHERE mp.match_id = ?`, communityID, string(mode), matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.RatingDelta{}
	for rows.Next() {
		var delta model.RatingDelta
		var mu, sigma sql.NullFloat64
		if err := rows.Scan(&delta.PlayerID, &delta.Delta, &mu, &sigma); err != nil {
			return nil, err
		}
		delta.Mode = mode
		delta.NewRating = rating.OpenSkillDisplayRating(mu.Float64, sigma.Float64)
		out = append(out, delta)
	}
	return out, rows.Err()
}
