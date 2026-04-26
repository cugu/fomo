-- migrate:up
CREATE INDEX IF NOT EXISTS idx_articles_published_at_desc
    ON articles (published_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_read_published_at_desc
    ON articles (read, published_at DESC);

CREATE INDEX IF NOT EXISTS idx_articles_bookmarked_published_at_desc
    ON articles (bookmarked, published_at DESC);

-- migrate:down
DROP INDEX IF EXISTS idx_articles_bookmarked_published_at_desc;
DROP INDEX IF EXISTS idx_articles_read_published_at_desc;
DROP INDEX IF EXISTS idx_articles_published_at_desc;
