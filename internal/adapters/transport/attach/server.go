package attach

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Server struct {
	static  http.Handler
	session Session
}

func New(staticDir string, session Session) (http.Handler, error) {
	static, err := newStaticHandler(staticDir)
	if err != nil {
		return nil, err
	}
	return &Server{static: static, session: session}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch r.URL.Path {
	case "/healthz":
		writeJSON(w, map[string]string{"service": "winch-sandbox", "status": "ok"})
	case "/api/session":
		writeJSON(w, s.session.current())
	default:
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("{\"error\":\"not found\"}\n"))
			return
		}
		s.static.ServeHTTP(w, r)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
