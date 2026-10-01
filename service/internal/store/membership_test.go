package store

import (
	"context"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func TestRemoveMemberHidesPlayerButPreservesHistoryAndRejoinRestoresEligibility(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_owner', 'Owner'), ('u_member', 'Member');
		INSERT INTO players (id, user_id, handle) VALUES ('p_owner', 'u_owner', 'Owner'), ('p_member', 'u_member', 'Member');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_owner');
		INSERT INTO community_members (community_id, user_id, role) VALUES
			('c_test', 'u_owner', 'owner'), ('c_test', 'u_member', 'member');`); err != nil {
		t.Fatal(err)
	}

	if _, err := st.ReportMatch(ctx, ReportMatchRequest{
		CommunityID: "c_test",
		Participants: []Participant{
			{PlayerID: "p_owner", Score: 1},
			{PlayerID: "p_member", Score: 0},
		},
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := st.Leaderboard(ctx, model.Mode1v1, 10, "u_owner", "c_test", "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("leaderboard before removal = %+v, want both players", rows)
	}

	if err := st.RemoveMember(ctx, "c_test", "u_owner", "u_member"); err != nil {
		t.Fatal(err)
	}

	rows, err = st.Leaderboard(ctx, model.Mode1v1, 10, "u_owner", "c_test", "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PlayerID != "p_owner" {
		t.Fatalf("leaderboard after removal = %+v, want only owner", rows)
	}

	history, err := st.History(ctx, "p_member", "c_test", model.Mode1v1, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 {
		t.Fatalf("member history after removal = %d rows, want 1", len(history))
	}

	var ratings int
	if err := st.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ratings WHERE community_id = 'c_test' AND player_id = 'p_member'`).Scan(&ratings); err != nil {
		t.Fatal(err)
	}
	if ratings != 1 {
		t.Fatalf("member ratings after removal = %d rows, want 1", ratings)
	}

	if err := st.AcceptCommunityMember(ctx, "c_test", "u_member"); err != nil {
		t.Fatal(err)
	}
	rows, err = st.Leaderboard(ctx, model.Mode1v1, 10, "u_owner", "c_test", "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("leaderboard after rejoin = %+v, want both players", rows)
	}
}

func TestApproveJoinRequestCreatesMembership(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_owner', 'Owner'), ('u_member', 'Member');
		INSERT INTO communities (id, name, mode, code, owner_id) VALUES ('c_approval', 'Approval', 'approval', 'APPR1', 'u_owner');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_approval', 'u_owner', 'owner');`); err != nil {
		t.Fatal(err)
	}

	if err := st.JoinCommunity(ctx, "c_approval", "u_member"); err != nil {
		t.Fatal(err)
	}
	status, err := st.Membership(ctx, "c_approval", "u_member")
	if err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("membership before approval = %q, want pending", status)
	}

	if err := st.ApproveJoinRequest(ctx, "c_approval", "u_member"); err != nil {
		t.Fatal(err)
	}
	status, err = st.Membership(ctx, "c_approval", "u_member")
	if err != nil {
		t.Fatal(err)
	}
	if status != "member" {
		t.Fatalf("membership after approval = %q, want member", status)
	}

	var requestStatus string
	if err := st.DB().QueryRowContext(ctx, `SELECT status FROM community_join_requests WHERE community_id = 'c_approval' AND user_id = 'u_member'`).Scan(&requestStatus); err != nil {
		t.Fatal(err)
	}
	if requestStatus != "approved" {
		t.Fatalf("join request status = %q, want approved", requestStatus)
	}
}
