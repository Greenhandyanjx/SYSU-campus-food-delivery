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

-- cleanup: 移除旧的 FOR EACH ROW trigger，它会在每次插入消息时覆盖
-- updated_at，导致 JSONL 迁移过来的会话日期全部变成当前时间
DROP TRIGGER IF EXISTS trg_messages_touch_session ON messages;
DROP FUNCTION IF EXISTS touch_session_updated_at;
