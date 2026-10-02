package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

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

	var drawsAllowed, isMember bool
	if err := tx.QueryRowContext(ctx, `
		SELECT c.draws,
		       c.owner_id = ? OR EXISTS(SELECT 1 FROM community_members cm WHERE cm.community_id = c.id AND cm.user_id = ?)
		FROM communities c WHERE c.id = ?`, actorID, actorID, req.CommunityID).Scan(&drawsAllowed, &isMember); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("community %q: %w", req.CommunityID, ErrNotFound)
		}
		return nil, fmt.Errorf("load community %q: %w", req.CommunityID, err)
	}
	if actorID != "" && !isMember {
		return nil, ErrForbidden
	}
	if err := validateCommunityParticipants(ctx, tx, req.CommunityID, req.Participants); err != nil {
		return nil, err
	}
	if !drawsAllowed && mode != model.ModeFFA && normalized.Ranks[0] == normalized.Ranks[1] {
		return nil, fmt.Errorf("draws are not allowed in this community")
	}

	matchID := newID("m")
	playedAt := req.PlayedAt
	if playedAt == "" {
		// Same format as SQLite's datetime('now').
		playedAt = time.Now().UTC().Format("2006-01-02 15:04:05")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO matches (id, community_id, mode, played_at, notes)
		VALUES (?, ?, ?, ?, ?)`, matchID, req.CommunityID, string(mode), playedAt, req.Notes); err != nil {
		return nil, fmt.Errorf("insert match: %w", err)
	}
	book, err := loadRatings(ctx, tx, req.CommunityID, mode, req.Participants)
	if err != nil {
		return nil, err
	}
	rows, deltas, err := book.rateMatch(matchID, mode, req.Participants, normalized, playedAt)
	if err != nil {
		return nil, err
	}
	if err := writeParticipants(ctx, tx, rows); err != nil {
		return nil, err
	}
	if err := writeRatings(ctx, tx, req.CommunityID, book); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &ReportOutcome{MatchID: matchID, Mode: mode, Deltas: deltas}, nil
}

type ratingKey struct {
	playerID string
	mode     model.Mode
}

type ratingState struct {
	mu, sigma                  float64
	games, wins, losses, draws int
	lastPlayed                 sql.NullString
}

// ratingBook holds ratings in memory so matches (or a full replay) are rated
// without a database round trip per participant.
type ratingBook map[ratingKey]*ratingState

func (b ratingBook) get(playerID string, mode model.Mode) *ratingState {
	key := ratingKey{playerID: playerID, mode: mode}
	if state, ok := b[key]; ok {
		return state
	}
	defaults := rating.NewOpenSkillPlayer(playerID)
	state := &ratingState{mu: defaults.Mu, sigma: defaults.Sigma}
	b[key] = state
	return state
}

// participantRow is a match_participants row to be written.
type participantRow struct {
	matchID string
	p       Participant
	delta   float64
}

func loadRatings(ctx context.Context, tx *sql.Tx, communityID string, mode model.Mode, participants []Participant) (ratingBook, error) {
	args := []any{communityID, string(mode)}
	for _, p := range participants {
		args = append(args, p.PlayerID)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT player_id, mu, sigma, games, wins, losses, draws, last_played
		FROM ratings WHERE community_id = ? AND mode = ? AND player_id IN (`+placeholders(len(participants))+`)`, args...)
	if err != nil {
		return nil, fmt.Errorf("load ratings: %w", err)
	}
	defer rows.Close()
	book := ratingBook{}
	for rows.Next() {
		var playerID string
		var state ratingState
		if err := rows.Scan(&playerID, &state.mu, &state.sigma, &state.games, &state.wins, &state.losses, &state.draws, &state.lastPlayed); err != nil {
			return nil, err
		}
		book[ratingKey{playerID: playerID, mode: mode}] = &state
	}
	return book, rows.Err()
}

// rateMatch applies one match to the in-memory ratings and returns the
// participant rows to persist. It is shared by reporting and full recomputation.
func (b ratingBook) rateMatch(matchID string, mode model.Mode, participants []Participant, normalized rating.NormalizedMatch, playedAt string) ([]participantRow, []model.RatingDelta, error) {
	states := make(map[string]*ratingState, len(participants))
	for _, p := range participants {
		states[p.PlayerID] = b.get(p.PlayerID, mode)
	}

	teams := make([]rating.OpenSkillTeam, len(normalized.Teams))
	for i, team := range normalized.Teams {
		for _, playerID := range team {
			if state, ok := states[playerID]; ok {
				teams[i].Players = append(teams[i].Players, rating.OpenSkillPlayer{ID: playerID, Mu: state.mu, Sigma: state.sigma})
			}
		}
	}
	updatedTeams, err := rating.RateOpenSkill(teams, normalized.Ranks)
	if err != nil {
		return nil, nil, fmt.Errorf("rate match: %w", err)
	}
	updates := make(map[string]rating.OpenSkillPlayer, len(participants))
	for _, team := range updatedTeams {
		for _, player := range team.Players {
			updates[player.ID] = player
		}
	}

	rows := make([]participantRow, 0, len(participants))
	deltas := make([]model.RatingDelta, 0, len(participants))
	for _, p := range participants {
		state := states[p.PlayerID]
		up := updates[p.PlayerID]
		delta := (rating.OpenSkillOrdinal(up.Mu, up.Sigma) - rating.OpenSkillOrdinal(state.mu, state.sigma)) * 60
		stats := normalized.Results[p.PlayerID]
		state.mu, state.sigma = up.Mu, up.Sigma
		state.games++
		state.wins += stats.Wins
		state.losses += stats.Losses
		state.draws += stats.Draws
		state.lastPlayed = sql.NullString{String: playedAt, Valid: true}
		rows = append(rows, participantRow{matchID: matchID, p: p, delta: delta})
		deltas = append(deltas, model.RatingDelta{
			PlayerID:  p.PlayerID,
			Mode:      mode,
			Delta:     delta,
			NewRating: rating.OpenSkillDisplayRating(up.Mu, up.Sigma),
		})
	}
	return rows, deltas, nil
}

// maxRowsPerStatement keeps multi-row statements well under SQLite's bound
// parameter limit and remote request size limits.
const maxRowsPerStatement = 200

// execValues runs prefix + "(?,...),(?,...)" + suffix in chunks of rows.
func execValues(ctx context.Context, tx *sql.Tx, prefix, suffix string, width int, rows [][]any) error {
	for start := 0; start < len(rows); start += maxRowsPerStatement {
		chunk := rows[start:min(start+maxRowsPerStatement, len(rows))]
		var sb strings.Builder
		sb.WriteString(prefix)
		args := make([]any, 0, len(chunk)*width)
		for i, row := range chunk {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("(" + placeholders(width) + ")")
			args = append(args, row...)
		}
		sb.WriteString(suffix)
		if _, err := tx.ExecContext(ctx, sb.String(), args...); err != nil {
			return err
		}
	}
	return nil
}

func placeholders(n int) string {
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

func writeParticipants(ctx context.Context, tx *sql.Tx, rows []participantRow) error {
	values := make([][]any, len(rows))
	for i, r := range rows {
		values[i] = []any{r.matchID, r.p.PlayerID, r.p.Team, r.p.Score, nullableFinish(r.p), r.delta}
	}
	if err := execValues(ctx, tx, `
		INSERT INTO match_participants (match_id, player_id, team, score, finish, rating_delta) VALUES `, `
		ON CONFLICT(match_id, player_id) DO UPDATE SET
			team = excluded.team, score = excluded.score,
			finish = excluded.finish, rating_delta = excluded.rating_delta`, 6, values); err != nil {
		return fmt.Errorf("write match participants: %w", err)
	}
	return nil
}

func writeRatings(ctx context.Context, tx *sql.Tx, communityID string, book ratingBook) error {
	values := make([][]any, 0, len(book))
	for key, s := range book {
		values = append(values, []any{communityID, key.playerID, string(key.mode), s.mu, s.sigma, s.games, s.wins, s.losses, s.draws, s.lastPlayed})
	}
	if err := execValues(ctx, tx, `
		INSERT INTO ratings (community_id, player_id, mode, mu, sigma, games, wins, losses, draws, last_played) VALUES `, `
		ON CONFLICT(community_id, player_id, mode) DO UPDATE SET
			mu = excluded.mu, sigma = excluded.sigma, games = excluded.games,
			wins = excluded.wins, losses = excluded.losses, draws = excluded.draws,
			last_played = excluded.last_played`, 10, values); err != nil {
		return fmt.Errorf("write ratings: %w", err)
	}
	return nil
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
	out := []MatchHistory{}
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
