package realtime

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"hrig/shared/model"
)

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}

func TestWebSocketEndToEnd_SnapshotBroadcastAckAndOriginCheck(t *testing.T) {
	epoch := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	hub := NewHub(15*time.Second, epoch)
	defer hub.Close()

	allowed := []string{"http://localhost:9245", "wails://wails"}
	srv := httptest.NewServer(ServeWS(hub, allowed))
	defer srv.Close()

	// Disallowed Origin is rejected
	badHeader := http.Header{"Origin": []string{"http://evil.example"}}
	if _, resp, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), badHeader); err == nil {
		t.Fatal("expected disallowed origin handshake to fail")
	} else if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for disallowed origin, got %d", resp.StatusCode)
	}

	// Allowed Origin (case-insensitive) succeeds
	goodHeader := http.Header{"Origin": []string{"HTTP://LOCALHOST:9245"}}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), goodHeader)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	// 1. Receive initial snapshot
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var snapEnv model.RealtimeEnvelope
	if err := conn.ReadJSON(&snapEnv); err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
	if snapEnv.Type != "snapshot" {
		t.Fatalf("expected snapshot envelope, got %q", snapEnv.Type)
	}
	if hub.ClientCount() != 1 {
		t.Fatalf("expected ClientCount == 1, got %d", hub.ClientCount())
	}

	// 2. Publish critical reading and receive live reading + alert
	val := 195.0
	unit := "/min"
	hub.PublishFrames([]model.TelemetryFrame{
		{
			Source:        "wearable",
			DeviceCode:    "biomonitor_01",
			Environment:   model.EnvIVA,
			MetricType:    1,
			MetricCode:    "heart_rate",
			UnitCode:      &unit,
			ValueNumeric:  &val,
			Status:        model.StatusCritical,
			RecordedAtUTC: time.Now().UTC(),
		},
	})

	var readingEnv model.RealtimeEnvelope
	if err := conn.ReadJSON(&readingEnv); err != nil {
		t.Fatalf("read reading envelope: %v", err)
	}
	if readingEnv.Type != "reading" || readingEnv.Reading == nil || readingEnv.Reading.MetricCode != "heart_rate" {
		t.Fatalf("unexpected reading envelope: %+v", readingEnv)
	}

	var alertEnv model.RealtimeEnvelope
	if err := conn.ReadJSON(&alertEnv); err != nil {
		t.Fatalf("read alert envelope: %v", err)
	}
	if alertEnv.Type != "alert" || alertEnv.Alert == nil || alertEnv.Alert.ID == "" {
		t.Fatalf("unexpected alert envelope: %+v", alertEnv)
	}

	// 3. Send inbound alert_ack over WS and receive alert_ack broadcast
	ackBy := "flight-surgeon"
	if err := conn.WriteJSON(model.RealtimeEnvelope{
		Type: "alert_ack",
		Alert: &model.AlertEvent{
			ID:             alertEnv.Alert.ID,
			AcknowledgedBy: &ackBy,
		},
		SentAtUTC: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("write alert_ack: %v", err)
	}

	var ackEnv model.RealtimeEnvelope
	if err := conn.ReadJSON(&ackEnv); err != nil {
		t.Fatalf("read alert_ack envelope: %v", err)
	}
	if ackEnv.Type != "alert_ack" || ackEnv.Alert == nil || !ackEnv.Alert.Acknowledged {
		t.Fatalf("unexpected alert_ack envelope: %+v", ackEnv)
	}
	if ackEnv.Alert.AcknowledgedBy == nil || *ackEnv.Alert.AcknowledgedBy != ackBy {
		t.Fatalf("expected acknowledgedBy=%q, got %v", ackBy, ackEnv.Alert.AcknowledgedBy)
	}
}

func TestWebSocketSlowClient_DoesNotBlockPublishFrames(t *testing.T) {
	hub := NewHub(15*time.Second, time.Now().UTC())
	defer hub.Close()

	// Register a slow client with no pump draining its send buffer
	slowClient := &Client{
		hub:  hub,
		send: make(chan model.RealtimeEnvelope, sendBufferSize),
	}
	if !hub.register(slowClient) {
		t.Fatal("failed to register slow client")
	}

	// Also connect a real fast WebSocket client
	srv := httptest.NewServer(ServeWS(hub, nil))
	defer srv.Close()

	fastConn, _, err := websocket.DefaultDialer.Dial(wsURL(srv.URL), nil)
	if err != nil {
		t.Fatalf("dial fast client: %v", err)
	}
	defer fastConn.Close()

	_ = fastConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var initial model.RealtimeEnvelope
	if err := fastConn.ReadJSON(&initial); err != nil {
		t.Fatalf("read fast client initial snapshot: %v", err)
	}

	// Continuously drain fastConn so the fast client's send buffer stays clear
	received := make(chan model.RealtimeEnvelope, sendBufferSize+30)
	readErrCh := make(chan error, 1)
	go func() {
		for {
			var env model.RealtimeEnvelope
			if err := fastConn.ReadJSON(&env); err != nil {
				readErrCh <- err
				return
			}
			received <- env
		}
	}()

	// Flood more than sendBufferSize total messages across two bursts while fastConn drains;
	// slowClient has no reader so its buffer overflows and it is dropped without blocking PublishFrames.
	done := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		unit := "kPa"
		publishOne := func(i int) {
			v := 100.0 + float64(i)*0.01
			hub.PublishFrames([]model.TelemetryFrame{
				{
					Source:        "environment_sensor",
					DeviceCode:    "cabin_sensor_01",
					Environment:   model.EnvIVA,
					MetricType:    2,
					MetricCode:    fmt.Sprintf("metric_%d", i%5),
					UnitCode:      &unit,
					ValueNumeric:  &v,
					Status:        model.StatusNormal,
					RecordedAtUTC: time.Now().UTC(),
				},
			})
		}

		const firstBurst = sendBufferSize - 24 // 40 (< sendBufferSize)
		const total = sendBufferSize + 20      // 84 (> sendBufferSize)
		for i := 0; i < firstBurst; i++ {
			publishOne(i)
		}

		// Wait until fastConn has drained at least 30 messages from the first burst
		for drained := 0; drained < 30; drained++ {
			select {
			case <-received:
			case err := <-readErrCh:
				errCh <- fmt.Errorf("fast client read failed early: %w", err)
				close(done)
				return
			case <-time.After(time.Second):
				errCh <- fmt.Errorf("timed out waiting for fast client to drain")
				close(done)
				return
			}
		}

		for i := firstBurst; i < total; i++ {
			publishOne(i)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("PublishFrames blocked on slow client")
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	// Slow client should have been dropped, leaving only fastClient
	if got := hub.ClientCount(); got != 1 {
		t.Fatalf("expected slow client dropped and ClientCount == 1, got %d", got)
	}

	// Fast client continues receiving remaining reading frames cleanly
	select {
	case gotReading := <-received:
		if gotReading.Type != "reading" {
			t.Fatalf("expected reading frame, got %q", gotReading.Type)
		}
	case err := <-readErrCh:
		t.Fatalf("fast client failed to read frame: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remaining reading frame on fast client")
	}
}
