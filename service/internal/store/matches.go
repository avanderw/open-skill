package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"open-skill/service/internal/model"
	"open-skill/service/internal/rating"
)

var (
	ErrUnknownPlayer         = errors.New("unknown player")
	ErrIneligibleParticipant = errors.New("ineligible participant")
)

// ReportMatchRequest is what the API accepts for a reported match.
// Mode is DERIVED here from the participant configuration:
//   - Each participant has a Team label. No teams => 1v1.
//   - Exactly one shared team label across all participants => single team (not
//     a rating change for a match with one side) — treated as 1vq.
//   - More than one team => team mode.
//
// finishRank (optional) disambiguates FFA finishes. For team/1v1 matches, a
// winner is determined from Score (higher wins; equal => draw if draws enabled).
type ReportMatchRequest struct {
	CommunityID  string        `json:"communityId"`
	Participants []Participant `json:"participants"`
	Notes        string        `json:"notes"`
	PlayedAt     string        `json:"playedAt,omitempty"` // RFC3339; default now
}

// contestant is a participant plus its current rating for recomputation.
type contestant struct {
	req       Participant
	mu, sigma float64
}

// Participant is one competitor in a reported match.
type Participant struct {
	PlayerID string  `json:"playerId"`
	Team     string  `json:"team"`   // team label; different => team mode
	Score    float64 `json:"score"`  // points scored
	Finish   int     `json:"finish"` // 0-based finish; -1 if N/A
}

// ReportMatch persists a match and applies OpenSkill rating updates atomically.
// The mode is derived from the participant/team config (the observable "projection").
func (s *Store) ReportMatch(ctx context.Context, req ReportMatchRequest) (*ReportOutcome, error) {
	return s.reportMatch(ctx, "", req)
}

// ReportMatchAsMember reports a match while checking the actor's membership
// inside the same transaction that writes the match and ratings.
func (s *Store) ReportMatchAsMember(ctx context.Context, actorID string, req ReportMatchRequest) (*ReportOutcome, error) {
	return s.reportMatch(ctx, actorID, req)
}

func (s *Store) reportMatch(ctx context.Context, actorID string, req ReportMatchRequest) (*ReportOutcome, error) {
	if req.CommunityID == "" {
		return nil, fmt.Errorf("communityId is required")
	}
	mode, normalized, err := normalizeReport(req.Participants)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var drawsAllowed bool
	if err := tx.QueryRowContext(ctx, `SELECT draws FROM communities WHERE id = ?`, req.CommunityID).Scan(&drawsAllowed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("community %q: %w", req.CommunityID, ErrNotFound)
		}
		return nil, fmt.Errorf("load community %q: %w", req.CommunityID, err)
	}
	if actorID != "" {
		var isMember bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM communities c WHERE c.id = ? AND c.owner_id = ?
				UNION ALL
				SELECT 1 FROM community_members cm WHERE cm.community_id = ? AND cm.user_id = ?
			)`, req.CommunityID, actorID, req.CommunityID, actorID).Scan(&isMember); err != nil {
			return nil, fmt.Errorf("check match reporter membership: %w", err)
		}
		if !isMember {
			return nil, ErrForbidden
		}
	}
	if err := validateCommunityParticipants(ctx, tx, req.CommunityID, req.Participants); err != nil {
		return nil, err
	}
	if !drawsAllowed && mode != model.ModeFFA && normalized.Ranks[0] == normalized.Ranks[1] {
		return nil, fmt.Errorf("draws are not allowed in this community")
	}

	matchID := newID("m")
	var playedAt string
	if req.PlayedAt != "" {
		playedAt = req.PlayedAt
	} else {
		if err := tx.QueryRowContext(ctx, `SELECT datetime('now')`).Scan(&playedAt); err != nil {
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO matches (id, community_id, mode, played_at, notes)
		VALUES (?, ?, ?, ?, ?)`, matchID, req.CommunityID, string(mode), playedAt, req.Notes); err != nil {
		return nil, fmt.Errorf("insert match: %w", err)
	}
	deltas, err := applyMatchRatings(ctx, tx, req.CommunityID, matchID, mode, req.Participants, normalized, playedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &ReportOutcome{MatchID: matchID, Mode: mode, Deltas: deltas}, nil
}

// applyMatchRatings writes the match participants and folds the result into the
// materialized ratings. It is shared by reporting and by full recomputation.
func applyMatchRatings(ctx context.Context, tx *sql.Tx, communityID, matchID string, mode model.Mode, participants []Participant, normalized rating.NormalizedMatch, playedAt string) ([]model.RatingDelta, error) {
	// Load current ratings for each participant in this mode.
	var contestants []contestant
	for _, p := range participants {
		var mu, sigma float64
		if err := tx.QueryRowContext(ctx, `
			SELECT mu, sigma FROM ratings WHERE community_id = ? AND player_id = ? AND mode = ?`,
			communityID, p.PlayerID, string(mode)).Scan(&mu, &sigma); err != nil {
			// New player: default rating.
			defaultPlayer := rating.NewOpenSkillPlayer(p.PlayerID)
			mu, sigma = defaultPlayer.Mu, defaultPlayer.Sigma
		}
		contestants = append(contestants, contestant{req: p, mu: mu, sigma: sigma})
	}

	teams := make([]rating.OpenSkillTeam, len(normalized.Teams))
	for i, team := range normalized.Teams {
		for _, playerID := range team {
			for _, contestant := range contestants {
				if contestant.req.PlayerID == playerID {
					teams[i].Players = append(teams[i].Players, rating.OpenSkillPlayer{ID: playerID, Mu: contestant.mu, Sigma: contestant.sigma})
					break
				}
			}
		}
	}
	updatedTeams, err := rating.RateOpenSkill(teams, normalized.Ranks)
	if err != nil {
		return nil, fmt.Errorf("rate match: %w", err)
	}
	updates := make(map[string]rating.OpenSkillPlayer, len(contestants))
	for _, team := range updatedTeams {
		for _, player := range team.Players {
			updates[player.ID] = player
		}
	}

	deltas := make([]model.RatingDelta, 0, len(contestants))
	for _, c := range contestants {
		up := updates[c.req.PlayerID]
		previousOrdinal := rating.OpenSkillOrdinal(c.mu, c.sigma)
		newOrdinal := rating.OpenSkillOrdinal(up.Mu, up.Sigma)
		delta := (newOrdinal - previousOrdinal) * 60
		stats := normalized.Results[c.req.PlayerID]
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO match_participants (match_id, player_id, team, score, finish, rating_delta)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(match_id, player_id) DO UPDATE SET
				team = excluded.team, score = excluded.score,
				finish = excluded.finish, rating_delta = excluded.rating_delta`,
			matchID, c.req.PlayerID, c.req.Team, c.req.Score, nullableFinish(c.req), delta); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO ratings (community_id, player_id, mode, mu, sigma, games, wins, losses, draws, last_played)
			VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, ?)
			ON CONFLICT(community_id, player_id, mode) DO UPDATE SET
				mu = excluded.mu, sigma = excluded.sigma,
				games = ratings.games + 1,
				wins = ratings.wins + ?, losses = ratings.losses + ?, draws = ratings.draws + ?,
				last_played = excluded.last_played`,
			communityID, c.req.PlayerID, string(mode), up.Mu, up.Sigma,
			stats.Wins, stats.Losses, stats.Draws, playedAt, stats.Wins, stats.Losses, stats.Draws); err != nil {
			return nil, err
		}
		deltas = append(deltas, model.RatingDelta{
			PlayerID:  c.req.PlayerID,
			Mode:      mode,
			Delta:     delta,
			NewRating: rating.OpenSkillDisplayRating(up.Mu, up.Sigma),
		})
	}
	return deltas, nil
}

func validateCommunityParticipants(ctx context.Context, tx *sql.Tx, communityID string, participants []Participant) error {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(participants)), ",")
	args := make([]any, 0, len(participants)+1)
	args = append(args, communityID)
	for _, participant := range participants {
		args = append(args, participant.PlayerID)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT p.id,
		       EXISTS(
			       SELECT 1
			       FROM community_members cm
			       WHERE cm.user_id = p.user_id AND cm.community_id = ?
		       )
		FROM players p
		WHERE p.id IN (`+placeholders+`)`, args...)
	if err != nil {
		return fmt.Errorf("check community participants: %w", err)
	}
	defer rows.Close()

	eligible := make(map[string]bool, len(participants))
	for rows.Next() {
		var playerID string
		var isEligible bool
		if err := rows.Scan(&playerID, &isEligible); err != nil {
			return fmt.Errorf("scan community participant: %w", err)
		}
		eligible[playerID] = isEligible
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read community participants: %w", err)
	}

	for _, participant := range participants {
		isEligible, exists := eligible[participant.PlayerID]
		if !exists {
			return fmt.Errorf("player %q: %w", participant.PlayerID, ErrUnknownPlayer)
		}
		if !isEligible {
			return fmt.Errorf("player %q is not owned by a member of community %q: %w", participant.PlayerID, communityID, ErrIneligibleParticipant)
		}
	}
	return nil
}

// ReportOutcome is returned to the client after a match report.
type ReportOutcome struct {
	MatchID string              `json:"matchId"`
	Mode    model.Mode          `json:"mode"`
	Deltas  []model.RatingDelta `json:"deltas"`
}

type MatchHistory struct {
	MatchID      string               `json:"matchId"`
	CommunityID  string               `json:"communityId"`
	Mode         model.Mode           `json:"mode"`
	PlayedAt     string               `json:"playedAt"`
	Participants []ParticipantHistory `json:"participants"`
}
type ParticipantHistory struct {
	PlayerID    string  `json:"playerId"`
	Handle      string  `json:"handle"`
	Team        string  `json:"team"`
	Score       float64 `json:"score"`
	Finish      int     `json:"finish"`
	RatingDelta float64 `json:"ratingDelta"`
}

// History lists matches, optionally narrowed to one player, community or mode.
func (s *Store) History(ctx context.Context, playerID, communityID string, mode model.Mode, limit, offset int) ([]MatchHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	query := `WITH selected_matches AS (
		SELECT DISTINCT m.id, m.community_id, m.mode, m.played_at
		FROM matches m
		JOIN match_participants filter_mp ON filter_mp.match_id = m.id
		WHERE 1 = 1`
	args := []any{}
	if playerID != "" {
		query += ` AND filter_mp.player_id = ?`
		args = append(args, playerID)
	}
	if communityID != "" {
		query += ` AND m.community_id = ?`
		args = append(args, communityID)
	}
	if mode != "" {
		query += ` AND m.mode = ?`
		args = append(args, string(mode))
	}
	query += ` ORDER BY m.played_at DESC LIMIT ? OFFSET ?)
		SELECT sm.id, sm.community_id, sm.mode, sm.played_at,
		       mp.player_id, COALESCE(p.handle, mp.player_id), COALESCE(mp.team, ''), mp.score,
		       COALESCE(mp.finish, -1), mp.rating_delta
		FROM selected_matches sm
		JOIN match_participants mp ON mp.match_id = sm.id
		LEFT JOIN players p ON p.id = mp.player_id
		ORDER BY sm.played_at DESC, sm.id, mp.finish, mp.score DESC, mp.player_id`
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MatchHistory
	indexes := make(map[string]int)
	for rows.Next() {
		var h MatchHistory
		var p ParticipantHistory
		if err := rows.Scan(&h.MatchID, &h.CommunityID, &h.Mode, &h.PlayedAt,
			&p.PlayerID, &p.Handle, &p.Team, &p.Score, &p.Finish, &p.RatingDelta); err != nil {
			return nil, err
		}
		index, ok := indexes[h.MatchID]
		if !ok {
			index = len(out)
			indexes[h.MatchID] = index
			out = append(out, h)
		}
		out[index].Participants = append(out[index].Participants, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// normalizeReport derives the mode from the participant set, validates it, and
// prepares algorithm-ready teams and ranks. Participants may be mutated to fill
// in finishes that were implied by scores.
func normalizeReport(participants []Participant) (model.Mode, rating.NormalizedMatch, error) {
	var empty rating.NormalizedMatch
	mode := deriveMode(participants)
	if !mode.Valid() {
		return "", empty, fmt.Errorf("invalid mode %q", mode)
	}
	if len(participants) < 2 {
		return "", empty, fmt.Errorf("a match needs at least two participants")
	}
	seen := map[string]bool{}
	for _, p := range participants {
		if p.PlayerID == "" || seen[p.PlayerID] {
			return "", empty, fmt.Errorf("participants must have unique player ids")
		}
		seen[p.PlayerID] = true
	}
	if mode == model.Mode1v1 && len(participants) != 2 {
		return "", empty, fmt.Errorf("1v1 requires exactly two participants")
	}
	if mode == model.ModeFFA && len(participants) < 3 {
		return "", empty, fmt.Errorf("ffa requires at least three participants")
	}
	if mode == model.ModeTeam {
		teams := map[string]bool{}
		for _, p := range participants {
			if p.Team == "" {
				return "", empty, fmt.Errorf("team participants need team labels")
			}
			teams[p.Team] = true
		}
		if len(teams) < 2 {
			return "", empty, fmt.Errorf("team mode requires at least two teams")
		}
	}
	if mode == model.ModeFFA {
		// Solo-team FFAs are reported with scores rather than finishes.
		if allTeamed(participants) {
			fillFinishesFromScores(participants)
		}
		for _, p := range participants {
			if p.Finish < 0 {
				return "", empty, fmt.Errorf("ffa participants need finish positions")
			}
		}
	}
	ratingParticipants := make([]rating.MatchParticipant, len(participants))
	for i, p := range participants {
		ratingParticipants[i] = rating.MatchParticipant{ID: p.PlayerID, Team: p.Team, Score: p.Score, Finish: p.Finish}
	}
	normalized, err := rating.NormalizeMatch(rating.MatchMode(mode), ratingParticipants)
	if err != nil {
		return "", empty, err
	}
	return mode, normalized, nil
}

// deriveMode maps the participant/team configuration to a mode.
func deriveMode(parts []Participant) model.Mode {
	if len(parts) <= 2 {
		return model.Mode1v1
	}

	teams := map[string]int{}
	for _, p := range parts {
		if p.Team != "" {
			teams[p.Team]++
		}
	}
	// Three or more unteamed participants are an FFA.
	if len(teams) == 0 {
		return model.ModeFFA
	}
	// More than two teams that each hold a single player is also an FFA.
	if len(teams) > 2 {
		solo := true
		for _, size := range teams {
			if size != 1 {
				solo = false
				break
			}
		}
		if solo {
			return model.ModeFFA
		}
	}
	return model.ModeTeam
}

func allTeamed(parts []Participant) bool {
	for _, p := range parts {
		if p.Team == "" {
			return false
		}
	}
	return true
}

// fillFinishesFromScores ranks participants by score (highest first, ties share
// a finish) when the reporter supplied scores instead of finish positions.
func fillFinishesFromScores(parts []Participant) {
	scores := make([]float64, 0, len(parts))
	for _, p := range parts {
		scores = append(scores, p.Score)
	}
	for i := range parts {
		ahead := 0
		for _, score := range scores {
			if score > parts[i].Score {
				ahead++
			}
		}
		parts[i].Finish = ahead
	}
}

func nullableFinish(p Participant) any {
	if p.Finish >= 0 {
		return p.Finish
	}
	return nil
}
