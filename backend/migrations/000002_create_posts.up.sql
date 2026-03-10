CREATE TABLE posts (
    id           BIGSERIAL PRIMARY KEY,
    title        VARCHAR(300) NOT NULL,
    slug         VARCHAR(350) UNIQUE NOT NULL,
    content      TEXT         NOT NULL,
    excerpt      VARCHAR(500),
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    author_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_posts_slug ON posts(slug);
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status_published ON posts(status, published_at DESC) WHERE status = 'published';
CREATE INDEX idx_posts_search ON posts USING gin(to_tsvector('english', title || ' ' || content));
