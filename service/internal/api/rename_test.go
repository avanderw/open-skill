package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"open-skill/service/internal/model"
	"open-skill/service/internal/store"
)

func TestUnclaimedPlayerHTTPWorkflow(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := New(st, AuthConfig{Verifier: fakeVerifier{
		"owner-token":   {UID: "owner", Name: "Owner"},
		"visitor-token": {UID: "visitor", Name: "Visitor"},
	}}).Handler()
	request := func(method, path, body, token string, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != want {
			t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, res.Code, want, res.Body.String())
		}
		return res
	}
	request(http.MethodGet, "/api/me", "", "owner-token", http.StatusOK)
	request(http.MethodGet, "/api/me", "", "visitor-token", http.StatusOK)
	request(http.MethodPost, "/api/communities", `{"name":"Unclaimed workflow","mode":"open","draws":true}`, "", http.StatusUnauthorized)
	response := request(http.MethodPost, "/api/communities", `{"name":"Unclaimed workflow","mode":"open","draws":true}`, "owner-token", http.StatusCreated)
	var community model.Community
	if err := json.Unmarshal(response.Body.Bytes(), &community); err != nil {
		t.Fatal(err)
	}
	path := "/api/communities/" + community.ID + "/players"
	request(http.MethodPost, path, `{"handle":"No account"}`, "", http.StatusUnauthorized)
	request(http.MethodPost, path, `{"handle":"No account"}`, "visitor-token", http.StatusForbidden)
	request(http.MethodPost, path, `{"handle":" "}`, "owner-token", http.StatusBadRequest)
	request(http.MethodPost, path, `{"handle":"`+strings.Repeat("x", 65)+`"}`, "owner-token", http.StatusBadRequest)
	response = request(http.MethodPost, path, `{"handle":"  No   account "}`, "owner-token", http.StatusCreated)
	var player model.Player
	if err := json.Unmarshal(response.Body.Bytes(), &player); err != nil {
		t.Fatal(err)
	}
	if !player.Unclaimed || player.UserID != "" || player.CommunityID != community.ID {
		t.Fatalf("player = %+v", player)
	}
	request(http.MethodPost, path, `{"handle":"No account"}`, "owner-token", http.StatusConflict)
	request(http.MethodPut, "/api/profile/"+player.ID, `{"handle":"Hijacked"}`, "visitor-token", http.StatusForbidden)
	request(http.MethodPost, "/api/matches", `{"communityId":"`+community.ID+`","participants":[{"playerId":"p_owner","score":1,"finish":-1},{"playerId":"`+player.ID+`","score":0,"finish":-1}]}`, "owner-token", http.StatusCreated)
	claimPath := "/api/profile/" + player.ID + "/claims"
	request(http.MethodPost, claimPath, "", "visitor-token", http.StatusForbidden)
	request(http.MethodPost, "/api/communities/"+community.Code+"/join", "", "visitor-token", http.StatusOK)
	request(http.MethodPost, claimPath, "", "visitor-token", http.StatusCreated)
	request(http.MethodGet, "/api/communities/"+community.ID+"/player-claims", "", "", http.StatusUnauthorized)
	request(http.MethodGet, "/api/communities/"+community.ID+"/player-claims", "", "visitor-token", http.StatusForbidden)
	response = request(http.MethodGet, "/api/communities/"+community.ID+"/player-claims", "", "owner-token", http.StatusOK)
	var queue []store.CommunityPlayerClaim
	if err := json.Unmarshal(response.Body.Bytes(), &queue); err != nil || len(queue) != 1 || queue[0].PlayerID != player.ID {
		t.Fatalf("claim queue = %+v, err = %v", queue, err)
	}
	request(http.MethodPost, claimPath+"/fb_visitor/approve", "", "visitor-token", http.StatusForbidden)
	response = request(http.MethodPost, claimPath+"/fb_visitor/approve", "", "owner-token", http.StatusOK)
	var approval struct {
		PlayerID string `json:"playerId"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &approval); err != nil || approval.PlayerID != "p_visitor" {
		t.Fatalf("merge approval = %+v, err = %v", approval, err)
	}
	response = request(http.MethodGet, "/api/communities/"+community.ID+"/player-claims", "", "owner-token", http.StatusOK)
	if err := json.Unmarshal(response.Body.Bytes(), &queue); err != nil || len(queue) != 0 {
		t.Fatalf("community claims after approval = %+v, err = %v", queue, err)
	}
	var remainingClaims int
	if err := st.DB().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM player_claim_requests WHERE player_id = ?", player.ID).Scan(&remainingClaims); err != nil || remainingClaims != 0 {
		t.Fatalf("claims after approval = %d, err = %v", remainingClaims, err)
	}
	response = request(http.MethodGet, "/api/profile/"+player.ID, "", "visitor-token", http.StatusOK)
	var profile struct {
		Player  model.Player   `json:"player"`
		Ratings []model.Rating `json:"ratings"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	games := 0
	for _, row := range profile.Ratings {
		games += row.Games
	}
	if profile.Player.ID != "p_visitor" || profile.Player.Unclaimed || profile.Player.UserID != "fb_visitor" || games != 1 {
		t.Fatalf("claimed profile = %+v", profile)
	}
}

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
