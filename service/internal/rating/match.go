package rating

import (
	"fmt"
	"math"
)

// MatchMode identifies the rules used to normalize one reported match.
type MatchMode string

const (
	Mode1v1  MatchMode = "1v1"
	ModeTeam MatchMode = "team"
	ModeFFA  MatchMode = "ffa"
)

type MatchParticipant struct {
	ID     string
	Team   string
	Score  float64
	Finish int
}

// NormalizedMatch contains algorithm-ready team membership and ranks.
type NormalizedMatch struct {
	Teams   [][]string
	Ranks   []int
	Results map[string]MatchStats
}

type MatchStats struct {
	Games  int
	Wins   int
	Losses int
	Draws  int
}

// NormalizeMatch converts participant data into native team/rank inputs.
// Ranks are zero-based; equal ranks represent ties.
func NormalizeMatch(mode MatchMode, participants []MatchParticipant) (NormalizedMatch, error) {
	if len(participants) < 2 {
		return NormalizedMatch{}, fmt.Errorf("a match needs at least two participants")
	}
	seen := make(map[string]bool, len(participants))
	for _, participant := range participants {
		if participant.ID == "" || seen[participant.ID] {
			return NormalizedMatch{}, fmt.Errorf("participants must have unique ids")
		}
		if math.IsNaN(participant.Score) || math.IsInf(participant.Score, 0) {
			return NormalizedMatch{}, fmt.Errorf("scores must be finite")
		}
		seen[participant.ID] = true
	}

	match := NormalizedMatch{Results: make(map[string]MatchStats, len(participants))}
	switch mode {
	case Mode1v1:
		if len(participants) != 2 {
			return NormalizedMatch{}, fmt.Errorf("1v1 requires exactly two participants")
		}
		match.Teams = [][]string{{participants[0].ID}, {participants[1].ID}}
		match.Ranks = ranksFromScores(participants[0].Score, participants[1].Score)
	case ModeTeam:
		teams := make(map[string][]string)
		order := make([]string, 0)
		for _, participant := range participants {
			if participant.Team == "" {
				return NormalizedMatch{}, fmt.Errorf("team participants need team labels")
			}
			if _, exists := teams[participant.Team]; !exists {
				order = append(order, participant.Team)
			}
			teams[participant.Team] = append(teams[participant.Team], participant.ID)
		}
		if len(order) < 2 {
			return NormalizedMatch{}, fmt.Errorf("team mode requires at least two teams")
		}
		teamScores := make([]float64, len(order))
		for i, team := range order {
			match.Teams = append(match.Teams, teams[team])
			teamScores[i] = teamScore(participants, team)
		}
		match.Ranks = ranksFromScores(teamScores...)
	case ModeFFA:
		if len(participants) < 3 {
			return NormalizedMatch{}, fmt.Errorf("ffa requires at least three participants")
		}
		for _, participant := range participants {
			if participant.Finish < 0 {
				return NormalizedMatch{}, fmt.Errorf("ffa participants need finish positions")
			}
			match.Teams = append(match.Teams, []string{participant.ID})
			match.Ranks = append(match.Ranks, participant.Finish)
		}
	default:
		return NormalizedMatch{}, fmt.Errorf("invalid mode %q", mode)
	}

	for i, team := range match.Teams {
		result := MatchStats{Games: 1}
		if match.Ranks[i] == 0 {
			firstPlaceTied := countRank(match.Ranks, 0) > 1
			if firstPlaceTied {
				result.Draws = 1
			} else {
				result.Wins = 1
			}
		} else {
			result.Losses = 1
		}
		for _, playerID := range team {
			match.Results[playerID] = result
		}
	}
	return match, nil
}

func teamScore(participants []MatchParticipant, team string) float64 {
	var total float64
	for _, participant := range participants {
		if participant.Team == team {
			total += participant.Score
		}
	}
	return total
}

func ranksFromScores(scores ...float64) []int {
	ranks := make([]int, len(scores))
	for i, score := range scores {
		for _, other := range scores {
			if other > score {
				ranks[i]++
			}
		}
	}
	return ranks
}

func countRank(ranks []int, rank int) int {
	count := 0
	for _, value := range ranks {
		if value == rank {
			count++
		}
	}
	return count
}
