package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
)

func TestPlayerOwnershipMigrationPreservesExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE users (id TEXT PRIMARY KEY, display TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT (datetime('now')));
		CREATE TABLE players (id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id), handle TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL DEFAULT (datetime('now')));
		INSERT INTO users (id, display) VALUES ('u_legacy', 'Legacy');
		INSERT INTO players (id, user_id, handle) VALUES ('p_legacy', 'u_legacy', 'Legacy');`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_legacy', 'Legacy', 'LEGACY', 'u_legacy');
		INSERT INTO ratings (community_id, player_id, mode, games) VALUES ('c_legacy', 'p_legacy', '1v1', 1);
		INSERT INTO matches (id, community_id, mode) VALUES ('m_legacy', 'c_legacy', '1v1');
		INSERT INTO match_participants (match_id, player_id, score) VALUES ('m_legacy', 'p_legacy', 1);`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := applySchema(db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	st, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var owner string
	if err := st.DB().QueryRow(`SELECT user_id FROM players WHERE id = 'p_legacy'`).Scan(&owner); err != nil || owner != "u_legacy" {
		t.Fatalf("migrated owner = %q, err = %v", owner, err)
	}
	if _, err := st.DB().Exec(`INSERT INTO players (id, handle) VALUES ('p_unclaimed', 'Unclaimed')`); err != nil {
		t.Fatalf("insert ownerless player after migration: %v", err)
	}
	if _, err := st.DB().Exec(`INSERT INTO players (id, user_id, handle) VALUES ('p_invalid', 'missing', 'Invalid')`); err == nil {
		t.Fatal("migration disabled foreign-key enforcement")
	}
	var games int
	if err := st.DB().QueryRow(`SELECT games FROM ratings WHERE player_id = 'p_legacy'`).Scan(&games); err != nil || games != 1 {
		t.Fatalf("migrated ratings: games=%d err=%v", games, err)
	}
	if err := st.DB().QueryRow(`SELECT COUNT(*) FROM match_participants WHERE player_id = 'p_legacy'`).Scan(&games); err != nil || games != 1 {
		t.Fatalf("migrated history: participants=%d err=%v", games, err)
	}
}

func TestUnclaimedPlayerCommunityReporting(t *testing.T) {
	st, ctx := testCommunityStore(t)
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_unclaimed", Name: "Unclaimed", Mode: "open", Draws: true, Code: "UNC", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	player, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, "  Visiting   Player  ")
	if err != nil {
		t.Fatal(err)
	}
	if !player.Unclaimed || player.UserID != "" || player.Handle != "Visiting Player" {
		t.Fatalf("unclaimed player = %+v", player)
	}
	if _, err := st.CreateUnclaimedPlayer(ctx, "outsider", community.ID, "Forbidden"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("outsider creation = %v", err)
	}
	if _, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, player.Handle); !errors.Is(err, ErrHandleTaken) {
		t.Fatalf("duplicate creation = %v", err)
	}
	rows, err := st.Leaderboard(ctx, model.Mode1v1, 100, "u_demo", community.ID, "", "rank")
	if err != nil || len(rows) != 2 {
		t.Fatalf("roster = %+v, err = %v", rows, err)
	}
	for _, row := range rows {
		if row.PlayerID == player.ID && (!row.Unclaimed || row.IsYou || row.Games != 0) {
			t.Fatalf("unclaimed row = %+v", row)
		}
	}
	req := ReportMatchRequest{CommunityID: community.ID, Participants: []Participant{{PlayerID: "p_demo", Score: 1, Finish: -1}, {PlayerID: player.ID, Finish: -1}}}
	outcome, err := st.ReportMatchAsMember(ctx, "u_demo", req)
	if err != nil {
		t.Fatal(err)
	}
	profile, ratings, err := st.Profile(ctx, player.ID)
	if err != nil || !profile.Unclaimed || len(ratings) != 1 || ratings[0].Games != 1 {
		t.Fatalf("profile = %+v, ratings = %+v, err = %v", profile, ratings, err)
	}
	match, err := st.Match(ctx, outcome.MatchID)
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range match.Participants {
		if participant.PlayerID == player.ID && !participant.Unclaimed {
			t.Fatal("match lost unclaimed status")
		}
	}
	other, err := st.CreateCommunity(ctx, model.Community{ID: "c_elsewhere", Name: "Elsewhere", Mode: "open", Draws: true, Code: "ELSE", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	req.CommunityID = other.ID
	if _, err := st.ReportMatchAsMember(ctx, "u_demo", req); !errors.Is(err, ErrIneligibleParticipant) {
		t.Fatalf("cross-community report = %v", err)
	}
	rows, err = st.Leaderboard(ctx, model.Mode1v1, 100, "u_demo", other.ID, "", "rank")
	if err != nil || len(rows) != 1 {
		t.Fatalf("other roster = %+v, err = %v", rows, err)
	}
}

func TestUnclaimedPlayerClaimRequiresApprovalAndPreservesResults(t *testing.T) {
	st, ctx := testCommunityStore(t)
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_claim", Name: "Claim", Mode: "open", Draws: true, Code: "CLAIM", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	player, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, "Future Member")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ReportMatchAsMember(ctx, "u_demo", ReportMatchRequest{CommunityID: community.ID, Participants: []Participant{{PlayerID: "p_demo", Score: 1, Finish: -1}, {PlayerID: player.ID, Finish: -1}}}); err != nil {
		t.Fatal(err)
	}
	if err := st.EnsureUser(ctx, "u_future", "p_future", "Signed Up"); err != nil {
		t.Fatal(err)
	}
	if err := st.RequestPlayerClaim(ctx, "u_future", player.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("nonmember claim = %v", err)
	}
	if err := st.JoinCommunity(ctx, community.ID, "u_future"); err != nil {
		t.Fatal(err)
	}
	if err := st.RequestPlayerClaim(ctx, "u_future", player.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ApprovePlayerClaim(ctx, "u_future", player.ID, "u_future"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("self approval = %v", err)
	}
	claims, err := st.PlayerClaims(ctx, "u_demo", player.ID)
	if err != nil || len(claims) != 1 || claims[0].UserID != "u_future" {
		t.Fatalf("pending claims = %+v, err = %v", claims, err)
	}
	queue, err := st.CommunityPlayerClaims(ctx, "u_demo", community.ID)
	if err != nil || len(queue) != 1 || queue[0].PlayerID != player.ID || queue[0].TargetPlayerID != "p_future" {
		t.Fatalf("community claim queue = %+v, err = %v", queue, err)
	}
	if _, err := st.CommunityPlayerClaims(ctx, "u_future", community.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("member claim queue access = %v", err)
	}
	other, err := st.CreateCommunity(ctx, model.Community{ID: "c_claim_other", Name: "Other", Mode: "open", Code: "CLMOTHER", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	queue, err = st.CommunityPlayerClaims(ctx, "u_demo", other.ID)
	if err != nil || len(queue) != 0 {
		t.Fatalf("cross-community claim queue = %+v, err = %v", queue, err)
	}
	if _, err := st.ReportMatchAsMember(ctx, "u_demo", ReportMatchRequest{CommunityID: community.ID, Participants: []Participant{{PlayerID: "p_demo", Score: 0, Finish: -1}, {PlayerID: "p_future", Score: 1, Finish: -1}}}); err != nil {
		t.Fatal(err)
	}
	mergedID, err := st.ApprovePlayerClaim(ctx, "u_demo", player.ID, "u_future")
	if err != nil || mergedID != "p_future" {
		t.Fatalf("merged ID = %q, err = %v", mergedID, err)
	}
	queue, err = st.CommunityPlayerClaims(ctx, "u_demo", community.ID)
	if err != nil || len(queue) != 0 {
		t.Fatalf("approved claim remains in queue = %+v, err = %v", queue, err)
	}
	var leftovers int
	if err := st.DB().QueryRow(`SELECT COUNT(*) FROM players WHERE id = ?`, player.ID).Scan(&leftovers); err != nil || leftovers != 0 {
		t.Fatalf("standalone player remains: count=%d err=%v", leftovers, err)
	}
	if err := st.DB().QueryRow(`SELECT COUNT(*) FROM match_participants WHERE player_id = ?`, player.ID).Scan(&leftovers); err != nil || leftovers != 0 {
		t.Fatalf("old match references remain: count=%d err=%v", leftovers, err)
	}
	profile, ratings, err := st.Profile(ctx, player.ID)
	if err != nil || profile.ID != "p_future" || profile.Unclaimed || profile.UserID != "u_future" || profile.Handle != "Signed Up" {
		t.Fatalf("merged profile = %+v, err = %v", profile, err)
	}
	games := 0
	for _, row := range ratings {
		games += row.Games
	}
	if games != 2 {
		t.Fatalf("merged ratings = %+v, want two games", ratings)
	}
	if err := st.RequestPlayerClaim(ctx, "u_demo", player.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second claim = %v", err)
	}
	history, err := st.History(ctx, "p_future", community.ID, "", 20, 0)
	if err != nil || len(history) != 2 {
		t.Fatalf("claimed history = %+v, err = %v", history, err)
	}
	oldHistory, err := st.History(ctx, player.ID, community.ID, "", 20, 0)
	if err != nil || len(oldHistory) != 2 {
		t.Fatalf("old history link = %+v, err = %v", oldHistory, err)
	}
	for _, match := range history {
		for _, participant := range match.Participants {
			if participant.PlayerID == player.ID || participant.Unclaimed {
				t.Fatal("history contains the old unclaimed player")
			}
		}
	}
}

func TestPlayerClaimMergeRejectsSharedMatchesWithoutWrites(t *testing.T) {
	st, ctx := testCommunityStore(t)
	community, err := st.CreateCommunity(ctx, model.Community{ID: "c_conflict", Name: "Conflict", Mode: "open", Draws: true, Code: "CONFLICT", OwnerID: "u_demo"})
	if err != nil {
		t.Fatal(err)
	}
	player, err := st.CreateUnclaimedPlayer(ctx, "u_demo", community.ID, "Duplicate")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ReportMatchAsMember(ctx, "u_demo", ReportMatchRequest{CommunityID: community.ID, Participants: []Participant{{PlayerID: "p_demo", Score: 1, Finish: -1}, {PlayerID: player.ID, Finish: -1}}}); err != nil {
		t.Fatal(err)
	}
	if err := st.RequestPlayerClaim(ctx, "u_demo", player.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ApprovePlayerClaim(ctx, "u_demo", player.ID, "u_demo"); !errors.Is(err, ErrPlayerMergeConflict) {
		t.Fatalf("merge conflict = %v", err)
	}
	profile, ratings, err := st.Profile(ctx, player.ID)
	if err != nil || !profile.Unclaimed || len(ratings) != 1 || ratings[0].Games != 1 {
		t.Fatalf("conflicting merge changed profile: %+v %+v err=%v", profile, ratings, err)
	}
	claims, err := st.PlayerClaims(ctx, "u_demo", player.ID)
	if err != nil || len(claims) != 1 {
		t.Fatalf("conflicting merge removed pending claim: %+v err=%v", claims, err)
	}
}

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
