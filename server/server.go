package server

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/cugu/fomo/db/sqlc"
	"github.com/cugu/fomo/feed"
	"github.com/cugu/fomo/ui"
)

type Server struct {
	baseURL        string
	password       string
	updateTimes    []int
	queries        *sqlc.Queries
	feeds          []feed.Feed
	sessionManager *scs.SessionManager
	templates      *template.Template
}

func New(
	baseURL string,
	password string,
	updateTimes []int,
	feeds []feed.Feed,
	queries *sqlc.Queries,
) *Server {
	sessionManager := scs.New()
	sessionManager.Lifetime = time.Hour * 24 * 7
	sessionManager.Store = &SQLiteStore{DB: queries}

	return &Server{
		baseURL:        baseURL,
		password:       password,
		updateTimes:    updateTimes,
		queries:        queries,
		feeds:          feeds,
		sessionManager: sessionManager,
		templates:      ui.Templates(),
	}
}

func (s *Server) static(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "public, max-age=3600")

	http.FileServer(http.FS(ui.StaticFS)).ServeHTTP(writer, request)
}

func (s *Server) error(msg string, writer http.ResponseWriter, _ *http.Request) {
	s.template(writer, "error", map[string]any{
		"Title": "Error",
		"Error": msg,
	})
}

func (s *Server) template(writer http.ResponseWriter, title string, data map[string]any) {
	data["BaseURL"] = s.baseURL

	zone, _ := time.Now().Zone()
	data["UpdateTimes"] = fmt.Sprintf("%s (%s)", formatTimes(s.updateTimes), zone)

	if err := s.templates.ExecuteTemplate(writer, title, data); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// formatTimes returns a string representation of the update times.
// e.g. "6:00AM, 12:00PM, and 6:00" or "6:00 and 12:00".
func formatTimes(updateTimes []int) string {
	if len(updateTimes) == 0 {
		return ""
	}

	slices.Sort(updateTimes)

	var times []string

	for _, h := range updateTimes {
		d := time.Date(0, 0, 0, h, 0, 0, 0, time.UTC)
		times = append(times, d.Format(time.Kitchen))
	}

	switch {
	case len(times) == 1:
		return times[0]
	case len(times) == 2:
		return times[0] + " and " + times[1]
	default:
		return strings.Join(times[:len(times)-1], ", ") + ", and " + times[len(times)-1]
	}
}
