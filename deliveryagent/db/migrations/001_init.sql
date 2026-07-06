CREATE TABLE IF NOT EXISTS sessions (
    key                TEXT PRIMARY KEY,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata           JSONB DEFAULT '{}',
    last_consolidated  INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS messages (
    id               BIGSERIAL PRIMARY KEY,
    session_key      TEXT NOT NULL REFERENCES sessions(key) ON DELETE CASCADE,
    role             TEXT NOT NULL,
    content          TEXT NOT NULL DEFAULT '',
    timestamp        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tool_calls       JSONB,
    tool_call_id     TEXT,
    name             TEXT
);

CREATE INDEX IF NOT EXISTS idx_messages_session_ts
    ON messages(session_key, timestamp, id);

CREATE INDEX IF NOT EXISTS idx_messages_role
    ON messages(role);

CREATE OR REPLACE FUNCTION touch_session_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE sessions SET updated_at = NOW()
    WHERE key = COALESCE(NEW.session_key, OLD.session_key);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_messages_touch_session ON messages;

CREATE TRIGGER trg_messages_touch_session
    AFTER INSERT OR UPDATE OR DELETE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION touch_session_updated_at();
