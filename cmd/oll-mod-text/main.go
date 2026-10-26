package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

type server struct {
	writeMu sync.Mutex
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func main() {
	s := &server{cancels:make(map[string]context.CancelFunc)}
	dec := protocol.NewDecoder(os.Stdin, protocol.DefaultMaxMessageBytes)
	for {
		var msg protocol.Message
		if err := dec.Decode(&msg); err != nil {
			return
		}
		switch msg.Type {
		case "hello":
			if !has(msg.Protocols, protocol.Version) {
				s.write(protocol.Message{Type:"error", Error:&protocol.ErrorPayload{Code:"PROTOCOL_ERROR", Message:"no compatible protocol"}})
				os.Exit(2)
			}
			s.write(protocol.Message{Type:"hello", Protocol:protocol.Version, Mod:&protocol.ModInfo{ID:"official.text", Version:"0.1.0"}})
		case "invoke":
			s.startInvocation(msg)
		case "cancel":
			s.cancel(msg.ID)
		case "shutdown":
			s.cancelAll()
			s.write(protocol.Message{Type:"shutdown_ack"})
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
		output, err := invoke(ctx, msg)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.write(protocol.Message{Type:"cancelled", ID:msg.ID})
				return
			}
			s.write(protocol.Message{Type:"error", ID:msg.ID, Error:&protocol.ErrorPayload{Code:"INVALID_REQUEST", Message:err.Error()}})
			return
		}
		payload, err := json.Marshal(output)
		if err != nil {
			s.write(protocol.Message{Type:"error", ID:msg.ID, Error:&protocol.ErrorPayload{Code:"PROTOCOL_ERROR", Message:"failed to encode result"}})
			return
		}
		s.write(protocol.Message{Type:"result", ID:msg.ID, Output:payload})
	}()
}

func invoke(ctx context.Context, msg protocol.Message) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if msg.Action != "read_text" {
		return nil, errors.New("unknown action")
	}
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(msg.Input, &input); err != nil {
		return nil, err
	}
	if input.Path == "" {
		return nil, errors.New("read_text requires path")
	}
	data, err := os.ReadFile(input.Path)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]any{
		"text":string(data),
		"bytes":len(data),
	}, nil
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

func (s *server) write(msg protocol.Message) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = protocol.Encode(os.Stdout, msg)
}

func has(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
