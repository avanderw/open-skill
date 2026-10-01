// Package model defines the domain types for Open-Skill.
//
// Identity model (per product decision):
//   - A user is an account identity supplied by the (stubbed) auth layer.
//   - A player is a competitive handle owned by a user. A user may own many
//     players (handles), each with its own ratings across modes.
//
// Rating model (per product decision):
//   - Ratings are materialized per community and mode as a projection of the
//     match log. When a match is reported, the mode (1v1 | team | ffa) is
//     derived from the submitted participant/team configuration, then that
//     community/mode rating projection is updated with OpenSkill.
package model

// Mode is the competitive ruleset of a rating row, derived from the match log.
type Mode string

const (
	Mode1v1  Mode = "1v1"
	ModeTeam Mode = "team"
	ModeFFA  Mode = "ffa"
)

// User is an account identity (auth-stub id for now).
type User struct {
	ID        string `json:"id"`
	Display   string `json:"display"`
	CreatedAt string `json:"createdAt"`
}

// Player is a competitive handle owned by a user.
type Player struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Handle    string `json:"handle"`
	Avatar    string `json:"avatar"` // derived initials for the mockups
	CreatedAt string `json:"createdAt"`
}

// Rating is the materialized community-scoped rating for a player in one mode.
type Rating struct {
	CommunityID   string  `json:"communityId"`
	CommunityName string  `json:"communityName,omitempty"`
	PlayerID      string  `json:"playerId"`
	Mode          Mode    `json:"mode"`
	Rating        float64 `json:"rating"`
	Uncertainty   float64 `json:"uncertainty"`
	Games         int     `json:"games"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Draws         int     `json:"draws"`
	LastPlayed    string  `json:"lastPlayed"`
	Rank          int     `json:"rank,omitempty"` // standing within the community for this mode
}

// Community is a group of players who share rankings and matches.
type Community struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Mode       string `json:"mode"`  // "open" | "approval" — how players join
	Draws      bool   `json:"draws"` // whether draws are rated
	Code       string `json:"code"`  // invite code
	OwnerID    string `json:"ownerId"`
	Members    int    `json:"members"`
	YourGames  int    `json:"yourGames"` // matches the requesting user has played here
	CreatedAt  string `json:"createdAt"`
	Membership string `json:"membership,omitempty"` // owner | member | pending | non_member
}

// RatingDelta is one player's rating change from a reported match.
type RatingDelta struct {
	PlayerID  string  `json:"playerId"`
	Mode      Mode    `json:"mode"`
	Delta     float64 `json:"delta"`
	NewRating float64 `json:"newRating"`
}

// ModeName is a human label.
func (m Mode) String() string { return string(m) }

// Valid reports whether m is a supported mode.
func (m Mode) Valid() bool {
	switch m {
	case Mode1v1, ModeTeam, ModeFFA:
		return true
	}
	return false
}
