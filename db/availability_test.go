package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/cugu/fomo/db/sqlc"
)

func TestArticleAvailability(t *testing.T) { //nolint:cyclop
	t.Parallel()

	database, queries, err := DB(t.TempDir())
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}

	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	createArticle(ctx, t, queries, "visible", sql.NullTime{})
	createArticle(ctx, t, queries, "released", sql.NullTime{
		Time:  time.Now().In(time.FixedZone("test", 2*60*60)).Add(-24 * time.Hour),
		Valid: true,
	})

	hiddenID := createArticle(ctx, t, queries, "hidden", sql.NullTime{
		Time:  time.Now().Add(24 * time.Hour),
		Valid: true,
	})

	articles, err := queries.ListArticles(ctx, sqlc.ListArticlesParams{Limit: 10})
	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("ListArticles() returned %d articles, want 2 visible articles", len(articles))
	}

	if _, err := queries.Article(ctx, hiddenID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Article(hidden) error = %v, want sql.ErrNoRows", err)
	}

	if _, err := queries.ArticleIDByGUID(ctx, "hidden"); err != nil {
		t.Fatalf("ArticleIDByGUID(hidden) error = %v", err)
	}

	if err := queries.MarkReadAllArticles(ctx); err != nil {
		t.Fatalf("MarkReadAllArticles() error = %v", err)
	}

	if err := queries.ReleasePendingArticles(ctx); err != nil {
		t.Fatalf("ReleasePendingArticles() error = %v", err)
	}

	hidden, err := queries.Article(ctx, hiddenID)
	if err != nil {
		t.Fatalf("Article(released) error = %v", err)
	}

	if hidden.Read {
		t.Error("hidden article was marked read before being released")
	}

	if hidden.AvailableAt.Valid {
		t.Errorf("released article AvailableAt = %v, want NULL", hidden.AvailableAt)
	}
}

func createArticle(
	ctx context.Context,
	t *testing.T,
	queries *sqlc.Queries,
	guid string,
	availableAt sql.NullTime,
) int64 {
	t.Helper()

	id, err := queries.CreateArticle(ctx, sqlc.CreateArticleParams{
		Guid:        guid,
		Title:       guid,
		Body:        guid,
		PublishedAt: time.Now(),
		Link:        "https://example.com/" + guid,
		Feed:        "test",
		Details:     "",
		AvailableAt: availableAt,
	})
	if err != nil {
		t.Fatalf("CreateArticle(%q) error = %v", guid, err)
	}

	return id
}
