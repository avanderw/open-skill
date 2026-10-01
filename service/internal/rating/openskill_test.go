package rating

import "testing"

func TestRateOpenSkillPreservesTeamShapeAndUpdatesPlayers(t *testing.T) {
	updated, err := RateOpenSkill([]OpenSkillTeam{
		{Players: []OpenSkillPlayer{{ID: "a1", Mu: 25, Sigma: 8.333}, {ID: "a2", Mu: 25, Sigma: 8.333}}},
		{Players: []OpenSkillPlayer{{ID: "b1", Mu: 25, Sigma: 8.333}, {ID: "b2", Mu: 25, Sigma: 8.333}}},
	}, []int{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 2 || len(updated[0].Players) != 2 || len(updated[1].Players) != 2 {
		t.Fatalf("unexpected team shape: %+v", updated)
	}
	if updated[0].Players[0].ID != "a1" || updated[1].Players[1].ID != "b2" {
		t.Fatalf("player IDs were not preserved: %+v", updated)
	}
	if updated[0].Players[0].Mu <= 25 || updated[1].Players[0].Mu >= 25 {
		t.Fatalf("decisive result did not update ratings: %+v", updated)
	}
}

func TestRateOpenSkillAcceptsTiedRanks(t *testing.T) {
	updated, err := RateOpenSkill([]OpenSkillTeam{
		{Players: []OpenSkillPlayer{{ID: "a", Mu: 25, Sigma: 8.333}}},
		{Players: []OpenSkillPlayer{{ID: "b", Mu: 25, Sigma: 8.333}}},
		{Players: []OpenSkillPlayer{{ID: "c", Mu: 25, Sigma: 8.333}}},
	}, []int{0, 0, 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 3 {
		t.Fatalf("updated teams = %d, want 3", len(updated))
	}
}
