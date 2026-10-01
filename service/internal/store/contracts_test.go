package store

import (
	"context"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func TestHistoryPaginationAndFiltering(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');
		INSERT INTO players (id, user_id, handle) VALUES
			('p_demo','u_demo','Demo'),('p_a','u_demo','Alpha'),('p_b','u_demo','Beta'),
			('p_x','u_demo','Xray'),('p_y','u_demo','Yank'),('p_z','u_demo','Zulu');
		INSERT INTO communities (id, name, mode, draws, code, owner_id) VALUES ('c_demo', 'Demo', 'open', 1, 'DEMO1', 'u_demo');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_demo', 'u_demo', 'owner');
		INSERT INTO ratings (community_id, player_id, mode, mu, sigma, games) VALUES ('c_demo', 'p_demo', '1v1', 25, 8.333333, 0), ('c_demo', 'p_demo', 'team', 25, 8.333333, 0);`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := st.ReportMatch(ctx, ReportMatchRequest{CommunityID: "c_demo", Participants: []Participant{{PlayerID: "p_demo", Score: 3}, {PlayerID: "p_a", Score: 1}}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := st.ReportMatch(ctx, ReportMatchRequest{CommunityID: "c_demo", Participants: []Participant{{PlayerID: "p_demo", Team: "A", Score: 1}, {PlayerID: "p_x", Team: "A", Score: 1}, {PlayerID: "p_y", Team: "B", Score: 0}, {PlayerID: "p_z", Team: "B", Score: 0}}}); err != nil {
		t.Fatal(err)
	}
	var games, wins, losses, draws int
	if err := st.DB().QueryRowContext(ctx, `SELECT games, wins, losses, draws FROM ratings WHERE community_id = 'c_demo' AND player_id = 'p_demo' AND mode = 'team'`).Scan(&games, &wins, &losses, &draws); err != nil {
		t.Fatal(err)
	}
	if games != 1 || wins != 1 || losses != 0 || draws != 0 {
		t.Fatalf("team statistics = games %d wins %d losses %d draws %d, want one match-level win", games, wins, losses, draws)
	}

	rows, err := st.History(ctx, "p_demo", "c_demo", model.Mode1v1, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("history filter = %d rows, want 2", len(rows))
	}

	page, err := st.History(ctx, "p_demo", "", "", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 {
		t.Fatalf("history page = %d rows, want 1", len(page))
	}
}

func TestCommunitySettingsAndJoinModePersistence(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');`); err != nil {
		t.Fatal(err)
	}
	c := model.Community{ID: "c_approve", Name: "Approval", Mode: "open", Draws: true, Code: "APPROV", OwnerID: "u_demo"}
	if _, err := st.CreateCommunity(ctx, c); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateCommunity(ctx, c.ID, "Approval", "approval", false); err != nil {
		t.Fatal(err)
	}

	updated, err := st.FindCommunityByID(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Mode != "approval" || updated.Draws {
		t.Fatalf("community not updated: %+v", updated)
	}

	if err := st.JoinCommunity(ctx, c.ID, "u_demo"); err != nil {
		t.Fatal(err)
	}
	requests, err := st.JoinRequests(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || requests[0].Status != "pending" {
		t.Fatalf("approval join = %+v", requests)
	}
}
