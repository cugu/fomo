-- migrate:up
ALTER TABLE articles
    ADD COLUMN available_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_articles_available_at
    ON articles (datetime(available_at));

-- migrate:down
DROP INDEX IF EXISTS idx_articles_available_at;
ALTER TABLE articles DROP COLUMN available_at;
