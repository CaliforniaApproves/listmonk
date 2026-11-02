package camigrations

// sql003Leaderboard adds list_referral_counts: per list, how many confirmed
// subscriptions each subscriber has referred. It is maintained by triggers on
// subscriber_lists so it stays correct across every code path (API, import,
// bulk operations, opt-in confirmation, unsubscribes, deletes).
//
// A subscription counts for its referrer (referred_by) in its list only while
// its status is 'confirmed'.
const sql003Leaderboard = `
CREATE TABLE IF NOT EXISTS list_referral_counts (
    list_id        INTEGER NOT NULL REFERENCES lists(id) ON DELETE CASCADE ON UPDATE CASCADE,
    subscriber_id  INTEGER NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE ON UPDATE CASCADE,
    count          INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (list_id, subscriber_id)
);
CREATE INDEX IF NOT EXISTS idx_list_referral_counts_rank ON list_referral_counts(list_id, count DESC);

CREATE OR REPLACE FUNCTION ca_referral_counts_trg() RETURNS TRIGGER AS $$
DECLARE
    old_counts BOOLEAN := FALSE;
    new_counts BOOLEAN := FALSE;
BEGIN
    IF TG_OP IN ('UPDATE', 'DELETE') THEN
        old_counts := OLD.referred_by IS NOT NULL AND OLD.status = 'confirmed';
    END IF;
    IF TG_OP IN ('INSERT', 'UPDATE') THEN
        new_counts := NEW.referred_by IS NOT NULL AND NEW.status = 'confirmed';
    END IF;

    -- Nothing changes if the subscription counts for the same (list, referrer) before and after.
    IF TG_OP = 'UPDATE' AND old_counts AND new_counts
       AND OLD.list_id = NEW.list_id AND OLD.referred_by = NEW.referred_by THEN
        RETURN NEW;
    END IF;

    IF old_counts THEN
        UPDATE list_referral_counts SET count = GREATEST(count - 1, 0)
            WHERE list_id = OLD.list_id AND subscriber_id = OLD.referred_by;
    END IF;

    IF new_counts THEN
        INSERT INTO list_referral_counts (list_id, subscriber_id, count)
            VALUES (NEW.list_id, NEW.referred_by, 1)
            ON CONFLICT (list_id, subscriber_id)
            DO UPDATE SET count = list_referral_counts.count + 1;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS ca_referral_counts_ins ON subscriber_lists;
CREATE TRIGGER ca_referral_counts_ins AFTER INSERT ON subscriber_lists
    FOR EACH ROW WHEN (NEW.referred_by IS NOT NULL AND NEW.status = 'confirmed')
    EXECUTE FUNCTION ca_referral_counts_trg();

DROP TRIGGER IF EXISTS ca_referral_counts_upd ON subscriber_lists;
CREATE TRIGGER ca_referral_counts_upd AFTER UPDATE ON subscriber_lists
    FOR EACH ROW WHEN (
        (OLD.referred_by IS NOT NULL AND OLD.status = 'confirmed') OR
        (NEW.referred_by IS NOT NULL AND NEW.status = 'confirmed'))
    EXECUTE FUNCTION ca_referral_counts_trg();

DROP TRIGGER IF EXISTS ca_referral_counts_del ON subscriber_lists;
CREATE TRIGGER ca_referral_counts_del AFTER DELETE ON subscriber_lists
    FOR EACH ROW WHEN (OLD.referred_by IS NOT NULL AND OLD.status = 'confirmed')
    EXECUTE FUNCTION ca_referral_counts_trg();

-- Backfill from existing data.
DELETE FROM list_referral_counts;
INSERT INTO list_referral_counts (list_id, subscriber_id, count)
    SELECT list_id, referred_by, COUNT(*) FROM subscriber_lists
    WHERE referred_by IS NOT NULL AND status = 'confirmed'
    GROUP BY list_id, referred_by;
`
