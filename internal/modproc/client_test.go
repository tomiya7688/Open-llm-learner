package modproc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

func TestInvokeAndEvent(t *testing.T) {
	client := startHelper(t, "normal")
	defer shutdownHelper(t, client)

	var mu sync.Mutex
	var events []string
	out, err := client.Invoke(context.Background(), "echo", map[string]any{"value": 7}, func(e protocol.Event) {
		mu.Lock()
		events = append(events, e.Name)
		mu.Unlock()
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got["value"].(float64) != 7 {
		t.Fatalf("unexpected output: %s", out)
	}
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(events) == 0 || events[0] != "status" {
		t.Fatalf("expected status event, got %v", events)
	}
}

func TestTimeoutSendsCancel(t *testing.T) {
	client := startHelper(t, "normal")
	defer shutdownHelper(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := client.Invoke(ctx, "block", map[string]any{}, nil)
	var remote *RemoteError
	if !errors.As(err, &remote) || remote.Code != "TIMEOUT" {
		t.Fatalf("expected TIMEOUT RemoteError, got %T %v", err, err)
	}
}

func TestIgnoredCancelForceStopsProcess(t *testing.T) {
	client := startHelper(t, "ignore-cancel")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := client.Invoke(ctx, "block", map[string]any{}, nil)
	var remote *RemoteError
	if !errors.As(err, &remote) || remote.Code != "TIMEOUT" {
		t.Fatalf("expected TIMEOUT RemoteError, got %T %v", err, err)
	}
}

func startHelper(t *testing.T, mode string) *Client {
	t.Helper()
	client, err := Start(context.Background(), Config{
		Executable:         os.Args[0],
		Args:               []string{"-test.run=TestHelperProcess", "--"},
		Env:                append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_MODE="+mode),
		ExpectedModID:      "test.helper",
		ExpectedModVersion: "0.1.0",
		HandshakeTimeout:   time.Second,
		CancellationGrace:  60 * time.Millisecond,
		ShutdownGrace:      200 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func shutdownHelper(t *testing.T, client *Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil && !strings.Contains(err.Error(), "terminated") {
		t.Logf("shutdown helper: %v", err)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	mode := os.Getenv("HELPER_MODE")
	dec := protocol.NewDecoder(os.Stdin, protocol.DefaultMaxMessageBytes)
	var writeMu sync.Mutex
	write := func(msg protocol.Message) {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = protocol.Encode(os.Stdout, msg)
	}
	blocked := map[string]chan struct{}{}
	var blockedMu sync.Mutex
	for {
		var msg protocol.Message
		if err := dec.Decode(&msg); err != nil {
			os.Exit(0)
		}
		switch msg.Type {
		case "hello":
			write(protocol.Message{Type: "hello", Protocol: protocol.Version, Mod: &protocol.ModInfo{ID: "test.helper", Version: "0.1.0"}})
		case "invoke":
			if msg.Action == "echo" {
				write(protocol.Message{Type: "event", ID: msg.ID, Event: "status", Data: json.RawMessage(`"working"`)})
				write(protocol.Message{Type: "result", ID: msg.ID, Output: msg.Input})
				continue
			}
			if msg.Action == "block" {
				blockedMu.Lock()
				blocked[msg.ID] = make(chan struct{})
				blockedMu.Unlock()
			}
		case "cancel":
			if mode == "ignore-cancel" {
				continue
			}
			blockedMu.Lock()
			if _, ok := blocked[msg.ID]; ok {
				delete(blocked, msg.ID)
				write(protocol.Message{Type: "cancelled", ID: msg.ID})
			}
			blockedMu.Unlock()
		case "shutdown":
			write(protocol.Message{Type: "shutdown_ack"})
			os.Exit(0)
		}
	}
}
