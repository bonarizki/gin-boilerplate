-- Store every timestamp as TIMESTAMPTZ so values are absolute instants and
-- no longer depend on the session or process time zone.
-- Existing TIMESTAMP values are interpreted as UTC, which is what the app and
-- the postgres containers have used. If a deployment wrote local wall-clock
-- times, change 'UTC' below to that zone before running this migration.
-- Migration: 000005_timestamps_to_timestamptz

ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN deleted_at TYPE TIMESTAMPTZ USING deleted_at AT TIME ZONE 'UTC',
    ALTER COLUMN password_reset_expiry TYPE TIMESTAMPTZ USING password_reset_expiry AT TIME ZONE 'UTC';

ALTER TABLE examples
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE refresh_tokens
    ALTER COLUMN revoked_at TYPE TIMESTAMPTZ USING revoked_at AT TIME ZONE 'UTC',
    ALTER COLUMN expires_at TYPE TIMESTAMPTZ USING expires_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC';
