CREATE TABLE post_likes (
    user_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    post_id    BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);
CREATE INDEX idx_post_likes_post ON post_likes(post_id);

CREATE TABLE comment_votes (
    user_id    BIGINT   REFERENCES users(id) ON DELETE CASCADE,
    comment_id BIGINT   REFERENCES comments(id) ON DELETE CASCADE,
    value      SMALLINT NOT NULL CHECK (value IN (1, -1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, comment_id)
);
CREATE INDEX idx_comment_votes_comment ON comment_votes(comment_id);
