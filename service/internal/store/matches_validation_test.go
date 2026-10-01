package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

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
