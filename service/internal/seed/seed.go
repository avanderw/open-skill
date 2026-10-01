// Package seed populates a fresh database with realistic demo content drawn
// from the mockup pages so the app has data to render on first run.
package seed

import (
	"database/sql"
	"fmt"

	"open-skill/service/internal/rating"
)

// Player is a player to seed with a starting rating.
type Player struct {
	Handle string
	Rating float64
}

// Seed sets up demo users, players, communities, memberships, and initial
// ratings. It is idempotent: if users already exist it does nothing.
func Seed(db *sql.DB) error {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return fmt.Errorf("check users: %w", err)
	}
	if n > 0 {
		return nil // already seeded
	}

	// Demo users (auth stub). Each player owns exactly one to make the "You"
	// highlight (by user) meaningful — the demo account "u_demo" owns IronCrest.
	users := []struct {
		id, display string
	}{
		{"u_demo", "IronCrest"},
		{"u_player", "Player"},
		{"u_member", "CommunityMember"},
		{"u_guest", "GuestPlayer"},
	}
	for _, u := range users {
		if _, err := db.Exec(
			`INSERT INTO users (id, display) VALUES (?, ?)`, u.id, u.display); err != nil {
			return fmt.Errorf("insert user %s: %w", u.id, err)
		}
	}

	// Players (handles + starting 1v1 ratings from the mockups).
	// All owned by u_player except IronCrest (u_demo).
	players := []struct {
		Handle string
		Rating float64
		UserID string
	}{
		{"CommunityMember", 1500, "u_member"},
		{"GuestPlayer", 1500, "u_guest"},
		{"shadowPlay", 1842, "u_player"},
		{"IronCrest", 1785, "u_demo"},
		{"LunarEcho", 1742, "u_player"},
		{"PixelMaster", 1698, "u_player"},
		{"VortexKing", 1645, "u_player"},
		{"NovaRift", 1621, "u_player"},
		{"TempestRook", 1598, "u_player"},
		{"BlazeLock", 1579, "u_player"},
		{"DriftKnight", 1552, "u_player"},
		{"QuantaEdge", 1528, "u_player"},
	}
	for i, p := range players {
		id := fmt.Sprintf("p_%03d", i)
		if _, err := db.Exec(
			`INSERT INTO players (id, user_id, handle) VALUES (?, ?, ?)`,
			id, p.UserID, p.Handle); err != nil {
			return fmt.Errorf("insert player %s: %w", p.Handle, err)
		}
	}

	// Communities.
	communities := []struct {
		id, name, mode, code string
		draws                bool
	}{
		{"c_apex", "Apex Legends Competitive", "approval", "APEX5K9", true},
		{"c_valorant", "Valorant Premier Hub", "open", "VALOR5", false},
	}
	for _, c := range communities {
		if _, err := db.Exec(
			`INSERT INTO communities (id, name, mode, draws, code, owner_id)
			 VALUES (?, ?, ?, ?, ?, ?)`, c.id, c.name, c.mode, c.draws, c.code, "u_demo"); err != nil {
			return fmt.Errorf("insert community %s: %w", c.id, err)
		}
		if _, err := db.Exec(
			`INSERT INTO community_members (community_id, user_id, role)
			 VALUES (?, ?, 'owner')`, c.id, "u_demo"); err != nil {
			return fmt.Errorf("insert membership %s: %w", c.id, err)
		}
		// Seed every materialized projection independently per community. The
		// 1v1 ratings reproduce the mockup; team and FFA start nearby but distinct.
		for i, p := range players {
			id := fmt.Sprintf("p_%03d", i)
			for _, projection := range []struct {
				mode   string
				rating float64
				games  int
			}{
				{"1v1", rating.DefaultOpenSkillMu + (p.Rating-1500)/60, 20},
				{"team", rating.DefaultOpenSkillMu + (p.Rating+60-1500)/60, 14},
				{"ffa", rating.DefaultOpenSkillMu + (p.Rating-35-1500)/60, 11},
			} {
				if _, err := db.Exec(
					`INSERT INTO ratings (community_id, player_id, mode, mu, sigma, games, wins)
					 VALUES (?, ?, ?, ?, ?, ?, ?)`,
					c.id, id, projection.mode, projection.rating, rating.DefaultOpenSkillSigma, projection.games, projection.games/2); err != nil {
					return fmt.Errorf("insert %s rating for %s in %s: %w", projection.mode, p.Handle, c.id, err)
				}
			}
		}
	}

	return nil
}
