package attach

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shaielc/code-winch/internal/runner"
)

type Server struct {
	static  http.Handler
	posture Posture
	session *runner.Session
}

func New(staticDir string, posture Posture, sessions ...*runner.Session) (http.Handler, error) {
	static, err := newStaticHandler(staticDir)
	if err != nil {
		return nil, err
	}
	var session *runner.Session
	if len(sessions) > 0 {
		session = sessions[0]
	}
	return &Server{static: static, posture: posture, session: session}, nil
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
		writeJSON(w, s.posture)
	case "/api/session/stream":
		if s.session == nil {
			http.NotFound(w, r)
			return
		}
		s.stream(w, r)
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
