package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"open-skill/service/internal/store"
)

func TestReportMatchRejectsParticipantOutsideCommunity(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `
		INSERT INTO users (id, display) VALUES ('u_demo', 'Owner'), ('u_guest', 'Guest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_demo', 'u_demo', 'Owner'), ('p_guest', 'u_guest', 'Guest');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_demo');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_demo', 'owner');`); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(map[string]any{
		"communityId": "c_test",
		"participants": []map[string]any{
			{"playerId": "p_demo", "score": 1},
			{"playerId": "p_guest", "score": 0},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/matches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Demo-User", "u_demo")
	res := httptest.NewRecorder()
	New(st).Handler().ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusUnprocessableEntity, res.Body.String())
	}
}

func TestReportMatchReturnsNotFoundForUnknownPlayer(t *testing.T) {
	st := newReportTestStore(t)
	defer st.Close()

	body := []byte(`{"communityId":"c_test","participants":[{"playerId":"p_demo","score":1},{"playerId":"p_missing","score":0}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/matches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Demo-User", "u_demo")
	res := httptest.NewRecorder()
	New(st).Handler().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusNotFound, res.Body.String())
	}
}

func TestReportMatchReturnsNotFoundForUnknownCommunity(t *testing.T) {
	st := newReportTestStore(t)
	defer st.Close()

	body := []byte(`{"communityId":"c_missing","participants":[{"playerId":"p_demo","score":1},{"playerId":"p_demo_2","score":0}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/matches", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Demo-User", "u_demo")
	res := httptest.NewRecorder()
	New(st).Handler().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusNotFound, res.Body.String())
	}
}

func TestGetCommunityRejectsUnknownDemoUser(t *testing.T) {
	st := newReportTestStore(t)
	defer st.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/communities/c_test", nil)
	req.Header.Set("X-Demo-User", "u_unknown")
	res := httptest.NewRecorder()
	New(st).Handler().ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusUnauthorized, res.Body.String())
	}
}

func TestEmailDemoUserIsMaterialized(t *testing.T) {
	st := newReportTestStore(t)
	defer st.Close()

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("X-Demo-User", "Tester@Example.com")
	res := httptest.NewRecorder()
	New(st).Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusOK, res.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["user"] != "tester@example.com" || body["handle"] != "tester@example.com" || body["playerId"] != "tester@example.com" {
		t.Fatalf("body = %#v, want materialized email identity", body)
	}
}

func TestCORSUsesConfiguredOrigins(t *testing.T) {
	t.Setenv("OPEN_SKILL_ALLOWED_ORIGINS", "http://localhost:5173, https://example.test")
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("Origin", "https://example.test")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://example.test" {
		t.Fatalf("allowed origin = %q, want configured origin", got)
	}
}

func TestCORSOmitsDisallowedOrigins(t *testing.T) {
	t.Setenv("OPEN_SKILL_ALLOWED_ORIGINS", "http://localhost:5173")
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("Origin", "https://untrusted.test")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("disallowed origin = %q, want no allow-origin header", got)
	}
}

func newReportTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB().ExecContext(context.Background(), `
		INSERT INTO users (id, display) VALUES ('u_demo', 'Owner');
		INSERT INTO players (id, user_id, handle) VALUES ('p_demo', 'u_demo', 'Owner');
		INSERT INTO communities (id, name, code, owner_id) VALUES ('c_test', 'Test', 'TEST1', 'u_demo');
		INSERT INTO community_members (community_id, user_id, role) VALUES ('c_test', 'u_demo', 'owner');`); err != nil {
		st.Close()
		t.Fatal(err)
	}
	return st
}
