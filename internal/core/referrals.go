package core

import (
	"strings"
	"time"

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
