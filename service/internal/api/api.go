// Package api exposes the HTTP server for Open-Skill.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"open-skill/service/internal/model"
	"open-skill/service/internal/notify"
	"open-skill/service/internal/store"
)

// Server is the HTTP API.
type Server struct {
	store    *store.Store
	auth     AuthConfig
	users    knownUsers
	stubUser string // demo-mode fallback user when no demo header is supplied
	notifier *notify.Notifier
}

// New builds an API server for the given store.
func New(st *store.Store, auth AuthConfig) *Server {
	return &Server{store: st, auth: auth, stubUser: "u_demo", notifier: notify.New()}
}

// Handler returns the root http.Handler with all routes mounted.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/me", s.handleMe)
	mux.HandleFunc("GET /api/communities", s.handleListCommunities)
	mux.HandleFunc("GET /api/communities/browse", s.handleBrowseCommunities)
	mux.HandleFunc("POST /api/communities", s.handleCreateCommunity)
	mux.HandleFunc("GET /api/communities/{code}", s.handleGetCommunityByCode)
	mux.HandleFunc("GET /api/communities/{code}/membership", s.handleMembership)
	mux.HandleFunc("PUT /api/communities/{code}", s.handleUpdateCommunity)
	mux.HandleFunc("DELETE /api/communities/{code}", s.handleDeleteCommunity)
	mux.HandleFunc("GET /api/communities/{code}/members", s.handleMembers)
	mux.HandleFunc("POST /api/communities/{code}/members/{user}/approve", s.handleApproveMember)
	mux.HandleFunc("POST /api/communities/{code}/members/{user}/promote", s.handlePromoteMember)
	mux.HandleFunc("POST /api/communities/{code}/members/{user}/demote", s.handleDemoteMember)
	mux.HandleFunc("DELETE /api/communities/{code}/members/{user}", s.handleRemoveMember)
	mux.HandleFunc("GET /api/communities/{code}/integrations", s.handleIntegration)
	mux.HandleFunc("PUT /api/communities/{code}/integrations", s.handleUpdateIntegration)
	mux.HandleFunc("POST /api/communities/{code}/join", s.handleJoinCommunity)
	mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("GET /api/leaderboard/modes", s.handleLeaderboardModes)
	mux.HandleFunc("GET /api/profile/{id}", s.handleProfile)
	mux.HandleFunc("GET /api/history", s.handleHistory)
	mux.HandleFunc("POST /api/matches", s.handleReportMatch)
	mux.HandleFunc("GET /api/matches/{id}", s.handleGetMatch)
	mux.HandleFunc("PUT /api/matches/{id}", s.handleUpdateMatch)
	mux.HandleFunc("DELETE /api/matches/{id}", s.handleDeleteMatch)
	return logMiddleware(cors(mux, s.auth.Demo))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "open-skill"})
}

// handleMe returns the fixed demo identity, documented as the FFI for the auth stub.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	player, err := s.store.PrimaryPlayerForUser(r.Context(), actor)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "load profile: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": actor, "handle": player.Handle, "playerId": player.ID})
}

func (s *Server) handleListCommunities(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	cs, err := s.store.ListCommunitiesForUser(r.Context(), actor)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "list communities: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func (s *Server) handleBrowseCommunities(w http.ResponseWriter, r *http.Request) {
	actor, err := s.optionalActor(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	cs, err := s.store.BrowseCommunities(r.Context(), actor)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "browse communities: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func (s *Server) handleCreateCommunity(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	var req struct {
		Name  string `json:"name"`
		Mode  string `json:"mode"`
		Draws bool   `json:"draws"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request: %v", err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeErr(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Mode != "open" && req.Mode != "approval" {
		req.Mode = "open"
	}
	c := model.Community{
		ID:      newID("c"),
		Name:    req.Name,
		Mode:    req.Mode,
		Draws:   req.Draws,
		OwnerID: actor,
	}
	// Codes are random, so a collision is unlikely but possible; retry with a
	// fresh code until the unique index accepts one.
	var created *model.Community
	for attempt := 0; attempt < 8; attempt++ {
		code, codeErr := genCode()
		if codeErr != nil {
			writeErr(w, http.StatusInternalServerError, "generate invite code: %v", codeErr)
			return
		}
		c.Code = code
		created, err = s.store.CreateCommunity(r.Context(), c)
		if err == nil {
			break
		}
		if !errors.Is(err, store.ErrCodeTaken) {
			writeErr(w, http.StatusInternalServerError, "create community: %v", err)
			return
		}
	}
	if created == nil {
		writeErr(w, http.StatusInternalServerError, "create community: could not allocate an unused invite code")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetCommunityByCode(w http.ResponseWriter, r *http.Request) {
	actor, err := s.optionalActor(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	key := r.PathValue("code")
	c, err := s.store.FindCommunityByID(r.Context(), key)
	if err != nil {
		c, err = s.store.FindCommunityByCode(r.Context(), strings.ToUpper(key))
	}
	if err != nil {
		writeErr(w, http.StatusNotFound, "community not found: %v", err)
		return
	}
	if status, membershipErr := s.store.Membership(r.Context(), c.ID, actor); membershipErr == nil {
		c.Membership = status
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleJoinCommunity(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	code := strings.ToUpper(r.PathValue("code"))
	c, err := s.store.FindCommunityByCode(r.Context(), code)
	if err != nil {
		writeErr(w, http.StatusNotFound, "community not found: %v", err)
		return
	}
	if err := s.store.JoinCommunity(r.Context(), c.ID, actor); err != nil {
		writeErr(w, http.StatusInternalServerError, "join community: %v", err)
		return
	}
	status, _ := s.store.Membership(r.Context(), c.ID, actor)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": status, "communityId": c.ID, "name": c.Name, "mode": c.Mode})
}

func (s *Server) handleMembership(w http.ResponseWriter, r *http.Request) {
	actor, err := s.optionalActor(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	status, canManage, err := s.store.MembershipAccess(r.Context(), r.PathValue("code"), actor)
	if err != nil {
		writeErr(w, http.StatusNotFound, "membership: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"communityId": r.PathValue("code"), "status": status, "canManage": canManage})
}

func (s *Server) handleUpdateCommunity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("code")
	if _, ok := s.requireAdmin(w, r, id); !ok {
		return
	}
	var req struct {
		Name  string `json:"name"`
		Mode  string `json:"mode"`
		Draws *bool  `json:"draws"`
	}
	if err := readJSON(r, &req); err != nil || strings.TrimSpace(req.Name) == "" || (req.Mode != "open" && req.Mode != "approval") || req.Draws == nil {
		writeErr(w, http.StatusBadRequest, "name, valid mode, and draws are required")
		return
	}
	if err := s.store.UpdateCommunity(r.Context(), id, strings.TrimSpace(req.Name), req.Mode, *req.Draws); err != nil {
		writeErr(w, http.StatusBadRequest, "update community: %v", err)
		return
	}
	c, err := s.store.FindCommunityByID(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "community not found: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleDeleteCommunity(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	if err := s.store.DeleteCommunity(r.Context(), actor, r.PathValue("code")); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "only the owner can delete a community")
		} else if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "%v", err)
		} else {
			writeErr(w, http.StatusInternalServerError, "delete community: %v", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleApproveMember(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r, r.PathValue("code")); !ok {
		return
	}
	if err := s.store.ApproveJoinRequest(r.Context(), r.PathValue("code"), r.PathValue("user")); err != nil {
		writeErr(w, http.StatusBadRequest, "approve member: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "communityId": r.PathValue("code"), "userId": r.PathValue("user"), "status": "member"})
}

func (s *Server) handlePromoteMember(w http.ResponseWriter, r *http.Request) {
	s.handleRoleMutation(w, r, "admin")
}
func (s *Server) handleDemoteMember(w http.ResponseWriter, r *http.Request) {
	s.handleRoleMutation(w, r, "member")
}
func (s *Server) handleRoleMutation(w http.ResponseWriter, r *http.Request, role string) {
	actor, actorErr := s.actorID(r)
	if actorErr != nil {
		writeErr(w, http.StatusUnauthorized, "%v", actorErr)
		return
	}
	if err := s.store.SetMemberRole(r.Context(), r.PathValue("code"), actor, r.PathValue("user"), role); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "%v", err)
		} else if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "%v", err)
		} else {
			writeErr(w, http.StatusBadRequest, "%v", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "role": role})
}
func (s *Server) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	actor, actorErr := s.actorID(r)
	if actorErr != nil {
		writeErr(w, http.StatusUnauthorized, "%v", actorErr)
		return
	}
	if err := s.store.RemoveMember(r.Context(), r.PathValue("code"), actor, r.PathValue("user")); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "%v", err)
		} else if errors.Is(err, store.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "%v", err)
		} else {
			writeErr(w, http.StatusBadRequest, "%v", err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMembers(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r, r.PathValue("code")); !ok {
		return
	}
	rows, err := s.store.JoinRequests(r.Context(), r.PathValue("code"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "members: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
func (s *Server) handleIntegration(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r, r.PathValue("code")); !ok {
		return
	}
	i, err := s.store.Integration(r.Context(), r.PathValue("code"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "integration: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, maskedIntegration(*i))
}

// maskedIntegration hides webhook secrets; they are write-only after saving.
func maskedIntegration(i store.CommunityIntegration) store.CommunityIntegration {
	i.Discord = notify.Mask(i.Discord)
	i.Webhook = notify.Mask(i.Webhook)
	return i
}
func (s *Server) handleUpdateIntegration(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r, r.PathValue("code")); !ok {
		return
	}
	var i store.CommunityIntegration
	if err := readJSON(r, &i); err != nil {
		writeErr(w, http.StatusBadRequest, "bad integration: %v", err)
		return
	}
	i.CommunityID = r.PathValue("code")
	current, err := s.store.Integration(r.Context(), i.CommunityID)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "update integration: %v", err)
		return
	}
	// Submitting the masked value back means "unchanged".
	if i.Discord == notify.Mask(current.Discord) {
		i.Discord = current.Discord
	}
	if i.Webhook == notify.Mask(current.Webhook) {
		i.Webhook = current.Webhook
	}
	i.Discord, i.Webhook = strings.TrimSpace(i.Discord), strings.TrimSpace(i.Webhook)
	if err := notify.ValidateURL(notify.Discord, i.Discord); err != nil {
		writeErr(w, http.StatusBadRequest, "discord URL: %v", err)
		return
	}
	if err := notify.ValidateURL(notify.Webhook, i.Webhook); err != nil {
		writeErr(w, http.StatusBadRequest, "webhook URL: %v", err)
		return
	}
	if err := s.store.UpdateIntegration(r.Context(), i); err != nil {
		writeErr(w, http.StatusBadRequest, "update integration: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, maskedIntegration(i))
}

// announceMatch notifies the community's integrations in the background so a
// slow endpoint never delays or fails match reporting.
func (s *Server) announceMatch(communityID string, outcome *store.ReportOutcome) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		integration, err := s.store.Integration(ctx, communityID)
		if err != nil || (integration.Discord == "" && integration.Webhook == "") {
			return
		}
		community, err := s.store.FindCommunityByID(ctx, communityID)
		if err != nil {
			return
		}
		match, err := s.store.Match(ctx, outcome.MatchID)
		if err != nil {
			return
		}
		newRatings := map[string]float64{}
		for _, d := range outcome.Deltas {
			newRatings[d.PlayerID] = d.NewRating
		}
		event := notify.Match{
			Event: "match.reported", CommunityID: communityID, CommunityName: community.Name,
			MatchID: match.MatchID, Mode: string(match.Mode), PlayedAt: match.PlayedAt,
		}
		for _, p := range match.Participants {
			event.Participants = append(event.Participants, notify.Participant{
				PlayerID: p.PlayerID, Handle: p.Handle, Team: p.Team, Score: p.Score, Finish: p.Finish,
				RatingDelta: p.RatingDelta, NewRating: newRatings[p.PlayerID],
			})
		}
		s.notifier.Dispatch(ctx, integration.Discord, integration.Webhook, event)
	}()
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	actor, err := s.optionalActor(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	mode := model.Mode(r.URL.Query().Get("mode"))
	if mode == "" {
		mode = model.Mode1v1
	}
	if !mode.Valid() {
		writeErr(w, http.StatusBadRequest, "invalid mode %q", mode)
		return
	}
	q := r.URL.Query().Get("q")
	sortBy := r.URL.Query().Get("sort")
	communityID := r.URL.Query().Get("communityId")
	limit := 10
	if r.URL.Query().Get("full") == "1" {
		limit = 100
	}
	rows, err := s.store.Leaderboard(r.Context(), mode, limit, actor, communityID, q, sortBy)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "leaderboard: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// handleLeaderboardModes lists the modes that actually have reported matches so
// the client can hide empty mode tabs.
func (s *Server) handleLeaderboardModes(w http.ResponseWriter, r *http.Request) {
	if _, err := s.optionalActor(r); err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	modes, err := s.store.AvailableModes(r.Context(), r.URL.Query().Get("communityId"))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "leaderboard modes: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, modes)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return
	}
	mode := model.Mode(r.URL.Query().Get("mode"))
	if mode != "" && !mode.Valid() {
		writeErr(w, http.StatusBadRequest, "invalid mode %q", mode)
		return
	}
	limit := 20
	if r.URL.Query().Get("limit") != "" {
		if n, parseErr := strconv.Atoi(r.URL.Query().Get("limit")); parseErr == nil {
			limit = n
		}
	}
	offset := 0
	if r.URL.Query().Get("offset") != "" {
		if n, parseErr := strconv.Atoi(r.URL.Query().Get("offset")); parseErr == nil {
			offset = n
		}
	}
	playerID := r.URL.Query().Get("playerId")
	communityID := r.URL.Query().Get("communityId")
	// Scoping to a community lists every match in it; otherwise fall back to the caller.
	if playerID == "" && communityID == "" {
		player, playerErr := s.store.PrimaryPlayerForUser(r.Context(), actor)
		if playerErr != nil {
			writeErr(w, http.StatusInternalServerError, "resolve player: %v", playerErr)
			return
		}
		playerID = player.ID
	}
	rows, err := s.store.History(r.Context(), playerID, communityID, mode, limit, offset)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "history: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	player, ratings, err := s.store.Profile(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "profile not found: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"player":  player,
		"ratings": ratings,
	})
}

func (s *Server) handleReportMatch(w http.ResponseWriter, r *http.Request) {
	actor, actorErr := s.actorID(r)
	if actorErr != nil {
		writeErr(w, http.StatusUnauthorized, "%v", actorErr)
		return
	}
	var req store.ReportMatchRequest
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request: %v", err)
		return
	}
	outcome, err := s.store.ReportMatchAsMember(r.Context(), actor, req)
	if err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "you must be a community member to report matches")
			return
		}
		if errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrUnknownPlayer) {
			writeErr(w, http.StatusNotFound, "report match: %v", err)
			return
		}
		if errors.Is(err, store.ErrIneligibleParticipant) {
			writeErr(w, http.StatusUnprocessableEntity, "report match: %v", err)
			return
		}
		writeErr(w, http.StatusBadRequest, "report match: %v", err)
		return
	}
	s.announceMatch(req.CommunityID, outcome)
	writeJSON(w, http.StatusCreated, outcome)
}

func (s *Server) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.actorOrUnauthorized(w, r); !ok {
		return
	}
	match, err := s.store.Match(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMatchErr(w, "load match", err)
		return
	}
	writeJSON(w, http.StatusOK, match)
}

func (s *Server) handleUpdateMatch(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.actorOrUnauthorized(w, r)
	if !ok {
		return
	}
	var req struct {
		Participants []store.Participant `json:"participants"`
		Notes        string              `json:"notes"`
	}
	if err := readJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "bad request: %v", err)
		return
	}
	outcome, err := s.store.UpdateMatch(r.Context(), actor, r.PathValue("id"), req.Participants, req.Notes)
	if err != nil {
		writeMatchErr(w, "update match", err)
		return
	}
	writeJSON(w, http.StatusOK, outcome)
}

func (s *Server) handleDeleteMatch(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.actorOrUnauthorized(w, r)
	if !ok {
		return
	}
	communityID, err := s.store.DeleteMatch(r.Context(), actor, r.PathValue("id"))
	if err != nil {
		writeMatchErr(w, "delete match", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "communityId": communityID})
}

func writeMatchErr(w http.ResponseWriter, action string, err error) {
	switch {
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, http.StatusForbidden, "only community owners and admins can change reported matches")
	case errors.Is(err, store.ErrNotFound) || errors.Is(err, store.ErrUnknownPlayer):
		writeErr(w, http.StatusNotFound, "%s: %v", action, err)
	case errors.Is(err, store.ErrIneligibleParticipant):
		writeErr(w, http.StatusUnprocessableEntity, "%s: %v", action, err)
	default:
		writeErr(w, http.StatusBadRequest, "%s: %v", action, err)
	}
}

// --- helpers ---

func (s *Server) actorOrUnauthorized(w http.ResponseWriter, r *http.Request) (string, bool) {
	actor, err := s.actorID(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "%v", err)
		return "", false
	}
	return actor, true
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request, communityID string) (string, bool) {
	actor, ok := s.actorOrUnauthorized(w, r)
	if !ok {
		return "", false
	}
	if err := s.store.CanAdmin(r.Context(), communityID, actor); err != nil {
		if errors.Is(err, store.ErrForbidden) {
			writeErr(w, http.StatusForbidden, "%v", err)
		} else {
			writeErr(w, http.StatusNotFound, "%v", err)
		}
		return "", false
	}
	return actor, true
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	// Responses are per-user and change on every reported match.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, format string, a ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf(format, a...)})
}

// cors allows the configured app origins to call the API cross-origin.
func cors(next http.Handler, demoAuth bool) http.Handler {
	allowedOrigins := configuredOrigins(os.Getenv("OPEN_SKILL_ALLOWED_ORIGINS"))
	allowedHeaders := "Content-Type, Authorization"
	if demoAuth {
		allowedHeaders += ", X-Demo-User"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins["*"] {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func configuredOrigins(value string) map[string]bool {
	if strings.TrimSpace(value) == "" {
		return map[string]bool{"*": true}
	}
	origins := make(map[string]bool)
	for _, origin := range strings.Split(value, ",") {
		// Origin headers never have a trailing slash; tolerate one in config.
		if origin = strings.TrimRight(strings.TrimSpace(origin), "/"); origin != "" {
			origins[origin] = true
		}
	}
	return origins
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
