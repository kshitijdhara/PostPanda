CREATE TABLE comments (
    id         BIGSERIAL PRIMARY KEY,
    content    TEXT      NOT NULL,
    post_id    BIGINT    NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id  BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id  BIGINT    REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
