-- SlackLite Database Schema

CREATE TABLE IF NOT EXISTS channels (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    author TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
);

-- Index for faster message queries by channel
CREATE INDEX IF NOT EXISTS idx_messages_channel_id ON messages(channel_id);

-- Index for message ordering
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
