package feed

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cugu/fomo/db/sqlc"
)

func Fetch(ctx context.Context, queries *sqlc.Queries, f Feed) error {
	seen := func(ctx context.Context, guid string) bool {
		if _, err := queries.ArticleIDByGUID(ctx, guid); err != nil {
			return false
		}

		return true
	}

	articles, err := f.Fetch(ctx, seen)
	if err != nil {
		return err
	}

	var errs []error

	for _, article := range articles {
		if len(article.Body) > 1_000_000 {
			article.Body = "error: MAX BODY SIZE REACHED" + article.Body[:1_000_000]
		}

		_, err := queries.CreateArticle(ctx, sqlc.CreateArticleParams{
			Guid:        article.Guid,
			Title:       article.Title,
			Body:        article.Body,
			PublishedAt: article.PublishedAt,
			Link:        article.Link,
			Feed:        article.Feed,
			Details:     article.Details,
			Read:        article.Read,
			Bookmarked:  article.Bookmarked,
		})

		// ignore duplicate entries
		if !errors.Is(err, sql.ErrNoRows) {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
