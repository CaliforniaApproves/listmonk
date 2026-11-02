package camigrations

// sql004LeaderboardPrivacy adds subscribers.hide_from_leaderboard (a subscriber's
// opt-out from being named on a public leaderboard; they are shown as
// "Anonymous" and keep their rank) and replaces the leaderboard rank index with
// one that serves "top N of a list" without a sort: it is ordered the way the
// query orders (count desc, then subscriber id) and skips zero counts.
const sql004LeaderboardPrivacy = `
ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS hide_from_leaderboard BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_list_referral_counts_top
    ON list_referral_counts(list_id, count DESC, subscriber_id) WHERE count > 0;
DROP INDEX IF EXISTS idx_list_referral_counts_rank;
`
