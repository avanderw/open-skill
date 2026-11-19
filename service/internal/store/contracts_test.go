package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func TestHistoryPlayersIncludesFullRosterAndFormerParticipants(t *testing.T) {
	st, ctx := testCommunityStore(t)
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_member', 'Member'), ('u_former', 'Former'), ('u_other', 'Other');
		INSERT INTO communities (id, name, code, owner_id) VALUES
			('c_history', 'History', 'HISTORY', 'u_demo'), ('c_other', 'Other', 'OTHER', 'u_other');
		INSERT INTO community_members (community_id, user_id, role) VALUES
			('c_history', 'u_demo', 'owner'), ('c_history', 'u_member', 'member');
		INSERT INTO players (id, user_id, handle) VALUES
			('p_former', 'u_former', 'Former player'), ('p_other', 'u_other', 'Other player');
		INSERT INTO players (id, handle, community_id) VALUES
			('p_guest', 'Local guest', 'c_history'), ('p_unrelated', 'Unrelated guest', 'c_other');
		INSERT INTO matches (id, community_id, mode, played_at) VALUES
			('m_old', 'c_history', '1v1', '2020-01-01'), ('m_other', 'c_other', '1v1', '2026-01-01');
		INSERT INTO match_participants (match_id, player_id) VALUES
			('m_old', 'p_former'), ('m_old', 'p_demo'), ('m_other', 'p_other');`); err != nil {
		t.Fatal(err)
	}
	tx, err := st.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	for i := 0; i < 105; i++ {
		if _, err := tx.ExecContext(ctx, `INSERT INTO players (id, user_id, handle) VALUES (?, 'u_member', ?)`,
			fmt.Sprintf("p_roster_%03d", i), fmt.Sprintf("Roster %03d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	players, err := st.HistoryPlayers(ctx, "c_history")
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 108 {
		t.Fatalf("options = %d, want all 105 members plus owner, former player, and guest", len(players))
	}
	found := map[string]HistoryPlayer{}
	for _, player := range players {
		found[player.PlayerID] = player
	}
	if len(found) != len(players) || found["p_former"].Handle != "Former player" || !found["p_guest"].Unclaimed {
		t.Fatalf("missing or duplicate historical/unclaimed players: %+v", found)
	}
	for _, excluded := range []string{"p_other", "p_unrelated"} {
		if _, ok := found[excluded]; ok {
			t.Fatalf("included unrelated player %s", excluded)
		}
	}
}

func TestHistoryPlayersDistinguishesEmptyAndMissingCommunities(t *testing.T) {
	st, ctx := testCommunityStore(t)
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_empty', 'Empty');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_empty', 'Empty', 'EMPTY', 'u_empty');`); err != nil {
		t.Fatal(err)
	}
	players, err := st.HistoryPlayers(ctx, "c_empty")
	if err != nil || players == nil || len(players) != 0 {
		t.Fatalf("empty community options = %+v, err = %v; want non-nil empty slice", players, err)
	}
	if _, err := st.HistoryPlayers(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown community error = %v, want ErrNotFound", err)
	}
}

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

// Empty results must marshal to JSON "[]", not "null" — a nil slice would crash
// clients that call .length on the decoded response.
func TestHistoryReturnsEmptySliceNotNil(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');
		INSERT INTO communities (id, name, mode, draws, code, owner_id) VALUES ('c_empty', 'Empty', 'open', 1, 'EMPT1', 'u_demo');`); err != nil {
		t.Fatal(err)
	}
	rows, err := st.History(ctx, "", "c_empty", "", 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rows == nil {
		t.Fatal("History() returned a nil slice for an empty community; want a non-nil empty slice")
	}
	if len(rows) != 0 {
		t.Fatalf("history for empty community = %d rows, want 0", len(rows))
	}
}

func TestJoinRequestsReturnsEmptySliceNotNil(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');
		INSERT INTO communities (id, name, mode, draws, code, owner_id) VALUES ('c_empty', 'Empty', 'open', 1, 'EMPT2', 'u_demo');`); err != nil {
		t.Fatal(err)
	}
	rows, err := st.JoinRequests(ctx, "c_empty")
	if err != nil {
		t.Fatal(err)
	}
	if rows == nil {
		t.Fatal("JoinRequests() returned a nil slice with no pending requests; want a non-nil empty slice")
	}
	if len(rows) != 0 {
		t.Fatalf("join requests for empty community = %d rows, want 0", len(rows))
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
