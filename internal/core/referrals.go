package core

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// DBClock returns the database's current wall-clock time. It is used to
// identify subscriptions created after this point, independent of app/DB clock skew.
func (c *Core) DBClock() (time.Time, error) {
	var t time.Time
	err := c.db.Get(&t, `SELECT clock_timestamp()`)
	return t, err
}

// ApplyReferral records the referrer (identified by their referral code) on the
// given subscribers' subscriptions to the given lists (by ID and/or UUID), but
// only on subscriptions created at or after `since` (so existing subscriptions
// keep their original attribution) and never on a self-referral. Unknown codes
// and blocklisted referrers are silently ignored.
func (c *Core) ApplyReferral(subIDs, listIDs []int, listUUIDs []string, code string, since time.Time) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || len(subIDs) == 0 || len(listIDs)+len(listUUIDs) == 0 {
		return nil
	}
	if listIDs == nil {
		listIDs = []int{}
	}
	if listUUIDs == nil {
		listUUIDs = []string{}
	}

	_, err := c.db.Exec(`
		UPDATE subscriber_lists sl SET referred_by = r.id
		FROM subscribers r, lists l
		WHERE r.referral_code = $2
			AND r.id != sl.subscriber_id
			AND r.status != 'blocklisted'
			AND sl.subscriber_id = ANY($1::INT[])
			AND sl.list_id = l.id
			AND (l.id = ANY($3::INT[]) OR l.uuid = ANY($4::UUID[]))
			AND sl.referred_by IS NULL
			AND sl.created_at >= $5`,
		pq.Array(subIDs), code, pq.Array(listIDs), pq.Array(listUUIDs), since)
	return err
}

// LeaderboardEntry is one referrer's standing on a list's leaderboard.
type LeaderboardEntry struct {
	Rank         int    `db:"rank" json:"rank"`
	SubscriberID int    `db:"subscriber_id" json:"subscriber_id"`
	Name         string `db:"name" json:"name"`
	Email        string `db:"email" json:"email"`
	ReferralCode string `db:"referral_code" json:"referral_code"`
	Count        int    `db:"count" json:"count"`
}

// GetListLeaderboard returns the top referrers of a list by confirmed referrals.
// Ties share a rank. The top N rows come straight off the
// idx_list_referral_counts_top index; ranking only the N rows is exact, because
// every row ranked above one of them is also in the top N.
func (c *Core) GetListLeaderboard(listID, limit int) ([]LeaderboardEntry, error) {
	out := []LeaderboardEntry{}
	err := c.db.Select(&out, `
		WITH top AS (
			SELECT subscriber_id, count FROM list_referral_counts
			WHERE list_id = $1 AND count > 0
			ORDER BY count DESC, subscriber_id
			LIMIT $2
		)
		SELECT RANK() OVER (ORDER BY t.count DESC)::INT AS rank,
			s.id AS subscriber_id, s.name, s.email, s.referral_code, t.count
		FROM top t JOIN subscribers s ON s.id = t.subscriber_id
		ORDER BY t.count DESC, t.subscriber_id`, listID, limit)
	if err != nil {
		c.log.Printf("error fetching leaderboard: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "leaderboard", "error", pqErrMsg(err)))
	}
	return out, nil
}

// PublicLeaderboardEntry is a leaderboard row that is safe to show publicly:
// no e-mail, ID or referral code.
type PublicLeaderboardEntry struct {
	Rank  int    `db:"rank" json:"rank"`
	Name  string `db:"name" json:"name"`
	Count int    `db:"count" json:"count"`
}

// GetPublicLeaderboard returns the top referrers of a list for public display.
// Names are "First L." (first word of the name, plus the initial of its last
// word). A subscriber who opted out, has no name, or whose name is just the
// local part of their e-mail address (the fallback used when none is given) is
// shown as "Anonymous" and keeps their rank.
func (c *Core) GetPublicLeaderboard(listID, limit int) ([]PublicLeaderboardEntry, error) {
	out := []PublicLeaderboardEntry{}
	err := c.db.Select(&out, `
		WITH top AS (
			SELECT subscriber_id, count FROM list_referral_counts
			WHERE list_id = $1 AND count > 0
			ORDER BY count DESC, subscriber_id
			LIMIT $2
		)
		SELECT RANK() OVER (ORDER BY t.count DESC)::INT AS rank, t.count,
			CASE
				WHEN s.hide_from_leaderboard
					OR BTRIM(s.name) = ''
					OR LOWER(BTRIM(s.name)) = LOWER(SPLIT_PART(s.email, '@', 1))
				THEN 'Anonymous'
				ELSE LEFT(SPLIT_PART(BTRIM(s.name), ' ', 1), 20)
					|| CASE WHEN BTRIM(s.name) ~ '\s'
						THEN ' ' || UPPER(LEFT(SUBSTRING(BTRIM(s.name) FROM '(\S+)$'), 1)) || '.'
						ELSE '' END
			END AS name
		FROM top t JOIN subscribers s ON s.id = t.subscriber_id
		ORDER BY t.count DESC, t.subscriber_id`, listID, limit)
	if err != nil {
		c.log.Printf("error fetching public leaderboard: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "leaderboard", "error", pqErrMsg(err)))
	}
	return out, nil
}

// SetLeaderboardHidden sets whether the subscribers are shown by name on public
// leaderboards.
func (c *Core) SetLeaderboardHidden(subIDs []int, hidden bool) error {
	if len(subIDs) == 0 {
		return nil
	}
	_, err := c.db.Exec(`UPDATE subscribers SET hide_from_leaderboard = $2 WHERE id = ANY($1::INT[])`,
		pq.Array(subIDs), hidden)
	if err != nil {
		c.log.Printf("error updating leaderboard visibility: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.subscriber}", "error", pqErrMsg(err)))
	}
	return nil
}
