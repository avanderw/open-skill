package store

import (
	"testing"

	"open-skill/service/internal/model"
)

func TestDeriveModeTwoUnteamedPlayersIs1v1(t *testing.T) {
	got := deriveMode([]Participant{{PlayerID: "a"}, {PlayerID: "b"}})
	if got != model.Mode1v1 {
		t.Fatalf("deriveMode() = %q, want %q", got, model.Mode1v1)
	}
}

func TestDeriveModeMultipleUnteamedPlayersIsFFA(t *testing.T) {
	got := deriveMode([]Participant{{PlayerID: "a"}, {PlayerID: "b"}, {PlayerID: "c"}})
	if got != model.ModeFFA {
		t.Fatalf("deriveMode() = %q, want %q", got, model.ModeFFA)
	}
}

func TestDeriveModeTwoTeamsIsTeam(t *testing.T) {
	got := deriveMode([]Participant{
		{PlayerID: "a", Team: "Blue"},
		{PlayerID: "b", Team: "Blue"},
		{PlayerID: "c", Team: "Orange"},
		{PlayerID: "d", Team: "Orange"},
	})
	if got != model.ModeTeam {
		t.Fatalf("deriveMode() = %q, want %q", got, model.ModeTeam)
	}
}

func TestDeriveModeExplicitIndividualTeamsIsFFA(t *testing.T) {
	got := deriveMode([]Participant{
		{PlayerID: "a", Team: "A"},
		{PlayerID: "b", Team: "B"},
		{PlayerID: "c", Team: "C"},
	})
	if got != model.ModeFFA {
		t.Fatalf("deriveMode() = %q, want %q", got, model.ModeFFA)
	}
}

func TestDeriveModeThreeTeamsWithPairIsTeam(t *testing.T) {
	got := deriveMode([]Participant{
		{PlayerID: "a", Team: "A"},
		{PlayerID: "b", Team: "B"},
		{PlayerID: "c", Team: "C"},
		{PlayerID: "d", Team: "C"},
	})
	if got != model.ModeTeam {
		t.Fatalf("deriveMode() = %q, want %q", got, model.ModeTeam)
	}
}
