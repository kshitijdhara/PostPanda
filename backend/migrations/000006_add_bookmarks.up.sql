CREATE TABLE post_bookmarks (
    user_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    post_id    BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);
CREATE INDEX idx_post_bookmarks_user ON post_bookmarks(user_id);
