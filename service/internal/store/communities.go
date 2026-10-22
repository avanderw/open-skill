package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"open-skill/service/internal/model"
	"open-skill/service/internal/rating"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ErrCodeTaken is returned when a community's invite code is already in use so
// the caller can retry with a freshly generated one.
var ErrCodeTaken = errors.New("invite code already in use")

// ListCommunities returns all communities with an owner display name.
func (s *Store) ListCommunities(ctx context.Context) ([]model.Community, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.name, c.mode, c.draws, c.code, c.owner_id,
		       (SELECT COUNT(*) FROM community_members m WHERE m.community_id = c.id) AS members
		FROM communities c
		ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list communities: %w", err)
	}
	defer rows.Close()

	out := []model.Community{}
	for rows.Next() {
		var c model.Community
		if err := rows.Scan(&c.ID, &c.Name, &c.Mode, &c.Draws, &c.Code, &c.OwnerID, &c.Members); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListCommunitiesForUser(ctx context.Context, userID string) ([]model.Community, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.name, c.mode, c.draws, c.code, c.owner_id,
		       (SELECT COUNT(*) FROM community_members m WHERE m.community_id = c.id),
		       (SELECT COUNT(DISTINCT mp.match_id)
		          FROM match_participants mp
		          JOIN matches mt ON mt.id = mp.match_id AND mt.community_id = c.id
		          JOIN players p ON p.id = mp.player_id AND p.user_id = ?)
		FROM communities c
		JOIN community_members me ON me.community_id = c.id AND me.user_id = ?
		ORDER BY c.created_at DESC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Community{}
	for rows.Next() {
		var c model.Community
		if err := rows.Scan(&c.ID, &c.Name, &c.Mode, &c.Draws, &c.Code, &c.OwnerID, &c.Members, &c.YourGames); err != nil {
			return nil, err
		}
		c.Membership = "member"
		if c.OwnerID == userID {
			c.Membership = "owner"
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) BrowseCommunities(ctx context.Context, userID string) ([]model.Community, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT c.id, c.name, c.mode, c.draws, c.code, c.owner_id, (SELECT COUNT(*) FROM community_members m WHERE m.community_id = c.id), CASE WHEN c.owner_id = ? THEN 'owner' WHEN EXISTS (SELECT 1 FROM community_members m2 WHERE m2.community_id = c.id AND m2.user_id = ?) THEN 'member' WHEN EXISTS (SELECT 1 FROM community_join_requests jr WHERE jr.community_id = c.id AND jr.user_id = ? AND jr.status = 'pending') THEN 'pending' ELSE 'non_member' END FROM communities c ORDER BY c.created_at DESC`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Community{}
	for rows.Next() {
		var c model.Community
		if err := rows.Scan(&c.ID, &c.Name, &c.Mode, &c.Draws, &c.Code, &c.OwnerID, &c.Members, &c.Membership); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) Membership(ctx context.Context, communityID, userID string) (string, error) {
	status, _, err := s.MembershipAccess(ctx, communityID, userID)
	return status, err
}

// MembershipAccess returns the user's membership status and whether they may
// manage the community, in one round trip.
func (s *Store) MembershipAccess(ctx context.Context, communityID, userID string) (status string, canManage bool, err error) {
	var owner string
	var role, request sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT c.owner_id, cm.role,
		       (SELECT jr.status FROM community_join_requests jr WHERE jr.community_id = c.id AND jr.user_id = ?)
		FROM communities c
		LEFT JOIN community_members cm ON cm.community_id = c.id AND cm.user_id = ?
		WHERE c.id = ?`, userID, userID, communityID).Scan(&owner, &role, &request); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, ErrNotFound
		}
		return "", false, err
	}
	switch {
	case owner == userID:
		return "owner", true, nil
	case role.Valid:
		return "member", role.String == "owner" || role.String == "admin", nil
	case request.Valid:
		return request.String, false, nil
	default:
		return "non_member", false, nil
	}
}

// CreateCommunity inserts a community, its join code, and the owner membership
// in one transaction.
func (s *Store) CreateCommunity(ctx context.Context, c model.Community) (*model.Community, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO communities (id, name, mode, draws, code, owner_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.Name, c.Mode, c.Draws, c.Code, c.OwnerID,
	); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: communities.code") {
			return nil, ErrCodeTaken
		}
		return nil, fmt.Errorf("insert community: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO community_members (community_id, user_id, role)
		VALUES (?, ?, 'owner')`, c.ID, c.OwnerID); err != nil {
		return nil, fmt.Errorf("insert owner: %w", err)
	}
	// Give the owner's players a default (unrated) leaderboard entry so the
	// creator shows up immediately instead of the leaderboard being empty.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO ratings (community_id, player_id, mode, mu, sigma)
		SELECT ?, p.id, m.mode, ?, ?
		FROM players p, (SELECT '1v1' AS mode UNION ALL SELECT 'team' UNION ALL SELECT 'ffa') m
		WHERE p.user_id = ?
		ON CONFLICT(community_id, player_id, mode) DO NOTHING`,
		c.ID, rating.DefaultOpenSkillMu, rating.DefaultOpenSkillSigma, c.OwnerID); err != nil {
		return nil, fmt.Errorf("insert owner default rating: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	c.Members = 1
	c.Membership = "owner"
	return &c, nil
}

// FindCommunityByID returns a community by its stable id.
func (s *Store) FindCommunityByID(ctx context.Context, id string) (*model.Community, error) {
	var c model.Community
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, mode, draws, code, owner_id,
		       (SELECT COUNT(*) FROM community_members m WHERE m.community_id = c.id)
		FROM communities c WHERE id = ?`, id).
		Scan(&c.ID, &c.Name, &c.Mode, &c.Draws, &c.Code, &c.OwnerID, &c.Members)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateCommunity(ctx context.Context, id, name, mode string, draws bool) error {
	result, err := s.db.ExecContext(ctx, `UPDATE communities SET name = ?, mode = ?, draws = ? WHERE id = ?`, name, mode, draws, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteCommunity permanently removes a community and everything scoped to it. Owner only.
func (s *Store) DeleteCommunity(ctx context.Context, actorID, id string) error {
	role, _, err := s.communityRoleAndOwner(ctx, id, actorID)
	if err != nil {
		return err
	}
	if role != "owner" {
		return ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmts := []string{
		`DELETE FROM match_participants WHERE match_id IN (SELECT id FROM matches WHERE community_id = ?)`,
		`DELETE FROM matches WHERE community_id = ?`,
		`DELETE FROM ratings WHERE community_id = ?`,
		`DELETE FROM community_integrations WHERE community_id = ?`,
		`DELETE FROM community_join_requests WHERE community_id = ?`,
		`DELETE FROM community_members WHERE community_id = ?`,
		`DELETE FROM player_claim_requests WHERE player_id IN (SELECT id FROM players WHERE community_id = ?)`,
		`UPDATE players SET community_id = NULL WHERE community_id = ?`,
		`DELETE FROM communities WHERE id = ?`,
	}
	for _, q := range stmts {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return fmt.Errorf("delete community: %w", err)
		}
	}
	return tx.Commit()
}

// FindCommunityByCode looks up a community by its invite code.
func (s *Store) FindCommunityByCode(ctx context.Context, code string) (*model.Community, error) {
	var c model.Community
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, mode, draws, code, owner_id,
		       (SELECT COUNT(*) FROM community_members m WHERE m.community_id = c.id)
		FROM communities c WHERE code = ?`, code).
		Scan(&c.ID, &c.Name, &c.Mode, &c.Draws, &c.Code, &c.OwnerID, &c.Members)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// JoinCommunity adds a user immediately for open communities, or records a pending request.
func (s *Store) JoinCommunity(ctx context.Context, communityID, userID string) error {
	var mode string
	if err := s.db.QueryRowContext(ctx, `SELECT mode FROM communities WHERE id = ?`, communityID).Scan(&mode); err != nil {
		return err
	}
	if mode == "approval" {
		_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO community_join_requests (community_id, user_id, status) VALUES (?, ?, 'pending')`, communityID, userID)
		return err
	}
	return s.AcceptCommunityMember(ctx, communityID, userID)
}

func (s *Store) AcceptCommunityMember(ctx context.Context, communityID, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := acceptCommunityMemberTx(ctx, tx, communityID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func acceptCommunityMemberTx(ctx context.Context, tx *sql.Tx, communityID, userID string) error {
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO community_members (community_id, user_id, role) VALUES (?, ?, 'member')`, communityID, userID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO ratings (community_id, player_id, mode, mu, sigma)
		SELECT ?, p.id, m.mode, ?, ?
		FROM players p, (SELECT '1v1' AS mode UNION ALL SELECT 'team' UNION ALL SELECT 'ffa') m
		WHERE p.user_id = ?
		ON CONFLICT(community_id, player_id, mode) DO NOTHING`,
		communityID, rating.DefaultOpenSkillMu, rating.DefaultOpenSkillSigma, userID)
	return err
}

var ErrForbidden = errors.New("forbidden")

func (s *Store) CommunityRole(ctx context.Context, communityID, userID string) (string, error) {
	role, _, err := s.communityRoleAndOwner(ctx, communityID, userID)
	return role, err
}

// communityRoleAndOwner resolves the user's role and the community owner in one
// round trip. Non-members get ErrNotFound.
func (s *Store) communityRoleAndOwner(ctx context.Context, communityID, userID string) (role, owner string, err error) {
	var memberRole sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT c.owner_id, cm.role
		FROM communities c
		LEFT JOIN community_members cm ON cm.community_id = c.id AND cm.user_id = ?
		WHERE c.id = ?`, userID, communityID).Scan(&owner, &memberRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrNotFound
		}
		return "", "", err
	}
	if owner == userID {
		return "owner", owner, nil
	}
	if !memberRole.Valid {
		return "", owner, ErrNotFound
	}
	return memberRole.String, owner, nil
}

func (s *Store) CanAdmin(ctx context.Context, communityID, userID string) error {
	role, err := s.CommunityRole(ctx, communityID, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return ErrForbidden
	}
	return nil
}

func (s *Store) SetMemberRole(ctx context.Context, communityID, actorID, targetID, role string) error {
	if role != "admin" && role != "member" {
		return fmt.Errorf("invalid role")
	}
	actor, owner, err := s.communityRoleAndOwner(ctx, communityID, actorID)
	if err != nil {
		return err
	}
	if actor != "owner" {
		return ErrForbidden
	}
	if targetID == owner {
		return fmt.Errorf("owner cannot be changed")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE community_members SET role = ? WHERE community_id = ? AND user_id = ?`, role, communityID, targetID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) RemoveMember(ctx context.Context, communityID, actorID, targetID string) error {
	var owner string
	var actorRole, targetRole sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT c.owner_id, a.role, t.role
		FROM communities c
		LEFT JOIN community_members a ON a.community_id = c.id AND a.user_id = ?
		LEFT JOIN community_members t ON t.community_id = c.id AND t.user_id = ?
		WHERE c.id = ?`, actorID, targetID, communityID).Scan(&owner, &actorRole, &targetRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	actor := actorRole.String
	if owner == actorID {
		actor = "owner"
	} else if !actorRole.Valid {
		return ErrNotFound
	}
	if actor != "owner" && actor != "admin" {
		return ErrForbidden
	}
	if targetID == owner {
		return fmt.Errorf("owner cannot be removed")
	}
	if actor == "admin" {
		if !targetRole.Valid {
			return ErrNotFound
		}
		if targetRole.String == "admin" {
			return ErrForbidden
		}
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM community_members WHERE community_id = ? AND user_id = ?`, communityID, targetID)
	return err
}

func (s *Store) ApproveJoinRequest(ctx context.Context, communityID, userID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE community_join_requests SET status = 'approved' WHERE community_id = ? AND user_id = ? AND status = 'pending'`, communityID, userID)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return ErrNotFound
	}
	if err := acceptCommunityMemberTx(ctx, tx, communityID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

type CommunityJoinRequest struct {
	CommunityID string `json:"communityId"`
	UserID      string `json:"userId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

func (s *Store) JoinRequests(ctx context.Context, communityID string) ([]CommunityJoinRequest, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT community_id, user_id, status, created_at FROM community_join_requests WHERE community_id = ? ORDER BY created_at DESC`, communityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CommunityJoinRequest{}
	for rows.Next() {
		var r CommunityJoinRequest
		if err := rows.Scan(&r.CommunityID, &r.UserID, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type CommunityIntegration struct {
	CommunityID string `json:"communityId"`
	Discord     string `json:"discord"`
	Webhook     string `json:"webhook"`
}

func (s *Store) Integration(ctx context.Context, communityID string) (*CommunityIntegration, error) {
	var i CommunityIntegration
	err := s.db.QueryRowContext(ctx, `SELECT community_id, discord, webhook FROM community_integrations WHERE community_id = ?`, communityID).Scan(&i.CommunityID, &i.Discord, &i.Webhook)
	if errors.Is(err, sql.ErrNoRows) {
		return &CommunityIntegration{CommunityID: communityID}, nil
	}
	return &i, err
}
func (s *Store) UpdateIntegration(ctx context.Context, i CommunityIntegration) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO community_integrations (community_id, discord, webhook) VALUES (?, ?, ?) ON CONFLICT(community_id) DO UPDATE SET discord=excluded.discord, webhook=excluded.webhook, updated_at=datetime('now')`, i.CommunityID, i.Discord, i.Webhook)
	return err
}
