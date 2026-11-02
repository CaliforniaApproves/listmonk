package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/labstack/echo/v5"
)

const leaderboardMaxLimit = 100

// GetListLeaderboard returns a list's referral leaderboard.
func (a *App) GetListLeaderboard(c *echo.Context) error {
	user := auth.GetUser(c)

	id := getID(c)
	if err := user.HasListPerm(auth.PermTypeGet, id); err != nil {
		return err
	}

	limit := 25
	if v := c.QueryParam("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > leaderboardMaxLimit {
		limit = leaderboardMaxLimit
	}

	out, err := a.core.GetListLeaderboard(id, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{out})
}

// applyReferral is a best-effort wrapper around core.ApplyReferral: a failure to
// record the referrer is logged but never fails the subscription itself. `since`
// and `clockErr` come from a core.DBClock() call made before the subscription.
func (a *App) applyReferral(subIDs, listIDs []int, listUUIDs []string, code string, since time.Time, clockErr error) {
	if code == "" {
		return
	}
	if clockErr != nil {
		a.log.Printf("error recording referral: %v", clockErr)
		return
	}
	if err := a.core.ApplyReferral(subIDs, listIDs, listUUIDs, code, since); err != nil {
		a.log.Printf("error recording referral: %v", err)
	}
}
