BEGIN;
    -- Backfill `status` for historical entries
    UPDATE subscriptions
        SET status = NULL
        WHERE status = 'confirmed';
    ALTER TABLE subscriptions ALTER COLUMN status SET NULL;
COMMIT;