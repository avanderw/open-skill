package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"path/filepath"
	"sync/atomic"
	"testing"

	"modernc.org/sqlite"
)

// countingConnector counts statements, BEGINs and COMMIT/ROLLBACKs as a proxy for
// network round trips against a remote libSQL database.
type countingConnector struct {
	dsn   string
	count *atomic.Int64
}

func (c countingConnector) Connect(context.Context) (driver.Conn, error) {
	conn, err := (&sqlite.Driver{}).Open(c.dsn)
	if err != nil {
		return nil, err
	}
	return countingConn{Conn: conn, count: c.count}, nil
}

func (c countingConnector) Driver() driver.Driver { return &sqlite.Driver{} }

type countingConn struct {
	driver.Conn
	count *atomic.Int64
}

func (c countingConn) Prepare(query string) (driver.Stmt, error) {
	c.count.Add(1)
	return c.Conn.Prepare(query)
}

func (c countingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.count.Add(1)
	return c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c countingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.count.Add(1)
	return c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (c countingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	c.count.Add(1)
	tx, err := c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return countingTx{Tx: tx, count: c.count}, nil
}

type countingTx struct {
	driver.Tx
	count *atomic.Int64
}

func (t countingTx) Commit() error   { t.count.Add(1); return t.Tx.Commit() }
func (t countingTx) Rollback() error { t.count.Add(1); return t.Tx.Rollback() }

func openCountingStore(t *testing.T) (*Store, *atomic.Int64) {
	t.Helper()
	count := &atomic.Int64{}
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=foreign_keys(1)"
	st, err := open(sql.OpenDB(countingConnector{dsn: dsn, count: count}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st, count
}

func seedRoundTripCommunity(t *testing.T, st *Store) {
	t.Helper()
	if _, err := st.DB().Exec(`
		INSERT INTO users (id, display) VALUES ('u1', 'One'), ('u2', 'Two'), ('u3', 'Three'), ('u4', 'Four');
		INSERT INTO players (id, user_id, handle) VALUES ('p1', 'u1', 'One'), ('p2', 'u2', 'Two'), ('p3', 'u3', 'Three'), ('p4', 'u4', 'Four');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_rt', 'RoundTrips', 'RT001', 'u1');
		INSERT INTO community_members (community_id, user_id, role) VALUES
			('c_rt', 'u1', 'owner'), ('c_rt', 'u2', 'admin'), ('c_rt', 'u3', 'member'), ('c_rt', 'u4', 'member');`); err != nil {
		t.Fatal(err)
	}
}

// reportMixedMatches reports n matches cycling through 1v1, team and ffa.
func reportMixedMatches(t *testing.T, ctx context.Context, st *Store, n int) []string {
	t.Helper()
	var ids []string
	for i := 0; i < n; i++ {
		var parts []Participant
		switch i % 3 {
		case 0:
			parts = []Participant{{PlayerID: "p1", Score: float64(i % 2)}, {PlayerID: "p2", Score: float64((i + 1) % 2)}}
		case 1:
			parts = []Participant{
				{PlayerID: "p1", Team: "A", Score: 3}, {PlayerID: "p3", Team: "A", Score: 3},
				{PlayerID: "p2", Team: "B", Score: float64(i % 5)}, {PlayerID: "p4", Team: "B", Score: float64(i % 5)},
			}
		default:
			parts = []Participant{{PlayerID: "p4", Finish: 0}, {PlayerID: "p3", Finish: 1}, {PlayerID: "p2", Finish: 2}}
		}
		out, err := st.ReportMatch(ctx, ReportMatchRequest{
			CommunityID:  "c_rt",
			Participants: parts,
			PlayedAt:     fmt.Sprintf("2026-01-01 00:%02d:%02d", i/60, i%60),
		})
		if err != nil {
			t.Fatalf("report match %d: %v", i, err)
		}
		ids = append(ids, out.MatchID)
	}
	return ids
}

type ratingSnapshot map[string][]float64

func snapshotRatings(t *testing.T, st *Store) ratingSnapshot {
	t.Helper()
	out := ratingSnapshot{}
	rows, err := st.DB().Query(`SELECT player_id, mode, mu, sigma, games, wins, losses, draws, COALESCE(last_played, '') FROM ratings WHERE community_id = 'c_rt'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var player, mode, lastPlayed string
		var mu, sigma float64
		var games, wins, losses, draws int
		if err := rows.Scan(&player, &mode, &mu, &sigma, &games, &wins, &losses, &draws, &lastPlayed); err != nil {
			t.Fatal(err)
		}
		out["rating "+player+" "+mode+" "+lastPlayed] = []float64{mu, sigma, float64(games), float64(wins), float64(losses), float64(draws)}
	}
	rows, err = st.DB().Query(`SELECT match_id, player_id, rating_delta FROM match_participants`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var match, player string
		var delta float64
		if err := rows.Scan(&match, &player, &delta); err != nil {
			t.Fatal(err)
		}
		out["delta "+match+" "+player] = []float64{delta}
	}
	return out
}

func assertSnapshotsEqual(t *testing.T, got, want ratingSnapshot) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("snapshot has %d entries, want %d", len(got), len(want))
	}
	for key, wantValues := range want {
		gotValues, ok := got[key]
		if !ok {
			t.Fatalf("missing %s", key)
		}
		for i := range wantValues {
			if math.Abs(gotValues[i]-wantValues[i]) > 1e-9 {
				t.Fatalf("%s = %v, want %v", key, gotValues, wantValues)
			}
		}
	}
}

func TestRecomputeReproducesReportedRatings(t *testing.T) {
	st, _ := openCountingStore(t)
	seedRoundTripCommunity(t, st)
	ctx := context.Background()
	reportMixedMatches(t, ctx, st, 12)
	want := snapshotRatings(t, st)

	tx, err := st.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := recomputeCommunity(ctx, tx, "c_rt"); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	assertSnapshotsEqual(t, snapshotRatings(t, st), want)
}

func TestDeleteMatchReplaysRemainingMatches(t *testing.T) {
	st, _ := openCountingStore(t)
	seedRoundTripCommunity(t, st)
	ctx := context.Background()
	reportMixedMatches(t, ctx, st, 1)
	want := snapshotRatings(t, st)

	ids := reportMixedMatches(t, ctx, st, 1)
	if _, err := st.DeleteMatch(ctx, "u1", ids[0]); err != nil {
		t.Fatal(err)
	}
	assertSnapshotsEqual(t, snapshotRatings(t, st), want)
}

func TestRoundTripBudgets(t *testing.T) {
	st, count := openCountingStore(t)
	seedRoundTripCommunity(t, st)
	ctx := context.Background()
	ids := reportMixedMatches(t, ctx, st, 30)

	measure := func(name string, budget int64, fn func() error) {
		t.Helper()
		count.Store(0)
		if err := fn(); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := count.Load()
		t.Logf("%s: %d round trips", name, got)
		if got > budget {
			t.Errorf("%s used %d round trips, budget %d", name, got, budget)
		}
	}

	measure("ReportMatchAsMember team", 8, func() error {
		_, err := st.ReportMatchAsMember(ctx, "u1", ReportMatchRequest{CommunityID: "c_rt", Participants: []Participant{
			{PlayerID: "p1", Team: "A", Score: 2}, {PlayerID: "p3", Team: "A", Score: 2},
			{PlayerID: "p2", Team: "B", Score: 1}, {PlayerID: "p4", Team: "B", Score: 1},
		}})
		return err
	})
	// Replays must not scale with the number of matches in the community.
	measure("UpdateMatch (30 matches)", 13, func() error {
		_, err := st.UpdateMatch(ctx, "u1", ids[0], []Participant{{PlayerID: "p1", Score: 0}, {PlayerID: "p2", Score: 1}}, "")
		return err
	})
	measure("DeleteMatch (30 matches)", 10, func() error {
		_, err := st.DeleteMatch(ctx, "u1", ids[1])
		return err
	})
	measure("Membership", 1, func() error {
		_, err := st.Membership(ctx, "c_rt", "u3")
		return err
	})
	measure("CanAdmin", 1, func() error { return st.CanAdmin(ctx, "c_rt", "u2") })
	measure("SetMemberRole", 2, func() error { return st.SetMemberRole(ctx, "c_rt", "u1", "u3", "admin") })
	measure("RemoveMember", 2, func() error { return st.RemoveMember(ctx, "c_rt", "u2", "u4") })
}
