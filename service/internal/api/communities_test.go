package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"open-skill/service/internal/model"
	"open-skill/service/internal/store"
)

func newCommunityTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if _, err := st.DB().ExecContext(context.Background(), `
		INSERT INTO users (id, display) VALUES ('u_demo', 'IronCrest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_003', 'u_demo', 'IronCrest');`); err != nil {
		t.Fatal(err)
	}
	return st
}

func createCommunity(t *testing.T, h http.Handler, name string) model.Community {
	t.Helper()
	body, err := json.Marshal(map[string]any{"name": name, "mode": "open", "draws": true})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/communities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Demo-User", "u_demo")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create %q status = %d, want %d; body = %s", name, res.Code, http.StatusCreated, res.Body.String())
	}
	var c model.Community
	if err := json.Unmarshal(res.Body.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCreateCommunityGeneratesUniqueCodesForIdenticalNames(t *testing.T) {
	h := New(newCommunityTestStore(t), AuthConfig{Demo: true}).Handler()

	seen := map[string]bool{}
	for i := 0; i < 25; i++ {
		c := createCommunity(t, h, "testing")
		if len(c.Code) != codeLength {
			t.Fatalf("code %q length = %d, want %d", c.Code, len(c.Code), codeLength)
		}
		if seen[c.Code] {
			t.Fatalf("duplicate invite code %q for identically named communities", c.Code)
		}
		seen[c.Code] = true
	}
}

func TestCreateCommunityReturnsOwnerAsMember(t *testing.T) {
	h := New(newCommunityTestStore(t), AuthConfig{Demo: true}).Handler()
	created := createCommunity(t, h, "Fresh Squad")
	if created.Members != 1 {
		t.Fatalf("created members = %d, want 1", created.Members)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/communities", nil)
	req.Header.Set("X-Demo-User", "u_demo")
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	var list []model.Community
	if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Members != 1 {
		t.Fatalf("my communities = %+v, want one community with 1 member", list)
	}
}

func TestCommunityLeaderboardShowsCreatorWithoutMatches(t *testing.T) {
	h := New(newCommunityTestStore(t), AuthConfig{Demo: true}).Handler()
	created := createCommunity(t, h, "Fresh Squad")

	for _, mode := range []string{"1v1", "team", "ffa"} {
		req := httptest.NewRequest(http.MethodGet, "/api/leaderboard?mode="+mode+"&communityId="+created.ID, nil)
		req.Header.Set("X-Demo-User", "u_demo")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		var rows []store.LeaderboardRow
		if err := json.Unmarshal(res.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatalf("%s leaderboard = %+v, want the creator only", mode, rows)
		}
		if rows[0].PlayerID != "p_003" || !rows[0].IsYou || rows[0].Games != 0 || rows[0].Rank != 1 {
			t.Fatalf("%s leaderboard row = %+v, want p_003 ranked 1 with default stats", mode, rows[0])
		}
	}
}
