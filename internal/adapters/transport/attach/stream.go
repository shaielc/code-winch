package attach

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
)

const recordWriteLimit = 64 * 1024

func (s *Server) stream(w http.ResponseWriter, request *http.Request) {
	connection, err := websocket.Accept(w, request, nil)
	if err != nil {
		return
	}
	defer func() { _ = connection.CloseNow() }()
	connection.SetReadLimit(recordWriteLimit)
	records, unsubscribe := s.session.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-request.Context().Done():
			return
		case record := <-records:
			data, err := json.Marshal(record)
			if err != nil || len(data) > recordWriteLimit {
				_ = connection.Close(websocket.StatusInternalError, "record unavailable")
				return
			}
			if err := connection.Write(request.Context(), websocket.MessageText, data); err != nil {
				if websocket.CloseStatus(err) == -1 && err != context.Canceled {
					_ = connection.Close(websocket.StatusInternalError, "write failed")
				}
				return
			}
		}
	}
}
