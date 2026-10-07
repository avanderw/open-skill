package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	cases := []struct {
		kind Kind
		url  string
		ok   bool
	}{
		{Discord, "", true},
		{Discord, "https://discord.com/api/webhooks/1/abc", true},
		{Discord, "https://example.com/api/webhooks/1/abc", false},
		{Discord, "https://discord.com/other", false},
		{Webhook, "https://example.com/hook", true},
		{Webhook, "http://example.com/hook", false},
		{Webhook, "https://localhost/hook", false},
		{Webhook, "https://127.0.0.1/hook", false},
		{Webhook, "https://10.0.0.5/hook", false},
		{Webhook, "https://169.254.169.254/latest", false},
		{Webhook, "https://[::1]/hook", false},
		{Webhook, "https://user:pw@example.com/hook", false},
		{Webhook, "not a url", false},
	}
	for _, c := range cases {
		if err := ValidateURL(c.kind, c.url); (err == nil) != c.ok {
			t.Errorf("ValidateURL(%s, %q) = %v, want ok=%v", c.kind, c.url, err, c.ok)
		}
	}
}

func TestMaskHidesSecret(t *testing.T) {
	got := Mask("https://discord.com/api/webhooks/123/supersecrettoken")
	if strings.Contains(got, "supersecret") || !strings.Contains(got, "/123/") {
		t.Errorf("Mask = %q", got)
	}
	if got := Mask("https://example.com/hooks/secretpath"); strings.Contains(got, "secretpath") {
		t.Errorf("Mask = %q", got)
	}
	if Mask("") != "" {
		t.Error("empty must stay empty")
	}
}

func sampleMatch() Match {
	return Match{
		Event: "match.reported", CommunityName: "Squash", MatchID: "m1", Mode: "team",
		Participants: []Participant{
			{Handle: "ann", Team: "Red", Score: 11, RatingDelta: 12.34, NewRating: 1512},
			{Handle: "@everyone", Team: "Blue", Score: 7, RatingDelta: -12.34, NewRating: 1488},
			{Handle: "bob", Team: "Red", Score: 11, RatingDelta: 10, NewRating: 1510},
		},
	}
}

func TestDiscordPayloadGroupsTeamsAndBlocksMentions(t *testing.T) {
	p := DiscordPayload(sampleMatch())
	body, _ := json.Marshal(p)
	s := string(body)
	if !strings.Contains(s, `"parse":[]`) {
		t.Errorf("mentions not suppressed: %s", s)
	}
	embeds := p["embeds"].([]map[string]any)
	desc := embeds[0]["description"].(string)
	if strings.Index(desc, "bob") > strings.Index(desc, "Blue") {
		t.Errorf("Red team not grouped: %q", desc)
	}
	if strings.Contains(desc, "@everyone") {
		t.Errorf("raw mention present: %q", desc)
	}
}

func TestDispatchPostsBothFormats(t *testing.T) {
	got := map[string]map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var v map[string]any
		_ = json.Unmarshal(b, &v)
		got[r.URL.Path] = v
	}))
	defer srv.Close()
	n := &Notifier{client: srv.Client()}
	n.Dispatch(context.Background(), srv.URL+"/discord", srv.URL+"/hook", sampleMatch())
	if _, ok := got["/discord"]["embeds"]; !ok {
		t.Errorf("discord payload = %v", got["/discord"])
	}
	if got["/hook"]["event"] != "match.reported" {
		t.Errorf("webhook payload = %v", got["/hook"])
	}
}

func TestNotifierRefusesLoopback(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit = true }))
	defer srv.Close()
	New().Dispatch(context.Background(), "", srv.URL, sampleMatch())
	if hit {
		t.Fatal("request reached a loopback address")
	}
}
