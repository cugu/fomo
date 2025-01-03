package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-shiori/go-readability"

	"github.com/cugu/fomo/db/sqlc"
	"github.com/cugu/fomo/feed"
)

func (s *Server) fetchFeeds(writer http.ResponseWriter, request *http.Request) {
	var errs []error

	for _, f := range s.feeds {
		errs = append(errs, feed.Fetch(request.Context(), s.queries, f))
	}

	if err := errors.Join(errs...); err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

// addArticle adds a new article to the database via a form submission.
func (s *Server) addArticle(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	target := request.Form.Get("url")

	targetURL, err := url.Parse(target)
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	article, err := readability.FromURL(target, 10*time.Second)
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	newArticleID, err := s.queries.SetArticle(request.Context(), sqlc.SetArticleParams{
		Guid:        target,
		Title:       article.Title,
		Body:        article.Content,
		PublishedAt: time.Now(),
		Link:        target,
		Feed:        "",
		Details:     targetURL.Hostname(),
		Read:        false,
		Bookmarked:  false,
	})
	if err != nil {
		s.error(err.Error(), writer, request)
		return
	}

	http.Redirect(writer, request, fmt.Sprintf("/articles/%d", newArticleID), http.StatusSeeOther)
}
