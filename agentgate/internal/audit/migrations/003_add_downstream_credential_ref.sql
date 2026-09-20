ALTER TABLE audit_events
    ADD COLUMN IF NOT EXISTS downstream_credential_ref VARCHAR(256) NOT NULL DEFAULT '';
