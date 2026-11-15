package realtime

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"hrig/shared/model"
)

const (
	sendBufferSize = 64
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 64 << 10
)

// Client bridges a single Gorilla WebSocket connection to a Hub.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan model.RealtimeEnvelope
}

// ServeWS upgrades HTTP requests to Gorilla WebSocket connections with origin validation.
func ServeWS(hub *Hub, allowedOrigins []string) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return true
			}
			for _, allowed := range allowedOrigins {
				if strings.EqualFold(origin, strings.TrimSpace(allowed)) {
					return true
				}
			}
			return false
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil || hub.isClosed() {
			http.Error(w, "realtime hub unavailable", http.StatusServiceUnavailable)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan model.RealtimeEnvelope, sendBufferSize),
		}

		if !hub.register(client) {
			_ = conn.Close()
			return
		}

		go client.writePump()
		go client.readPump()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var env model.RealtimeEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}

		if env.Type == "alert_ack" && env.Alert != nil {
			alertID := strings.TrimSpace(env.Alert.ID)
			if alertID == "" {
				continue
			}
			ackBy := ""
			if env.Alert.AcknowledgedBy != nil {
				ackBy = *env.Alert.AcknowledgedBy
			}
			_, _ = c.hub.AckAlert(alertID, ackBy)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case env, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				return
			}
			if err := c.conn.WriteJSON(env); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
