package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// Identity is a caller whose bearer token has been verified.
type Identity struct {
	UID  string
	Name string
}

// TokenVerifier verifies bearer tokens (Firebase ID tokens in production).
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (Identity, error)
}

// AuthConfig selects how callers are identified.
type AuthConfig struct {
	// Verifier enables Authorization: Bearer tokens; nil disables them.
	Verifier TokenVerifier
	// Demo trusts the X-Demo-User header. Development and tests only.
	Demo bool
}

var errSignInRequired = errors.New("sign in required")

type firebaseVerifier struct{ client *fbauth.Client }

// NewFirebaseVerifier verifies Firebase ID tokens for projectID. Verification
// uses Google's public signing keys, so no service account credentials are needed.
func NewFirebaseVerifier(ctx context.Context, projectID string) (TokenVerifier, error) {
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("init firebase: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase auth: %w", err)
	}
	return firebaseVerifier{client: client}, nil
}

func (v firebaseVerifier) VerifyToken(ctx context.Context, token string) (Identity, error) {
	t, err := v.client.VerifyIDToken(ctx, token)
	if err != nil {
		return Identity{}, err
	}
	name, _ := t.Claims["name"].(string)
	return Identity{UID: t.UID, Name: name}, nil
}

// knownUsers remembers users already materialized so steady-state requests
// skip the database.
type knownUsers struct{ m sync.Map }

func (k *knownUsers) has(id string) bool { _, ok := k.m.Load(id); return ok }
func (k *knownUsers) add(id string)      { k.m.Store(id, struct{}{}) }

// identify resolves the caller. It returns "" for anonymous callers and an
// error for invalid credentials.
func (s *Server) identify(r *http.Request) (string, error) {
	if header := r.Header.Get("Authorization"); header != "" {
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || s.auth.Verifier == nil {
			return "", errors.New("unsupported authorization")
		}
		identity, err := s.auth.Verifier.VerifyToken(r.Context(), strings.TrimSpace(token))
		if err != nil || identity.UID == "" {
			return "", errors.New("invalid or expired token")
		}
		userID := "fb_" + identity.UID
		if !s.users.has(userID) {
			if err := s.store.EnsureUser(r.Context(), userID, "p_"+identity.UID, playerHandle(identity.Name)); err != nil {
				return "", err
			}
			s.users.add(userID)
		}
		return userID, nil
	}
	if !s.auth.Demo {
		return "", nil
	}
	id := strings.TrimSpace(strings.ToLower(r.Header.Get("X-Demo-User")))
	if id == "" {
		id = s.stubUser
	} else if id != "u_demo" && id != "u_member" && id != "u_guest" && !strings.Contains(id, "@") {
		return "", fmt.Errorf("unknown demo user")
	}
	if !s.users.has(id) {
		// The database starts empty, so the demo identity is created on first use.
		if err := s.store.EnsureUser(r.Context(), id, id, id); err != nil {
			return "", err
		}
		s.users.add(id)
	}
	return id, nil
}

// actorID requires a signed-in caller.
func (s *Server) actorID(r *http.Request) (string, error) {
	id, err := s.identify(r)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", errSignInRequired
	}
	return id, nil
}

// optionalActor allows anonymous callers (""), but still rejects bad credentials.
func (s *Server) optionalActor(r *http.Request) (string, error) {
	return s.identify(r)
}

// maxHandleRunes bounds handles to keep leaderboards readable while fitting full names.
const maxHandleRunes = 64

// playerHandle derives a public handle from the display name. Emails are never
// used so they are not exposed on leaderboards.
func playerHandle(name string) string {
	handle := strings.Join(strings.Fields(name), " ")
	if handle == "" {
		return "player"
	}
	if runes := []rune(handle); len(runes) > maxHandleRunes {
		handle = string(runes[:maxHandleRunes])
	}
	return handle
}
