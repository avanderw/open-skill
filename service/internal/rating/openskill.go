package rating

import (
	"fmt"

	openskill "github.com/intinig/go-openskill/rating"
	"github.com/intinig/go-openskill/types"
)

const (
	DefaultOpenSkillMu    = 25.0
	DefaultOpenSkillSigma = DefaultOpenSkillMu / 3
)

// OpenSkillPlayer is the persisted skill state for one player in one mode and
// community. IDs are kept outside the third-party positional team slices.
type OpenSkillPlayer struct {
	ID    string
	Mu    float64
	Sigma float64
}

// OpenSkillTeam groups players that share one match rank.
type OpenSkillTeam struct {
	Players []OpenSkillPlayer
}

func NewOpenSkillPlayer(id string) OpenSkillPlayer {
	return OpenSkillPlayer{ID: id, Mu: DefaultOpenSkillMu, Sigma: DefaultOpenSkillSigma}
}

func OpenSkillOrdinal(mu, sigma float64) float64 {
	return mu - 3*sigma
}

func OpenSkillDisplayRating(mu, sigma float64) float64 {
	return 1500 + OpenSkillOrdinal(mu, sigma)*60
}

// RateOpenSkill applies one OpenSkill match. Ranks are zero-based and may be
// equal for ties. The returned teams preserve the input player ordering.
func RateOpenSkill(teams []OpenSkillTeam, ranks []int) ([]OpenSkillTeam, error) {
	if len(teams) < 2 {
		return nil, fmt.Errorf("a match needs at least two teams")
	}
	if len(ranks) != len(teams) {
		return nil, fmt.Errorf("rank count must match team count")
	}

	input := make([]types.Team, len(teams))
	for i, team := range teams {
		if len(team.Players) == 0 {
			return nil, fmt.Errorf("team %d is empty", i)
		}
		input[i] = make(types.Team, len(team.Players))
		for j, player := range team.Players {
			if player.ID == "" {
				return nil, fmt.Errorf("team %d player %d has no id", i, j)
			}
			input[i][j] = types.Rating{Mu: player.Mu, Sigma: player.Sigma, Z: 3}
		}
	}

	updated := openskill.Rate(input, &types.OpenSkillOptions{Rank: ranks})
	if len(updated) != len(teams) {
		return nil, fmt.Errorf("rating engine returned %d teams, want %d", len(updated), len(teams))
	}
	result := make([]OpenSkillTeam, len(teams))
	for i, team := range updated {
		if len(team) != len(teams[i].Players) {
			return nil, fmt.Errorf("rating engine changed team %d size", i)
		}
		result[i].Players = make([]OpenSkillPlayer, len(team))
		for j, player := range team {
			result[i].Players[j] = OpenSkillPlayer{ID: teams[i].Players[j].ID, Mu: player.Mu, Sigma: player.Sigma}
		}
	}
	return result, nil
}
