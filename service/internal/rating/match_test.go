package rating

import "testing"

func TestNormalizeMatchAggregatesTeams(t *testing.T) {
	match, err := NormalizeMatch(ModeTeam, []MatchParticipant{
		{ID: "a1", Team: "A", Score: 10},
		{ID: "a2", Team: "A", Score: 1},
		{ID: "b1", Team: "B", Score: 5},
		{ID: "b2", Team: "B", Score: 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(match.Teams) != 2 || len(match.Teams[0]) != 2 || match.Ranks[0] != 0 || match.Ranks[1] != 1 {
		t.Fatalf("normalized team match = %+v", match)
	}
	if got := match.Results["a1"]; got.Wins != 1 || got.Games != 1 {
		t.Fatalf("winner stats = %+v", got)
	}
	if got := match.Results["b1"]; got.Losses != 1 || got.Games != 1 {
		t.Fatalf("loser stats = %+v", got)
	}
}

func TestNormalizeMatchPreservesFFATies(t *testing.T) {
	match, err := NormalizeMatch(ModeFFA, []MatchParticipant{
		{ID: "a", Finish: 0},
		{ID: "b", Finish: 0},
		{ID: "c", Finish: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(match.Teams) != 3 || match.Ranks[0] != 0 || match.Ranks[1] != 0 || match.Ranks[2] != 2 {
		t.Fatalf("normalized FFA = %+v", match)
	}
	if match.Results["a"].Draws != 1 || match.Results["b"].Draws != 1 || match.Results["c"].Losses != 1 {
		t.Fatalf("FFA stats = %+v", match.Results)
	}
}

func TestNormalizeMatchAllowsUnevenTeams(t *testing.T) {
	match, err := NormalizeMatch(ModeTeam, []MatchParticipant{
		{ID: "a1", Team: "A"},
		{ID: "a2", Team: "A"},
		{ID: "b1", Team: "B"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(match.Teams) != 2 || len(match.Teams[0]) != 2 || len(match.Teams[1]) != 1 {
		t.Fatalf("normalized team match = %+v", match)
	}
}
