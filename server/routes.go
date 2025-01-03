package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) ListenAndServe(port int) error {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(s.sessionManager.LoadAndSave)
	router.Use(middleware.Compress(5))

	router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/articles?filter=unread", http.StatusFound)
	})

	// auth
	router.Get("/login", s.login)
	router.Post("/login", s.validateLogin)
	router.Get("/logout", s.logout)

	// serve static files
	router.Get("/static/*", s.static)

	// private routes
	router.Group(func(privateRouter chi.Router) {
		privateRouter.Use(s.requireAuth)
		privateRouter.Get("/fetch", s.fetchFeeds)

		// articles
		privateRouter.Get("/articles", s.articlesPage)
		privateRouter.Post("/articles/read", s.articlesMarkRead)
		privateRouter.Post("/articles", s.addArticle)

		// article
		privateRouter.Get("/articles/{id}", s.articlePage)
		privateRouter.Get("/next", s.nextUnread)
		privateRouter.Get("/articles/{id}/edit", s.articleEdit)
		privateRouter.Post("/articles/{id}/edit", s.articleUpdate)
		privateRouter.Post("/articles/{id}/refetch", s.articleReFetch)
		privateRouter.Post("/articles/{id}/read", s.articleMarkRead)
		privateRouter.Post("/articles/{id}/bookmark", s.articleBookmark)
	})

	timeoutServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return timeoutServer.ListenAndServe()
}
