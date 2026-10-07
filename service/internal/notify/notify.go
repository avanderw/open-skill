// Package notify delivers match reports to a community's Discord and generic
// webhook integrations.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// Kind selects the payload shape and URL rules for an integration.
type Kind string

const (
	Discord Kind = "discord"
	Webhook Kind = "webhook"
)

// Participant is one player's line in a reported match.
type Participant struct {
	PlayerID    string  `json:"playerId"`
	Handle      string  `json:"handle"`
	Team        string  `json:"team,omitempty"`
	Score       float64 `json:"score"`
	Finish      int     `json:"finish"`
	RatingDelta float64 `json:"ratingDelta"`
	NewRating   float64 `json:"newRating"`
}

// Match is the event delivered for a reported match.
type Match struct {
	Event         string        `json:"event"`
	CommunityID   string        `json:"communityId"`
	CommunityName string        `json:"communityName"`
	MatchID       string        `json:"matchId"`
	Mode          string        `json:"mode"`
	PlayedAt      string        `json:"playedAt"`
	Participants  []Participant `json:"participants"`
}

// ValidateURL checks that raw is empty (integration disabled) or a safe target
// for kind.
func ValidateURL(kind Kind, raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("must be a valid URL")
	}
	if u.Scheme != "https" {
		return errors.New("must use https")
	}
	if u.User != nil {
		return errors.New("must not contain credentials")
	}
	host := u.Hostname()
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".internal") {
		return errors.New("must not point at a local address")
	}
	if ip, err := netip.ParseAddr(host); err == nil && blockedIP(ip) {
		return errors.New("must not point at a private address")
	}
	if kind == Discord {
		switch host {
		case "discord.com", "discordapp.com", "ptb.discord.com", "canary.discord.com":
		default:
			return errors.New("must be a Discord webhook URL (discord.com)")
		}
		if !strings.HasPrefix(u.Path, "/api/webhooks/") {
			return errors.New("must be a Discord webhook URL (/api/webhooks/...)")
		}
	}
	return nil
}

// Mask hides the secret part of an integration URL for display.
func Mask(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "••••"
	}
	tail := ""
	if len(raw) > 4 {
		tail = raw[len(raw)-4:]
	}
	if strings.HasPrefix(u.Path, "/api/webhooks/") {
		// Keep the public webhook id, hide the token.
		parts := strings.Split(strings.TrimPrefix(u.Path, "/api/webhooks/"), "/")
		return fmt.Sprintf("%s://%s/api/webhooks/%s/••••%s", u.Scheme, u.Host, parts[0], tail)
	}
	return fmt.Sprintf("%s://%s/••••%s", u.Scheme, u.Host, tail)
}

func blockedIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified()
}

// Notifier sends events over HTTP.
type Notifier struct {
	client *http.Client
}

// New returns a Notifier that refuses to connect to private addresses, even
// after DNS resolution, and does not follow redirects.
func New() *Notifier {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip, err := netip.ParseAddr(host)
			if err != nil || blockedIP(ip) {
				return errors.New("blocked address")
			}
			return nil
		},
	}
	return &Notifier{client: &http.Client{
		Timeout:       10 * time.Second,
		Transport:     &http.Transport{DialContext: dialer.DialContext},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// Dispatch posts m to each configured URL. Failures are logged, never returned,
// so a broken integration cannot affect match reporting.
func (n *Notifier) Dispatch(ctx context.Context, discordURL, webhookURL string, m Match) {
	if discordURL != "" {
		n.send(ctx, Discord, discordURL, DiscordPayload(m))
	}
	if webhookURL != "" {
		n.send(ctx, Webhook, webhookURL, m)
	}
}

func (n *Notifier) send(ctx context.Context, kind Kind, target string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("notify %s: marshal: %v", kind, err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		log.Printf("notify %s: build request: %v", kind, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "open-skill")
	resp, err := n.client.Do(req)
	if err != nil {
		// The URL embeds the secret, so log only the kind and host.
		log.Printf("notify %s: request failed (%s)", kind, redactErr(err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("notify %s: unexpected status %d", kind, resp.StatusCode)
	}
}

func redactErr(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		return ue.Err.Error()
	}
	return "error"
}

// DiscordPayload formats m as a Discord webhook message.
func DiscordPayload(m Match) map[string]any {
	var b strings.Builder
	// Group by team in order of first appearance so a team's players stay together.
	var order []string
	byTeam := map[string][]Participant{}
	for _, p := range m.Participants {
		if _, seen := byTeam[p.Team]; !seen {
			order = append(order, p.Team)
		}
		byTeam[p.Team] = append(byTeam[p.Team], p)
	}
	for _, t := range order {
		if t != "" {
			fmt.Fprintf(&b, "\n**%s**\n", escapeMD(t))
		}
		for _, p := range byTeam[t] {
			fmt.Fprintf(&b, "%s — score %g, %s → %.1f\n", escapeMD(p.Handle), p.Score, signed(p.RatingDelta), p.NewRating)
		}
	}
	desc := strings.TrimSpace(b.String())
	if len(desc) > 4000 {
		desc = desc[:4000] + "…"
	}
	return map[string]any{
		"embeds": []map[string]any{{
			"title":       fmt.Sprintf("Match reported in %s", m.CommunityName),
			"description": desc,
			"footer":      map[string]string{"text": "Mode: " + m.Mode},
		}},
		// Handles are user-supplied; never let them ping anyone.
		"allowed_mentions": map[string]any{"parse": []string{}},
	}
}

func signed(v float64) string {
	if v >= 0 {
		return fmt.Sprintf("+%.1f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

var mdEscaper = strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "~", "\\~", "`", "\\`", "|", "\\|", ">", "\\>", "@", "@\u200b")

func escapeMD(s string) string { return mdEscaper.Replace(s) }
