package server

import (
	"net/http"
)

const sessionKey = "authenticated"

func (s *Server) login(writer http.ResponseWriter, request *http.Request) {
	s.template(writer, "login", map[string]any{
		"Title": "Login",
		"Error": request.URL.Query().Get("error"),
	})
}

func (s *Server) validateLogin(writer http.ResponseWriter, request *http.Request) {
	if request.FormValue("password") != s.password {
		http.Redirect(writer, request, "/login?error=Password+incorrect", http.StatusFound)
		return
	}

	s.sessionManager.Put(request.Context(), sessionKey, true)
	http.Redirect(writer, request, "/", http.StatusFound)
}

func (s *Server) logout(writer http.ResponseWriter, request *http.Request) {
	s.sessionManager.Remove(request.Context(), sessionKey)
	http.Redirect(writer, request, "/login", http.StatusFound)
}

func (s *Server) requireAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.sessionManager.GetBool(r.Context(), sessionKey) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
