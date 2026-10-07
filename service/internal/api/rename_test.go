package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"open-skill/service/internal/store"
)

func TestRenamePlayer(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.DB().ExecContext(context.Background(), `
		INSERT INTO users (id, display) VALUES ('u_demo', 'Andrew'), ('u_guest', 'Guest');
		INSERT INTO players (id, user_id, handle) VALUES ('p_demo', 'u_demo', 'Andrew v'), ('p_guest', 'u_guest', 'Guest');`); err != nil {
		t.Fatal(err)
	}
	h := New(st, AuthConfig{Demo: true}).Handler()
	put := func(user, player, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/api/profile/"+player, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Demo-User", user)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		return res
	}
	cases := []struct {
		name, user, player, body string
		want                     int
	}{
		{"owner renames", "u_demo", "p_demo", `{"handle":"  Andrew   van der Westhuizen "}`, http.StatusOK},
		{"other user forbidden", "u_guest", "p_demo", `{"handle":"Hacked"}`, http.StatusForbidden},
		{"taken name conflicts", "u_demo", "p_demo", `{"handle":"Guest"}`, http.StatusConflict},
		{"blank rejected", "u_demo", "p_demo", `{"handle":"   "}`, http.StatusBadRequest},
		{"too long rejected", "u_demo", "p_demo", `{"handle":"` + strings.Repeat("x", 65) + `"}`, http.StatusBadRequest},
		{"unknown player", "u_demo", "p_nope", `{"handle":"X"}`, http.StatusNotFound},
	}
	for _, c := range cases {
		if res := put(c.user, c.player, c.body); res.Code != c.want {
			t.Errorf("%s: status = %d, want %d; body = %s", c.name, res.Code, c.want, res.Body.String())
		}
	}
	player, _, err := st.Profile(context.Background(), "p_demo")
	if err != nil || player.Handle != "Andrew van der Westhuizen" {
		t.Errorf("handle = %v, err = %v", player, err)
	}
}
