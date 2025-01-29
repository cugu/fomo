package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-shiori/go-readability"

	"github.com/cugu/fomo/db/sqlc"
)

var feeds sync.Map

type Generator func(string, json.RawMessage) (Feed, error)

func RegisterGenerator(feedType string, generator Generator) {
	feeds.Store(feedType, generator)
}

func LookupFeed(feedType string) (Generator, bool) {
	feed, ok := feeds.Load(feedType)
	if !ok {
		return nil, false
	}

	if f, ok := feed.(Generator); ok {
		return f, true
	}

	return nil, false
}

type SeenFunc func(context.Context, string) bool

type Feed interface {
	Name() string
	Fetch(ctx context.Context, seen SeenFunc) ([]*sqlc.Article, error)
	ReFetch(ctx context.Context, article *sqlc.Article) (*sqlc.Article, error)
}

func loadContent(ctx context.Context, link, username, password string) (string, error) {
	linkURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	resp, err := request(ctx, link, username, password)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	article, err := readability.FromReader(resp.Body, linkURL)
	if err != nil {
		return "", err
	}

	return article.Content, nil
}

func request(ctx context.Context, link, username, password string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	return http.DefaultClient.Do(req)
}
