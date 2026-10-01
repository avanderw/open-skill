package store

import (
	"context"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func testCommunityStore(t *testing.T) (*Store, context.Context) {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	ctx := context.Background()
	_, err = st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_demo', 'u_demo', 'IronCrest')`)
	if err != nil {
		t.Fatal(err)
	}
	return st, ctx
}

func TestNewCommunityLeaderboardShowsOwnerWithDefaultStats(t *testing.T) {
	st, ctx := testCommunityStore(t)
	other, err := st.CreateCommunity(ctx, model.Community{ID: "c_other", Name: "Other", Mode: "open", Draws: true, Code: "OTHER", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(ctx, `UPDATE ratings SET mu = 29.75, games = 20 WHERE community_id = ? AND player_id = 'p_demo' AND mode = '1v1'`, other.ID); err != nil {
		t.Fatal(err)
	}
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_new", Name: "Fresh", Mode: "open", Draws: true, Code: "FRESH", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}

	rows, err := st.Leaderboard(ctx, model.Mode1v1, 10, "u_demo", community.ID, "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PlayerID != "p_demo" || rows[0].Games != 0 {
		t.Fatalf("new community leaderboard = %+v, want only the owner with 0 games", rows)
	}
}

func TestCommunityLeaderboardShowsOnlyCommunityMembers(t *testing.T) {
	st, ctx := testCommunityStore(t)
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_one", Name: "One", Mode: "open", Draws: true, Code: "ONE", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(ctx, `UPDATE ratings SET mu = 29.75, games = 20 WHERE community_id = ? AND player_id = 'p_demo' AND mode = '1v1'`, community.ID); err != nil {
		t.Fatal(err)
	}

	rows, err := st.Leaderboard(ctx, model.Mode1v1, 10, "u_demo", community.ID, "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PlayerID != "p_demo" || rows[0].Games != 20 {
		t.Fatalf("community leaderboard = %+v, want only p_demo with 20 games", rows)
	}
}

func TestGlobalLeaderboardDeduplicatesPlayersAcrossCommunities(t *testing.T) {
	st, ctx := testCommunityStore(t)
	one, err := st.CreateCommunity(ctx, model.Community{ID: "c_one", Name: "One", Mode: "open", Draws: true, Code: "ONE", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	two, err := st.CreateCommunity(ctx, model.Community{ID: "c_two", Name: "Two", Mode: "open", Draws: true, Code: "TWO", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(ctx, `UPDATE ratings SET mu = 28, games = 7 WHERE community_id = ? AND player_id = 'p_demo' AND mode = '1v1'`, one.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(ctx, `UPDATE ratings SET mu = 30, games = 11 WHERE community_id = ? AND player_id = 'p_demo' AND mode = '1v1'`, two.ID); err != nil {
		t.Fatal(err)
	}

	rows, err := st.Leaderboard(ctx, model.Mode1v1, 10, "u_demo", "", "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PlayerID != "p_demo" || rows[0].Games != 18 {
		t.Fatalf("global leaderboard = %+v, want one aggregated row for p_demo", rows)
	}
}

func TestCommunityLeaderboardIncludesMembersWithoutRatings(t *testing.T) {
	st, ctx := testCommunityStore(t)
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_guest', 'GuestPlayer');
		INSERT INTO players (id, user_id, handle) VALUES ('p_guest', 'u_guest', 'GuestPlayer')`); err != nil {
		t.Fatal(err)
	}
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_one", Name: "One", Mode: "open", Draws: true, Code: "ONE", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.JoinCommunity(ctx, community.ID, "u_guest"); err != nil {
		t.Fatal(err)
	}

	rows, err := st.Leaderboard(ctx, model.Mode1v1, 10, "u_guest", community.ID, "", "rank")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("community leaderboard = %+v, want both members", rows)
	}
	var guest *LeaderboardRow
	for i := range rows {
		if rows[i].PlayerID == "p_guest" {
			guest = &rows[i]
		}
	}
	if guest == nil || !guest.IsYou || guest.Games != 0 {
		t.Fatalf("guest row = %+v, want the joined member with default stats", guest)
	}
}
