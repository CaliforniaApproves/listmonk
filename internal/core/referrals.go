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
// given subscriber's subscriptions to the given lists, but only on subscriptions
// created at or after `since` (so existing subscriptions keep their original
// attribution) and never on a self-referral. Unknown codes and blocklisted
// referrers are silently ignored.
func (c *Core) ApplyReferral(subID int, listUUIDs []string, code string, since time.Time) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || len(listUUIDs) == 0 {
		return nil
	}

	_, err := c.db.Exec(`
		UPDATE subscriber_lists sl SET referred_by = r.id
		FROM subscribers r, lists l
		WHERE r.referral_code = $2
			AND r.id != $1
			AND r.status != 'blocklisted'
			AND sl.subscriber_id = $1
			AND sl.list_id = l.id
			AND l.uuid = ANY($3::UUID[])
			AND sl.referred_by IS NULL
			AND sl.created_at >= $4`,
		subID, code, pq.Array(listUUIDs), since)
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
// Ties share a rank.
func (c *Core) GetListLeaderboard(listID, limit int) ([]LeaderboardEntry, error) {
	out := []LeaderboardEntry{}
	err := c.db.Select(&out, `
		SELECT RANK() OVER (ORDER BY lrc.count DESC)::INT AS rank,
			s.id AS subscriber_id, s.name, s.email, s.referral_code, lrc.count
		FROM list_referral_counts lrc
		JOIN subscribers s ON s.id = lrc.subscriber_id
		WHERE lrc.list_id = $1 AND lrc.count > 0
		ORDER BY lrc.count DESC, s.id ASC
		LIMIT $2`, listID, limit)
	if err != nil {
		c.log.Printf("error fetching leaderboard: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "leaderboard", "error", pqErrMsg(err)))
	}
	return out, nil
}

// ReferralWindow is how recent a subscription must be for ApplyReferralByEmail
// to attribute it, so that pre-existing subscriptions keep their original
// attribution.
const ReferralWindow = 60 * time.Second

// ApplyReferralByEmail records the referrer (by referral code) on the e-mail's
// subscriptions to the given lists, for subscriptions created within the last
// ReferralWindow that have no referrer yet. Self-referrals, unknown codes and
// blocklisted referrers match nothing. Returns the number of subscriptions updated.
func (c *Core) ApplyReferralByEmail(email string, listIDs []int, code string) (int, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || len(listIDs) == 0 {
		return 0, nil
	}

	res, err := c.db.Exec(`
		UPDATE subscriber_lists sl SET referred_by = r.id
		FROM subscribers s, subscribers r
		WHERE LOWER(s.email) = LOWER($1)
			AND sl.subscriber_id = s.id
			AND r.referral_code = $2
			AND r.id != s.id
			AND r.status != 'blocklisted'
			AND sl.list_id = ANY($3::INT[])
			AND sl.referred_by IS NULL
			AND sl.created_at >= NOW() - MAKE_INTERVAL(secs => $4)`,
		email, code, pq.Array(listIDs), ReferralWindow.Seconds())
	if err != nil {
		c.log.Printf("error applying referral: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "referral", "error", pqErrMsg(err)))
	}

	n, _ := res.RowsAffected()
	return int(n), nil
}
