package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/cugu/fomo/db/sqlc"
	"github.com/cugu/fomo/feed"
)

const pageSize = 10

func (s *Server) articlesPage(writer http.ResponseWriter, request *http.Request) { //nolint:funlen
	filter := request.URL.Query().Get("filter")
	page, offset := parsePage(request)
	query := request.URL.Query().Get("q")

	var (
		title    string
		articles []sqlc.Article
		err      error
	)

	switch filter {
	case "unread":
		title = "Unread articles"
		articles, err = s.queries.ListUnreadArticles(request.Context(), sqlc.ListUnreadArticlesParams{
			Offset: offset,
			Limit:  pageSize + 1,
		})
	case "bookmarked":
		title = "Bookmarked articles"
		articles, err = s.queries.ListBookmarkedArticles(request.Context(), sqlc.ListBookmarkedArticlesParams{
			Offset: offset,
			Limit:  pageSize + 1,
		})
	case "read":
		title = "Read articles"
		articles, err = s.queries.ListReadArticles(request.Context(), sqlc.ListReadArticlesParams{
			Offset: offset,
			Limit:  pageSize + 1,
		})
	case "search":
		title = "Search results"
		articles, err = s.queries.SearchArticles(request.Context(), sqlc.SearchArticlesParams{
			Query:  sql.NullString{String: query, Valid: true},
			Offset: offset,
			Limit:  pageSize + 1,
		})
	default:
		title = "Articles"
		articles, err = s.queries.ListArticles(request.Context(), sqlc.ListArticlesParams{
			Offset: offset,
			Limit:  pageSize + 1,
		})
	}

	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	hasNext := len(articles) > pageSize
	if hasNext {
		articles = articles[:pageSize]
	}

	s.template(writer, "articles", map[string]any{
		"Title":    title,
		"Filter":   filter,
		"Articles": articles,
		"Page":     page,
		"HasNext":  hasNext,
		"Previous": page - 1,
		"Next":     page + 1,
	})
}

func parsePage(request *http.Request) (page int, offset int64) {
	pageValue := request.URL.Query().Get("page")

	page, err := strconv.Atoi(pageValue)
	if err != nil {
		page = 1
	}

	return page, (int64(page) - 1) * pageSize
}

func (s *Server) articlePage(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	article, err := s.queries.Article(request.Context(), int64(id))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	s.template(writer, "article", map[string]any{
		"Title":   fmt.Sprintf("Article: %s", article.Title),
		"Article": article,
	})
}

func (s *Server) articleReFetch(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	article, err := s.queries.Article(request.Context(), int64(id))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	var feed feed.Feed

	for _, f := range s.feeds {
		if f.Name() == article.Feed {
			feed = f
			break
		}
	}

	if feed == nil {
		s.error("Feed not found", writer, request)
		return
	}

	updated, err := feed.ReFetch(request.Context(), &article)
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	replacedArticleID, err := s.queries.SetArticle(request.Context(), sqlc.SetArticleParams{
		Guid:        updated.Guid,
		Title:       updated.Title,
		Body:        updated.Body,
		PublishedAt: updated.PublishedAt,
		Link:        updated.Link,
		Feed:        updated.Feed,
		Details:     updated.Details,
		Read:        updated.Read,
		Bookmarked:  updated.Bookmarked,
	})
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, fmt.Sprintf("/articles/%d", replacedArticleID), http.StatusSeeOther)
}

func (s *Server) articleMarkRead(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	if err := s.queries.MarkReadArticle(request.Context(), int64(id)); err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, "/next", http.StatusSeeOther)
}

func (s *Server) articlesMarkRead(writer http.ResponseWriter, request *http.Request) {
	if err := s.queries.MarkReadAllArticles(request.Context()); err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, "/articles?filter=unread", http.StatusSeeOther)
}

func (s *Server) articleBookmark(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	a, err := s.queries.Article(request.Context(), int64(id))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	if a.Bookmarked {
		err = s.queries.UnbookmarkArticle(request.Context(), int64(id))
	} else {
		err = s.queries.BookmarkArticle(request.Context(), int64(id))
	}

	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, fmt.Sprintf("/articles/%d", id), http.StatusSeeOther)
}

func (s *Server) nextUnread(writer http.ResponseWriter, request *http.Request) {
	nextNewArticle, err := s.queries.NextUnreadArticle(request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Redirect(writer, request, "/articles?filter=unread", http.StatusSeeOther)
			return
		}

		s.error(err.Error(), writer, request)

		return
	}

	http.Redirect(writer, request, fmt.Sprintf("/articles/%d", nextNewArticle.ID), http.StatusSeeOther)
}

func (s *Server) articleEdit(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	article, err := s.queries.Article(request.Context(), int64(id))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	s.template(writer, "article_edit", map[string]any{
		"Title":   fmt.Sprintf("Edit: %s", article.Title),
		"Article": article,
	})
}

func (s *Server) articleUpdate(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	if err := request.ParseForm(); err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	article, err := s.queries.Article(request.Context(), int64(id))
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	publishedAt := time.Now()
	publishedAtValue := request.FormValue("published_at")

	if publishedAtValue != "" {
		publishedAt, err = time.Parse("2006-01-02T15:04:05", publishedAtValue)
		if err != nil {
			s.error(err.Error(), writer, request)
			return
		}
	}

	updatedID, err := s.queries.SetArticle(request.Context(), sqlc.SetArticleParams{
		Guid:        article.Guid,
		Title:       request.FormValue("title"),
		Body:        request.FormValue("body"),
		PublishedAt: publishedAt,
		Link:        article.Link,
		Feed:        request.FormValue("feed"),
		Details:     request.FormValue("details"),
		Read:        article.Read,
		Bookmarked:  article.Bookmarked,
	})
	if err != nil {
		s.error(err.Error(), writer, request)
	}

	http.Redirect(writer, request, fmt.Sprintf("/articles/%d", updatedID), http.StatusSeeOther)
}
