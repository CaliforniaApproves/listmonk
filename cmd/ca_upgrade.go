package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/knadh/listmonk/internal/camigrations"
)

// runForkMigrations applies the CaliforniaApproves fork's own migrations
// (internal/camigrations), which are tracked separately from upstream's.
func runForkMigrations(db *sqlx.DB) {
	if err := camigrations.Run(db, lo); err != nil {
		lo.Fatalf("error running fork migrations: %v", err)
	}
}

// checkForkUpgrade aborts startup if fork migrations are pending.
func checkForkUpgrade(db *sqlx.DB) {
	pending, err := camigrations.Pending(db)
	if err != nil {
		lo.Fatalf("error checking fork migrations: %v", err)
	}
	if len(pending) > 0 {
		lo.Fatalf("there are %d pending fork database upgrade(s): %v. Backup the database and run listmonk --upgrade",
			len(pending), pending)
	}
}
