package feed

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cugu/fomo/db/sqlc"
)

func init() {
	RegisterGenerator("lobsters", func(name string, _ json.RawMessage) (Feed, error) {
		return &Lobsters{name: name}, nil
	})
}

type Lobsters struct {
	name string
}

func NewLobsters(name string) *Lobsters {
	return &Lobsters{name: name}
}

func (s *Lobsters) Name() string {
	return s.name
}

type LobsterResponse struct {
	ShortIDURL  string    `json:"short_id_url"`
	CreatedAt   time.Time `json:"created_at"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Score       int       `json:"score"`
	CommentsURL string    `json:"comments_url"`
}

func (s *Lobsters) Fetch(ctx context.Context, seen SeenFunc) ([]*sqlc.Article, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://lobste.rs/hottest.json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response []LobsterResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	var articles []*sqlc.Article

	for _, item := range response {
		if item.Score < 10 || seen(ctx, item.ShortIDURL) {
			continue
		}

		content, err := loadContent(ctx, item.URL, "", "")
		if err != nil {
			content = fmt.Sprintf("failed to load content: %s", err.Error())
		}

		u, _ := url.Parse(item.URL)

		details := u.Hostname()
		if item.CommentsURL != "" {
			details += fmt.Sprintf(" | <a href=\"%s\">comments</a>", item.CommentsURL)
		}

		articles = append(articles, &sqlc.Article{
			Guid:        item.ShortIDURL,
			Title:       item.Title,
			Body:        content,
			PublishedAt: item.CreatedAt,
			Link:        cmp.Or(item.URL, item.ShortIDURL),
			Feed:        s.name,
			Details:     details,
		})
	}

	return articles, nil
}

func (s *Lobsters) ReFetch(ctx context.Context, a *sqlc.Article) (*sqlc.Article, error) {
	content, err := loadContent(ctx, a.Link, "", "")
	if err != nil {
		content = fmt.Sprintf("failed to load content: %s", err.Error())
	}

	a.Body = content

	return a, nil
}
