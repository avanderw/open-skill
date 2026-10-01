package store

import (
	"context"
	"os"
	"testing"
)

// Runs only when OPEN_SKILL_DB_URL is set, e.g. via ./with-env.ps1 against a Turso database.
func TestRemoteOpenEnforcesForeignKeys(t *testing.T) {
	url := os.Getenv("OPEN_SKILL_DB_URL")
	if url == "" {
		t.Skip("OPEN_SKILL_DB_URL not set")
	}
	st, err := OpenRemote(url, os.Getenv("OPEN_SKILL_DB_TOKEN"))
	if err != nil {
		t.Fatalf("open remote: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	if err := st.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
	_, err = st.DB().ExecContext(ctx,
		`INSERT INTO players (id, user_id, handle) VALUES ('fk_probe_player', 'fk_probe_missing_user', 'fk_probe_handle')`)
	if err == nil {
		_, _ = st.DB().ExecContext(ctx, `DELETE FROM players WHERE id = 'fk_probe_player'`)
		t.Fatal("expected foreign key violation inserting a player for a missing user")
	}
}
