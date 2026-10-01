package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeVerifier map[string]Identity

func (f fakeVerifier) VerifyToken(_ context.Context, token string) (Identity, error) {
	if identity, ok := f[token]; ok {
		return identity, nil
	}
	return Identity{}, errors.New("bad token")
}

func tokenServer(t *testing.T) http.Handler {
	t.Helper()
	st := newReportTestStore(t)
	t.Cleanup(func() { st.Close() })
	return New(st, AuthConfig{Verifier: fakeVerifier{
		"alice-token": {UID: "alice", Name: "Alice"},
		"other-alice": {UID: "alice2", Name: "Alice"},
	}}).Handler()
}

func call(h http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	return res
}

func TestProtectedEndpointRequiresToken(t *testing.T) {
	h := tokenServer(t)
	if res := call(h, http.MethodGet, "/api/communities", nil); res.Code != http.StatusUnauthorized {
		t.Fatalf("no token: status = %d, want 401", res.Code)
	}
}

func TestDemoHeaderIgnoredWhenDemoAuthDisabled(t *testing.T) {
	h := tokenServer(t)
	res := call(h, http.MethodGet, "/api/me", map[string]string{"X-Demo-User": "u_demo"})
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("forged demo header: status = %d, want 401; body = %s", res.Code, res.Body.String())
	}
}

func TestInvalidTokenRejectedEvenOnPublicEndpoints(t *testing.T) {
	h := tokenServer(t)
	for _, header := range []string{"Bearer forged", "Basic abc"} {
		res := call(h, http.MethodGet, "/api/leaderboard?mode=1v1", map[string]string{"Authorization": header})
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("%q: status = %d, want 401", header, res.Code)
		}
	}
}

func TestBearerRejectedWithoutVerifier(t *testing.T) {
	st := newReportTestStore(t)
	defer st.Close()
	h := New(st, AuthConfig{Demo: true}).Handler()
	res := call(h, http.MethodGet, "/api/me", map[string]string{"Authorization": "Bearer alice-token"})
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Code)
	}
}

func TestAnonymousCanReadPublicEndpoints(t *testing.T) {
	h := tokenServer(t)
	for _, path := range []string{
		"/api/leaderboard?mode=1v1&communityId=c_test",
		"/api/leaderboard/modes?communityId=c_test",
		"/api/communities/c_test",
		"/api/communities/c_test/membership",
	} {
		if res := call(h, http.MethodGet, path, nil); res.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200; body = %s", path, res.Code, res.Body.String())
		}
	}
}

func TestTokenUserIsCreatedOnceWithUniqueHandles(t *testing.T) {
	h := tokenServer(t)
	me := func(token string) map[string]string {
		t.Helper()
		res := call(h, http.MethodGet, "/api/me", map[string]string{"Authorization": "Bearer " + token})
		if res.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", res.Code, res.Body.String())
		}
		var body map[string]string
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body
	}

	first := me("alice-token")
	if first["user"] != "fb_alice" || first["playerId"] != "p_alice" || first["handle"] != "Alice" {
		t.Fatalf("first sign-in = %#v", first)
	}
	if again := me("alice-token"); again["playerId"] != first["playerId"] {
		t.Fatalf("second sign-in created a new player: %#v", again)
	}
	other := me("other-alice")
	if other["user"] != "fb_alice2" || other["handle"] == "Alice" || !strings.HasPrefix(other["handle"], "Alice-") {
		t.Fatalf("same display name should get a distinct handle, got %#v", other)
	}
}

func TestCORSAllowsDemoHeaderOnlyInDemoMode(t *testing.T) {
	t.Setenv("OPEN_SKILL_ALLOWED_ORIGINS", "http://localhost:5173")
	noop := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})
	for demo, want := range map[bool]bool{true: true, false: false} {
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/api/me", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		cors(noop, demo).ServeHTTP(res, req)
		if got := strings.Contains(res.Header().Get("Access-Control-Allow-Headers"), "X-Demo-User"); got != want {
			t.Fatalf("demo=%v: X-Demo-User allowed = %v, want %v", demo, got, want)
		}
	}
}

func TestPlayerHandle(t *testing.T) {
	cases := map[string]string{
		"":                                 "player",
		"  Ada   Lovelace ":                "Ada Lovelace",
		"An Extremely Long Display Name X": "An Extremely Long Displa",
	}
	for in, want := range cases {
		if got := playerHandle(in); got != want {
			t.Errorf("playerHandle(%q) = %q, want %q", in, got, want)
		}
	}
}
