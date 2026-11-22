-- Migration: 000005_timestamps_to_timestamptz (rollback to UTC wall-clock TIMESTAMP)

ALTER TABLE refresh_tokens
    ALTER COLUMN revoked_at TYPE TIMESTAMP USING revoked_at AT TIME ZONE 'UTC',
    ALTER COLUMN expires_at TYPE TIMESTAMP USING expires_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC';

ALTER TABLE examples
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN deleted_at TYPE TIMESTAMP USING deleted_at AT TIME ZONE 'UTC',
    ALTER COLUMN password_reset_expiry TYPE TIMESTAMP USING password_reset_expiry AT TIME ZONE 'UTC';
