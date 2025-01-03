package feed

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"

	"github.com/cugu/fomo/db/sqlc"
)

func init() {
	RegisterGenerator("rss", newRSS)
}

type RSS struct {
	name   string
	config *RSSConfig
}

type RSSConfig struct {
	URL              string `json:"url"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	FetchLinkContent bool   `json:"fetch_link_content"`
}

func newRSS(name string, config json.RawMessage) (Feed, error) {
	var cfg RSSConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RSS config: %w", err)
	}

	return NewRSSWithConfig(name, &cfg), nil
}

func NewRSSWithConfig(name string, config *RSSConfig) *RSS {
	return &RSS{name: name, config: config}
}

func (s *RSS) Fetch(ctx context.Context, seen SeenFunc) ([]*sqlc.Article, error) {
	resp, err := request(ctx, s.config.URL, s.config.Username, s.config.Password)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if gofeed.DetectFeedType(bytes.NewReader(b)) == gofeed.FeedTypeRSS {
		return s.fetchRSS(ctx, seen, bytes.NewReader(b))
	}

	return s.fetchFeed(ctx, bytes.NewReader(b))
}

func (s *RSS) fetchRSS(ctx context.Context, seen SeenFunc, body io.Reader) ([]*sqlc.Article, error) {
	slog.InfoContext(ctx, "Fetching RSS feed", "name", s.name)

	fp := &rss.Parser{}

	feed, err := fp.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	var articles []*sqlc.Article

	for _, item := range feed.Items {
		// skip articles that have already been seen, if we're not fetching link content
		// this prevents us from fetching the same article multiple times
		if s.config.FetchLinkContent && seen(ctx, item.GUID.Value) {
			continue
		}

		content := item.Description
		details := ""

		if s.config.FetchLinkContent {
			content, err = loadContent(ctx, item.Link, s.config.Username, s.config.Password)
			if err != nil {
				content = fmt.Sprintf("failed to load content: %s,\ndescription: %s", err.Error(), item.Description)
			}

			u, _ := url.Parse(item.Link)

			details = u.Hostname()
			if item.Comments != "" {
				details += fmt.Sprintf(" | <a href=\"%s\">comments</a>", item.Comments)
			}
		}

		articles = append(articles, &sqlc.Article{
			Guid:        item.GUID.Value,
			Title:       item.Title,
			Body:        content,
			PublishedAt: *cmp.Or(item.PubDateParsed, &time.Time{}),
			Link:        cmp.Or(item.Link, item.GUID.Value),
			Feed:        s.name,
			Details:     details,
		})
	}

	return articles, nil
}

func (s *RSS) fetchFeed(ctx context.Context, body io.Reader) ([]*sqlc.Article, error) {
	slog.InfoContext(ctx, "Fetching feed", "name", s.name)

	feed, err := gofeed.NewParser().Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var articles []*sqlc.Article

	for _, item := range feed.Items {
		articles = append(articles, &sqlc.Article{
			Guid:        item.GUID,
			Title:       item.Title,
			Body:        item.Description,
			PublishedAt: *cmp.Or(item.PublishedParsed, &time.Time{}),
			Link:        cmp.Or(item.Link, item.GUID),
			Feed:        s.name,
		})
	}

	return articles, nil
}

func (s *RSS) ReFetch(ctx context.Context, a *sqlc.Article) (*sqlc.Article, error) {
	if s.config.FetchLinkContent {
		return s.reFetchLink(ctx, a)
	}

	return s.reFetchRSS(ctx, a)
}

func (s *RSS) reFetchLink(ctx context.Context, article *sqlc.Article) (*sqlc.Article, error) {
	content, err := loadContent(ctx, article.Link, s.config.Username, s.config.Password)
	if err != nil {
		content = fmt.Sprintf("failed to load content: %s", err.Error())
	}

	article.Body = content

	return article, nil
}

func (s *RSS) reFetchRSS(ctx context.Context, a *sqlc.Article) (*sqlc.Article, error) {
	articles, err := s.Fetch(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	for _, article := range articles {
		if article.Guid == a.Guid {
			return article, nil
		}
	}

	return nil, errors.New("article not found")
}
