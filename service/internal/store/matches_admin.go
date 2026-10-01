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
	rows := make([]participantRow, len(participants))
	for i, p := range participants {
		rows[i] = participantRow{matchID: matchID, p: p}
	}
	if err := writeParticipants(ctx, tx, rows); err != nil {
		return nil, err
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
// applied by rebuilding rather than by trying to invert a single result. The
// log is loaded and rated in memory so the cost does not grow with match count.
func recomputeCommunity(ctx context.Context, tx *sql.Tx, communityID string) error {
	defaults := rating.NewOpenSkillPlayer("")
	if _, err := tx.ExecContext(ctx, `
		UPDATE ratings SET mu = ?, sigma = ?, games = 0, wins = 0, losses = 0, draws = 0, last_played = NULL
		WHERE community_id = ?`, defaults.Mu, defaults.Sigma, communityID); err != nil {
		return fmt.Errorf("reset ratings: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT m.id, m.mode, m.played_at, mp.player_id, COALESCE(mp.team, ''), mp.score, COALESCE(mp.finish, -1)
		FROM matches m
		JOIN match_participants mp ON mp.match_id = m.id
		WHERE m.community_id = ?
		ORDER BY m.played_at, m.id, mp.player_id`, communityID)
	if err != nil {
		return fmt.Errorf("load community match log: %w", err)
	}
	type replayMatch struct {
		id           string
		mode         model.Mode
		playedAt     string
		participants []Participant
	}
	var replays []*replayMatch
	for rows.Next() {
		var id, playedAt string
		var mode model.Mode
		var p Participant
		if err := rows.Scan(&id, &mode, &playedAt, &p.PlayerID, &p.Team, &p.Score, &p.Finish); err != nil {
			rows.Close()
			return err
		}
		if len(replays) == 0 || replays[len(replays)-1].id != id {
			replays = append(replays, &replayMatch{id: id, mode: mode, playedAt: playedAt})
		}
		last := replays[len(replays)-1]
		last.participants = append(last.participants, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	book := ratingBook{}
	var participantRows []participantRow
	for _, replay := range replays {
		_, normalized, err := normalizeReport(replay.participants)
		if err != nil {
			return fmt.Errorf("replay match %s: %w", replay.id, err)
		}
		matchRows, _, err := book.rateMatch(replay.id, replay.mode, replay.participants, normalized, replay.playedAt)
		if err != nil {
			return fmt.Errorf("replay match %s: %w", replay.id, err)
		}
		participantRows = append(participantRows, matchRows...)
	}
	if err := writeParticipants(ctx, tx, participantRows); err != nil {
		return err
	}
	return writeRatings(ctx, tx, communityID, book)
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
