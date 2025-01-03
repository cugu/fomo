-- migrate:up
CREATE TABLE IF NOT EXISTS articles
(
    id           INTEGER PRIMARY KEY,

    -- rss fields
    guid         VARCHAR(255)  NOT NULL UNIQUE,
    title        TEXT          NOT NULL,
    body         TEXT          NOT NULL,
    published_at TIMESTAMP     NOT NULL,
    link         VARCHAR(2048) NOT NULL UNIQUE,

    -- custom metadata fields
    details      TEXT          NOT NULL,
    feed         VARCHAR(255)  NOT NULL,
    read         BOOLEAN       NOT NULL,
    bookmarked   BOOLEAN       NOT NULL
);

-- migrate:down
DROP TABLE IF EXISTS articles;
