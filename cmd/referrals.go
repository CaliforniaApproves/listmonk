package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/labstack/echo/v5"
)

const leaderboardMaxLimit = 100

// leaderboardLimit reads ?per_page=, falling back to def and capped at leaderboardMaxLimit.
func leaderboardLimit(c *echo.Context, def int) int {
	limit := def
	if v := c.QueryParam("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > leaderboardMaxLimit {
		limit = leaderboardMaxLimit
	}
	return limit
}

// GetPublicLeaderboard returns a list's leaderboard sanitised for public
// display (rank, "First L." name, count; no e-mails or IDs). It is meant for a
// trusted caller, such as the website's server, to relay.
func (a *App) GetPublicLeaderboard(c *echo.Context) error {
	user := auth.GetUser(c)

	id := getID(c)
	if err := user.HasListPerm(auth.PermTypeGet, id); err != nil {
		return err
	}

	out, err := a.core.GetPublicLeaderboard(id, leaderboardLimit(c, 10))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{out})
}

// SetSubscriberLeaderboardHidden sets whether a subscriber is shown by name on
// public leaderboards.
func (a *App) SetSubscriberLeaderboardHidden(c *echo.Context) error {
	user := auth.GetUser(c)

	id := getID(c)
	if err := a.hasSubPerm(user, auth.PermTypeManage, []int{id}); err != nil {
		return err
	}

	var req struct {
		Hidden bool `json:"hidden"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := a.core.SetLeaderboardHidden([]int{id}, req.Hidden); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{true})
}

// GetListLeaderboard returns a list's referral leaderboard.
func (a *App) GetListLeaderboard(c *echo.Context) error {
	user := auth.GetUser(c)

	id := getID(c)
	if err := user.HasListPerm(auth.PermTypeGet, id); err != nil {
		return err
	}

	limit := leaderboardLimit(c, 25)

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
