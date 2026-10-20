package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

type server struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func main() {
	s := &server{cancels: make(map[string]context.CancelFunc)}
	dec := protocol.NewDecoder(bufio.NewReader(os.Stdin), protocol.DefaultMaxMessageBytes)
	for {
		var msg protocol.Message
		if err := dec.Decode(&msg); err != nil {
			return
		}
		switch msg.Type {
		case "hello":
			if !has(msg.Protocols, protocol.Version) {
				s.write(protocol.Message{Type: "error", Error: &protocol.ErrorPayload{Code: "PROTOCOL_ERROR", Message: "no compatible protocol"}})
				os.Exit(2)
			}
			s.write(protocol.Message{Type: "hello", Protocol: protocol.Version, Mod: &protocol.ModInfo{ID: "official.echo", Version: "0.1.0"}})
		case "invoke":
			s.startInvocation(msg)
		case "cancel":
			s.cancel(msg.ID)
		case "shutdown":
			s.cancelAll()
			s.write(protocol.Message{Type: "shutdown_ack"})
			return
		}
	}
}

func (s *server) startInvocation(msg protocol.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancels[msg.ID] = cancel
	s.mu.Unlock()
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.cancels, msg.ID)
			s.mu.Unlock()
		}()
		switch msg.Action {
		case "echo":
			s.write(protocol.Message{Type: "result", ID: msg.ID, Output: msg.Input})
		case "sleep":
			var input map[string]any
			_ = json.Unmarshal(msg.Input, &input)
			ms := 1000
			if v, ok := input["ms"].(float64); ok && v >= 0 {
				ms = int(v)
			}
			s.write(protocol.Message{Type: "event", ID: msg.ID, Event: "status", Data: json.RawMessage(strconv.Quote(fmt.Sprintf("sleeping %dms", ms)))})
			t := time.NewTimer(time.Duration(ms) * time.Millisecond)
			defer t.Stop()
			select {
			case <-ctx.Done():
				s.write(protocol.Message{Type: "cancelled", ID: msg.ID})
			case <-t.C:
				s.write(protocol.Message{Type: "result", ID: msg.ID, Output: json.RawMessage(`{"slept":true}`)})
			}
		default:
			s.write(protocol.Message{Type: "error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: "ACTION_NOT_FOUND", Message: "unknown action"}})
		}
	}()
}

func (s *server) cancel(id string) {
	s.mu.Lock()
	cancel := s.cancels[id]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *server) cancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cancel := range s.cancels {
		cancel()
	}
}

var writeMu sync.Mutex

func (s *server) write(msg protocol.Message) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_ = protocol.Encode(os.Stdout, msg)
}

func has(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
