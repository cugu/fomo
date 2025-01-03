CREATE TABLE IF NOT EXISTS "migrations" (version varchar(128) primary key);
CREATE TABLE sessions
(
    token  CHAR(43) PRIMARY KEY,
    data   BLOB         NOT NULL,
    expiry TIMESTAMP(6) NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);
CREATE TABLE articles
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
    read         BOOLEAN       NOT NULL DEFAULT FALSE,
    bookmarked   BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Dbmate schema migrations
INSERT INTO "migrations" (version) VALUES
  ('00010101000000'),
  ('00010101000001');
