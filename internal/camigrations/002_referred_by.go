package camigrations

// sql002ReferredBy adds subscriber_lists.referred_by: the subscriber whose
// referral code was used when this subscription was created. It is set once,
// at subscription time, and is cleared (not cascaded) if the referrer is deleted.
const sql002ReferredBy = `
ALTER TABLE subscriber_lists
    ADD COLUMN IF NOT EXISTS referred_by INTEGER NULL
    REFERENCES subscribers(id) ON DELETE SET NULL ON UPDATE CASCADE;
CREATE INDEX IF NOT EXISTS idx_sub_lists_referred_by ON subscriber_lists(referred_by) WHERE referred_by IS NOT NULL;
`
