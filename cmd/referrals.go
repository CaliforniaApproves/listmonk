package main

import (
	"net/http"
	"strconv"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/labstack/echo/v4"
)

const leaderboardMaxLimit = 100

// GetListLeaderboard returns a list's referral leaderboard.
func (a *App) GetListLeaderboard(c echo.Context) error {
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

// ApplyReferral attributes recently created subscriptions of a subscriber to a
// referrer. It is for trusted server-side callers (e.g. the website's form
// handler) that add subscribers through the admin API.
func (a *App) ApplyReferral(c echo.Context) error {
	user := auth.GetUser(c)

	var req struct {
		Email      string `json:"email"`
		Lists      []int  `json:"list_ids"`
		ReferredBy string `json:"referred_by"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Only lists the caller may manage.
	listIDs := user.FilterListsByPerm(auth.PermTypeManage, req.Lists)
	if len(req.Lists) > 0 && len(listIDs) == 0 {
		return echo.NewHTTPError(http.StatusForbidden, a.i18n.Ts("globals.messages.permissionDenied", "name", "lists"))
	}

	n, err := a.core.ApplyReferralByEmail(req.Email, listIDs, req.ReferredBy)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{struct {
		Updated int `json:"updated"`
	}{n}})
}
