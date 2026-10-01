package store

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"open-skill/service/internal/model"
	"open-skill/service/internal/rating"
)

type LeaderboardRow struct {
	Rank     int     `json:"rank"`
	PlayerID string  `json:"playerId"`
	Handle   string  `json:"handle"`
	Avatar   string  `json:"avatar"`
	Rating   float64 `json:"rating"`
	Games    int     `json:"games"`
	Wins     int     `json:"wins"`
	Losses   int     `json:"losses"`
	IsYou    bool    `json:"isYou"`
	Trend    float64 `json:"trend"`          // rating change over the player's last trendWindow matches
	Role     string  `json:"role,omitempty"` // community role: owner | admin | member
}

// trendWindow is how many recent matches feed the leaderboard trend. Three is
// enough to smooth a single fluke result without going stale in a small community.
const trendWindow = 3

// AvailableModes returns the modes that have at least one reported match,
// scoped to communityID when supplied. Order is stable: 1v1, team, ffa.
func (s *Store) AvailableModes(ctx context.Context, communityID string) ([]model.Mode, error) {
	query := `SELECT DISTINCT mode FROM matches`
	args := []any{}
	if communityID != "" {
		query += ` WHERE community_id = ?`
		args = append(args, communityID)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("available modes: %w", err)
	}
	defer rows.Close()
	found := map[model.Mode]bool{}
	for rows.Next() {
		var mode string
		if err := rows.Scan(&mode); err != nil {
			return nil, err
		}
		found[model.Mode(mode)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := []model.Mode{}
	for _, mode := range []model.Mode{model.Mode1v1, model.ModeTeam, model.ModeFFA} {
		if found[mode] {
			out = append(out, mode)
		}
	}
	return out, nil
}

// Leaderboard returns ratings scoped to communityID when supplied.
func (s *Store) Leaderboard(ctx context.Context, mode model.Mode, limit int, userID, communityID, q, sortBy string) ([]LeaderboardRow, error) {
	if communityID != "" {
		return s.communityLeaderboard(ctx, mode, limit, userID, communityID, q, sortBy)
	}
	query := `
		WITH recent AS (
			SELECT mp.player_id, mp.rating_delta,
			       ROW_NUMBER() OVER (PARTITION BY mp.player_id ORDER BY m.played_at DESC, m.id DESC) AS recency
			FROM match_participants mp
			JOIN matches m ON m.id = mp.match_id
			WHERE m.mode = ?
		),
		trend AS (
			SELECT player_id, SUM(rating_delta) AS trend FROM recent WHERE recency <= ? GROUP BY player_id
		),
		player_scores AS (
			SELECT p.id AS player_id, p.handle, p.user_id,
			       MAX(r.mu - 3 * r.sigma) AS ordinal,
			       SUM(r.games) AS games, SUM(r.wins) AS wins, SUM(r.losses) AS losses,
			       COALESCE(MAX(t.trend), 0) AS trend
			FROM ratings r
			JOIN players p ON p.id = r.player_id
			LEFT JOIN trend t ON t.player_id = p.id
			WHERE r.mode = ?`
	args := []any{string(mode), trendWindow, string(mode)}
	if q != "" {
		query += ` AND p.handle LIKE ? ESCAPE '\'`
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
		args = append(args, "%"+escaped+"%")
	}
	query += `
			GROUP BY p.id, p.handle, p.user_id
		)
		SELECT player_id, handle, 1500 + (ordinal * 60), games, wins, losses, trend,
		       (SELECT COUNT(*) + 1 FROM player_scores higher WHERE higher.ordinal > player_scores.ordinal) AS rank,
		       CASE WHEN user_id = ? THEN 1 ELSE 0 END AS is_you
		FROM player_scores`
	args = append(args, userID)
	switch sortBy {
	case "games":
		query += ` ORDER BY games DESC, ordinal DESC, handle ASC`
	default:
		query += ` ORDER BY ordinal DESC, handle ASC`
	}
	query += ` LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("leaderboard %s: %w", mode, err)
	}
	defer rows.Close()
	out := []LeaderboardRow{}
	for rows.Next() {
		var row LeaderboardRow
		var isYou int
		if err := rows.Scan(&row.PlayerID, &row.Handle, &row.Rating, &row.Games, &row.Wins, &row.Losses, &row.Trend, &row.Rank, &isYou); err != nil {
			return nil, err
		}
		row.Avatar = initials(row.Handle)
		row.IsYou = isYou == 1
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if userID != "" && !hasYou(out) {
		if you, err := s.youRow(ctx, mode, userID); err == nil {
			out = append(out, you)
		}
	}
	return out, nil
}

// communityLeaderboard lists every member of a community, falling back to the
// default (unrated) OpenSkill values so members appear before their first match.
func (s *Store) communityLeaderboard(ctx context.Context, mode model.Mode, limit int, userID, communityID, q, sortBy string) ([]LeaderboardRow, error) {
	query := `
		WITH recent AS (
			SELECT mp.player_id, mp.rating_delta,
			       ROW_NUMBER() OVER (PARTITION BY mp.player_id ORDER BY m.played_at DESC, m.id DESC) AS recency
			FROM match_participants mp
			JOIN matches m ON m.id = mp.match_id
			WHERE m.community_id = ? AND m.mode = ?
		),
		trend AS (
			SELECT player_id, SUM(rating_delta) AS trend FROM recent WHERE recency <= ? GROUP BY player_id
		)
		SELECT p.id, p.handle,
		       COALESCE(r.mu, ?) - 3 * COALESCE(r.sigma, ?) AS ordinal,
		       COALESCE(r.games, 0), COALESCE(r.wins, 0), COALESCE(r.losses, 0),
		       COALESCE(t.trend, 0),
		       CASE WHEN p.user_id = ? THEN 1 ELSE 0 END AS is_you,
		       cm.role
		FROM community_members cm
		JOIN players p ON p.user_id = cm.user_id
		LEFT JOIN ratings r
		       ON r.community_id = cm.community_id AND r.player_id = p.id AND r.mode = ?
		LEFT JOIN trend t ON t.player_id = p.id
		WHERE cm.community_id = ?`
	args := []any{communityID, string(mode), trendWindow, rating.DefaultOpenSkillMu, rating.DefaultOpenSkillSigma, userID, string(mode), communityID}
	if q != "" {
		query += ` AND p.handle LIKE ? ESCAPE '\'`
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
		args = append(args, "%"+escaped+"%")
	}
	query += ` ORDER BY ordinal DESC, p.handle ASC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("community leaderboard %s: %w", mode, err)
	}
	defer rows.Close()
	out := []LeaderboardRow{}
	for rows.Next() {
		var row LeaderboardRow
		var ordinal float64
		var isYou int
		if err := rows.Scan(&row.PlayerID, &row.Handle, &ordinal, &row.Games, &row.Wins, &row.Losses, &row.Trend, &isYou, &row.Role); err != nil {
			return nil, err
		}
		row.Rating = 1500 + ordinal*60
		row.Rank = len(out) + 1
		row.Avatar = initials(row.Handle)
		row.IsYou = isYou == 1
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if sortBy == "games" {
		sort.SliceStable(out, func(i, j int) bool { return out[i].Games > out[j].Games })
	}
	return out, nil
}

func (s *Store) youRow(ctx context.Context, mode model.Mode, userID string) (LeaderboardRow, error) {
	query := `WITH player_scores AS (
		SELECT p.id AS player_id, p.handle, p.user_id,
		       MAX(r.mu - 3 * r.sigma) AS ordinal,
		       SUM(r.games) AS games, SUM(r.wins) AS wins, SUM(r.losses) AS losses
		FROM ratings r JOIN players p ON p.id = r.player_id
		WHERE r.mode = ?
		GROUP BY p.id, p.handle, p.user_id
	)
	SELECT player_id, handle, 1500 + (ordinal * 60), games, wins, losses,
	       (SELECT COUNT(*) + 1 FROM player_scores higher WHERE higher.ordinal > player_scores.ordinal) AS rank
	FROM player_scores WHERE user_id = ? ORDER BY ordinal DESC, handle ASC LIMIT 1`
	var row LeaderboardRow
	err := s.db.QueryRowContext(ctx, query, string(mode), userID).Scan(&row.PlayerID, &row.Handle, &row.Rating, &row.Games, &row.Wins, &row.Losses, &row.Rank)
	if err != nil {
		return row, err
	}
	_ = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(rating_delta), 0) FROM (
			SELECT mp.rating_delta
			FROM match_participants mp
			JOIN matches m ON m.id = mp.match_id
			WHERE mp.player_id = ? AND m.mode = ?
			ORDER BY m.played_at DESC, m.id DESC
			LIMIT ?
		)`, row.PlayerID, string(mode), trendWindow).Scan(&row.Trend)
	row.Avatar = initials(row.Handle)
	row.IsYou = true
	return row, nil
}

func hasYou(rows []LeaderboardRow) bool {
	for _, r := range rows {
		if r.IsYou {
			return true
		}
	}
	return false
}

func (s *Store) PlayerByHandle(ctx context.Context, handle string) (*model.Player, error) {
	var p model.Player
	err := s.db.QueryRowContext(ctx, `SELECT id, user_id, handle FROM players WHERE handle = ?`, handle).Scan(&p.ID, &p.UserID, &p.Handle)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) Profile(ctx context.Context, playerID string) (*model.Player, []model.Rating, error) {
	var p model.Player
	if err := s.db.QueryRowContext(ctx, `SELECT id, user_id, handle FROM players WHERE id = ?`, playerID).Scan(&p.ID, &p.UserID, &p.Handle); err != nil {
		return nil, nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.community_id, COALESCE(c.name, r.community_id), r.player_id, r.mode,
		       1500 + ((r.mu - 3 * r.sigma) * 60), r.sigma, r.games, r.wins, r.losses, r.draws,
		       COALESCE(r.last_played, ''),
		       (SELECT COUNT(*) + 1
		          FROM community_members cm
		          JOIN players p2 ON p2.user_id = cm.user_id
		          LEFT JOIN ratings r2
		                 ON r2.community_id = cm.community_id AND r2.player_id = p2.id AND r2.mode = r.mode
		         WHERE cm.community_id = r.community_id
		           AND p2.id <> r.player_id
		           AND (COALESCE(r2.mu, ?) - 3 * COALESCE(r2.sigma, ?)) > (r.mu - 3 * r.sigma)) AS rank
		FROM ratings r
		LEFT JOIN communities c ON c.id = r.community_id
		WHERE r.player_id = ?
		ORDER BY COALESCE(c.name, r.community_id), r.mode`,
		rating.DefaultOpenSkillMu, rating.DefaultOpenSkillSigma, playerID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	ratings := []model.Rating{}
	for rows.Next() {
		var r model.Rating
		if err := rows.Scan(&r.CommunityID, &r.CommunityName, &r.PlayerID, &r.Mode, &r.Rating, &r.Uncertainty, &r.Games, &r.Wins, &r.Losses, &r.Draws, &r.LastPlayed, &r.Rank); err != nil {
			return nil, nil, err
		}
		ratings = append(ratings, r)
	}
	return &p, ratings, rows.Err()
}

func initials(name string) string {
	var b strings.Builder
	for i, w := range strings.Fields(name) {
		if i > 1 {
			break
		}
		b.WriteString(strings.ToUpper(w[:1]))
	}
	return b.String()
}
