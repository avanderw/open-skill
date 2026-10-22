package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func TestMatchPlayedDatePersistenceAndEditing(t *testing.T) {
	st, ctx := testCommunityStore(t)
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_dates", Name: "Dates", Code: "DATES", Mode: "open", Draws: true, OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	player, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, "Opponent")
	if err != nil {
		t.Fatal(err)
	}
	participants := []Participant{{PlayerID: "p_demo", Score: 1, Finish: -1}, {PlayerID: player.ID, Finish: -1}}
	for _, invalid := range []string{"2026-02-30", "yesterday", "2026-13-01"} {
		if _, err := st.ReportMatchAsMember(ctx, "u_demo", ReportMatchRequest{CommunityID: community.ID, Participants: participants, PlayedAt: invalid}); err == nil {
			t.Fatalf("accepted invalid date %q", invalid)
		}
	}
	outcome, err := st.ReportMatchAsMember(ctx, "u_demo", ReportMatchRequest{CommunityID: community.ID, Participants: participants, PlayedAt: "2026-10-12"})
	if err != nil {
		t.Fatal(err)
	}
	match, err := st.Match(ctx, outcome.MatchID)
	if err != nil || match.PlayedAt != "2026-10-12 00:00:00" {
		t.Fatalf("stored date = %+v, err = %v", match, err)
	}
	if _, err := st.UpdateMatch(ctx, "u_demo", outcome.MatchID, participants, "", "2026-10-10"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpdateMatch(ctx, "u_demo", outcome.MatchID, participants, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpdateMatch(ctx, "u_demo", outcome.MatchID, participants, "", "2026-02-30"); err == nil {
		t.Fatal("accepted invalid edit date")
	}
	history, err := st.History(ctx, "", community.ID, "", 20, 0)
	if err != nil || len(history) != 1 || history[0].PlayedAt != "2026-10-10 00:00:00" {
		t.Fatalf("edited date = %+v, err = %v", history, err)
	}
}

func TestBackdatedMatchReplaysRatingsInPlayedOrder(t *testing.T) {
	var expectedMu, expectedSigma float64
	for _, reversed := range []bool{false, true} {
		st, ctx := testCommunityStore(t)
		community, err := st.CreateCommunity(ctx, model.Community{ID: "c_order", Name: "Order", Code: "ORDER", Mode: "open", Draws: true, OwnerID: "u_demo"})
		if err != nil {
			t.Fatal(err)
		}
		player, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, "Opponent")
		if err != nil {
			t.Fatal(err)
		}
		reports := []ReportMatchRequest{
			{CommunityID: community.ID, PlayedAt: "2026-10-01", Participants: []Participant{{PlayerID: "p_demo", Score: 1, Finish: -1}, {PlayerID: player.ID, Finish: -1}}},
			{CommunityID: community.ID, PlayedAt: "2026-10-10", Participants: []Participant{{PlayerID: "p_demo", Finish: -1}, {PlayerID: player.ID, Score: 1, Finish: -1}}},
		}
		if reversed {
			reports[0], reports[1] = reports[1], reports[0]
		}
		for _, report := range reports {
			if _, err := st.ReportMatchAsMember(ctx, "u_demo", report); err != nil {
				t.Fatal(err)
			}
		}
		var mu, sigma float64
		if err := st.DB().QueryRow(`SELECT mu, sigma FROM ratings WHERE community_id = ? AND player_id = 'p_demo' AND mode = '1v1'`, community.ID).Scan(&mu, &sigma); err != nil {
			t.Fatal(err)
		}
		if !reversed {
			expectedMu, expectedSigma = mu, sigma
		} else if mu != expectedMu || sigma != expectedSigma {
			t.Fatalf("backdated ratings = (%v,%v), chronological = (%v,%v)", mu, sigma, expectedMu, expectedSigma)
		}
	}
}

func TestReportMatchRejectsDuplicatePlayers(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	_, err = st.ReportMatch(context.Background(), ReportMatchRequest{CommunityID: "c_demo", Participants: []Participant{{PlayerID: "p_1"}, {PlayerID: "p_1"}}})
	if err == nil {
		t.Fatal("expected duplicate-player validation error")
	}
}

func TestReportMatchRejectsNonMemberParticipantWithoutWrites(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_member', 'Member'), ('u_guest', 'Guest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_member', 'u_member', 'Member'), ('p_guest', 'u_guest', 'Guest');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_member');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_member', 'owner');`); err != nil {
		t.Fatal(err)
	}

	_, err = st.ReportMatch(ctx, ReportMatchRequest{
		CommunityID: "c_test",
		Participants: []Participant{
			{PlayerID: "p_member", Score: 1},
			{PlayerID: "p_guest", Score: 0},
		},
	})
	if err == nil {
		t.Fatal("expected non-member participant validation error")
	}

	var matches, ratings int
	if err := st.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE community_id = 'c_test'`).Scan(&matches); err != nil {
		t.Fatal(err)
	}
	if err := st.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ratings WHERE community_id = 'c_test'`).Scan(&ratings); err != nil {
		t.Fatal(err)
	}
	if matches != 0 || ratings != 0 {
		t.Fatalf("rejected match wrote matches=%d ratings=%d", matches, ratings)
	}
}

func TestReportMatchAsMemberRejectsNonMemberWithoutWrites(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_owner', 'Owner'), ('u_guest', 'Guest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_owner', 'u_owner', 'Owner'), ('p_guest', 'u_guest', 'Guest');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_owner');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_owner', 'owner'), ('c_test', 'u_guest', 'member');`); err != nil {
		t.Fatal(err)
	}

	if _, err := st.ReportMatchAsMember(ctx, "u_unknown", ReportMatchRequest{
		CommunityID: "c_test",
		Participants: []Participant{
			{PlayerID: "p_owner", Score: 1},
			{PlayerID: "p_guest", Score: 0},
		},
	}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}

	var matches int
	if err := st.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM matches WHERE community_id = 'c_test'`).Scan(&matches); err != nil {
		t.Fatal(err)
	}
	if matches != 0 {
		t.Fatalf("rejected report wrote %d matches, want 0", matches)
	}
}

func TestReportMatchRejectsUnknownParticipant(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_member', 'Member');
		INSERT INTO players (id, user_id, handle) VALUES ('p_member', 'u_member', 'Member');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_member');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_member', 'owner');`); err != nil {
		t.Fatal(err)
	}

	_, err = st.ReportMatch(ctx, ReportMatchRequest{
		CommunityID: "c_test",
		Participants: []Participant{
			{PlayerID: "p_member", Score: 1},
			{PlayerID: "p_missing", Score: 0},
		},
	})
	if err == nil {
		t.Fatal("expected unknown participant validation error")
	}
}

func TestReportMatchRejectsFFAWithoutFinishes(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	_, err = st.ReportMatch(context.Background(), ReportMatchRequest{CommunityID: "c_demo", Participants: []Participant{{PlayerID: "p_1"}, {PlayerID: "p_2"}, {PlayerID: "p_3"}}})
	if err == nil {
		t.Fatal("expected FFA finish validation error")
	}
}

func TestReportMatchDerivesModeFromParticipants(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.DB().Exec(`
		INSERT INTO users (id, display) VALUES ('u_test', 'Test');
		INSERT INTO players (id, user_id, handle) VALUES
			('p_1', 'u_test', 'One'), ('p_2', 'u_test', 'Two'), ('p_3', 'u_test', 'Three');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_test');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_test', 'owner');`); err != nil {
		t.Fatal(err)
	}

	outcome, err := st.ReportMatch(context.Background(), ReportMatchRequest{
		CommunityID: "c_test",
		Participants: []Participant{
			{PlayerID: "p_1", Finish: 0},
			{PlayerID: "p_2", Finish: 1},
			{PlayerID: "p_3", Finish: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Mode != model.ModeFFA {
		t.Fatalf("expected derived FFA mode, got %q", outcome.Mode)
	}
}
