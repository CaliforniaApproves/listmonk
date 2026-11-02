// Package camigrations is the CaliforniaApproves fork's own database
// migration track. It is deliberately independent of upstream's semver based
// migration list (cmd/upgrade.go `migList`) so that fork schema changes never
// collide with upstream versions. Applied step IDs are recorded in the
// `ca_migrations` key of the `settings` table.
//
// Steps run (in order) after upstream's schema install / migrations, and must
// be idempotent.
package camigrations

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// step is a single fork migration identified by a unique, ordered ID.
type step struct {
	id  string
	sql string
}

// steps is the ordered list of fork migrations. Append only; never edit or
// reorder a step that has shipped.
var steps = []step{
	{"001_referral_code", sql001ReferralCode},
	{"002_referred_by", sql002ReferredBy},
	{"003_leaderboard", sql003Leaderboard},
}

const settingsKey = "ca_migrations"

// applied returns the set of already applied step IDs.
func applied(db *sqlx.DB) (map[string]bool, error) {
	var ids []string
	err := db.Select(&ids, `
		SELECT JSONB_ARRAY_ELEMENTS_TEXT(value) FROM settings WHERE key = $1`, settingsKey)
	if err != nil {
		return nil, err
	}

	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// Pending returns the IDs of fork migrations that have not been applied.
func Pending(db *sqlx.DB) ([]string, error) {
	done, err := applied(db)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, s := range steps {
		if !done[s.id] {
			out = append(out, s.id)
		}
	}
	return out, nil
}

// Run applies all pending fork migrations in order. Each step and its
// bookkeeping record are committed in one transaction.
func Run(db *sqlx.DB, lo *log.Logger) error {
	done, err := applied(db)
	if err != nil {
		return err
	}

	for _, s := range steps {
		if done[s.id] {
			continue
		}

		lo.Printf("running fork migration %s", s.id)
		if err := runStep(db, s); err != nil {
			return err
		}
	}
	return nil
}

func runStep(db *sqlx.DB, s step) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(s.sql); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO settings (key, value) VALUES ($1, JSONB_BUILD_ARRAY($2::TEXT))
		ON CONFLICT (key) DO UPDATE SET value = settings.value || EXCLUDED.value`,
		settingsKey, s.id); err != nil {
		return err
	}

	return tx.Commit()
}
