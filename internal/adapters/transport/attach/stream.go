package attach

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/shaielc/code-winch/internal/runner"
)

// recordWriteLimit is the largest record the stream writes. The runner bounds
// every record to this size, so exceeding it is a defect, not a payload to drop.
const recordWriteLimit = runner.MaxRecordBytes

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
