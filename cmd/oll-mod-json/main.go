package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

type pendingCall struct {
	parent string
	result chan protocol.Message
}

type server struct {
	writeMu sync.Mutex

	invokeMu sync.Mutex
	cancels  map[string]context.CancelFunc

	callMu  sync.Mutex
	pending map[string]pendingCall

	counter atomic.Uint64
}

func main() {
	s := &server{
		cancels: make(map[string]context.CancelFunc),
		pending: make(map[string]pendingCall),
	}
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
			s.write(protocol.Message{Type:"hello", Protocol:protocol.Version, Mod:&protocol.ModInfo{ID:"official.json", Version:"0.1.0"}})
		case "invoke":
			s.startInvocation(msg)
		case "call_result", "call_error":
			s.finishCall(msg)
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
	s.invokeMu.Lock()
	s.cancels[msg.ID] = cancel
	s.invokeMu.Unlock()

	go func() {
		defer func() {
			s.invokeMu.Lock()
			delete(s.cancels, msg.ID)
			s.invokeMu.Unlock()
		}()
		output, err := s.invoke(ctx, msg)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.write(protocol.Message{Type:"cancelled", ID:msg.ID})
				return
			}
			var remote *remoteFailure
			if errors.As(err, &remote) {
				s.write(protocol.Message{Type:"error", ID:msg.ID, Error:&protocol.ErrorPayload{Code:remote.code, Message:remote.message}})
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

func (s *server) invoke(ctx context.Context, msg protocol.Message) (any, error) {
	var input map[string]any
	if len(msg.Input) > 0 {
		if err := json.Unmarshal(msg.Input, &input); err != nil {
			return nil, fmt.Errorf("input must be a JSON object: %w", err)
		}
	}
	if input == nil {
		input = map[string]any{}
	}

	switch msg.Action {
	case "identity":
		return input, nil
	case "pick":
		object, ok := input["object"].(map[string]any)
		if !ok {
			return nil, errors.New("pick requires object")
		}
		key, ok := input["key"].(string)
		if !ok || key == "" {
			return nil, errors.New("pick requires non-empty key")
		}
		value, exists := object[key]
		return map[string]any{"value":value, "exists":exists}, nil
	case "merge":
		left, ok := input["left"].(map[string]any)
		if !ok {
			return nil, errors.New("merge requires left object")
		}
		right, ok := input["right"].(map[string]any)
		if !ok {
			return nil, errors.New("merge requires right object")
		}
		merged := make(map[string]any, len(left)+len(right))
		for key, value := range left {
			merged[key] = value
		}
		for key, value := range right {
			merged[key] = value
		}
		return merged, nil
	case "delegate":
		capability := "reference.echo"
		if value, ok := input["capability"].(string); ok && value != "" {
			capability = value
		}
		payload := input["payload"]
		if payload == nil {
			payload = map[string]any{}
		}
		return s.callCapability(ctx, msg.ID, capability, payload)
	default:
		return nil, &remoteFailure{code:"ACTION_NOT_FOUND", message:"unknown action"}
	}
}

func (s *server) callCapability(ctx context.Context, parentID, capability string, input any) (any, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	callID := fmt.Sprintf("call-%d", s.counter.Add(1))
	waiter := pendingCall{parent:parentID, result:make(chan protocol.Message, 1)}
	s.callMu.Lock()
	s.pending[callID] = waiter
	s.callMu.Unlock()
	defer func() {
		s.callMu.Lock()
		delete(s.pending, callID)
		s.callMu.Unlock()
	}()

	s.write(protocol.Message{
		Type:"call",
		ID:callID,
		Parent:parentID,
		Target:&protocol.CallTarget{Capability:capability},
		Input:payload,
	})
	s.write(protocol.Message{
		Type:"event",
		ID:parentID,
		Event:"status",
		Data:json.RawMessage([]byte("\"waiting for delegated capability\"")),
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-waiter.result:
		if response.Type == "call_error" {
			if response.Error == nil {
				return nil, &remoteFailure{code:"PROTOCOL_ERROR", message:"call_error missing payload"}
			}
			return nil, &remoteFailure{code:response.Error.Code, message:response.Error.Message}
		}
		var output any
		if len(response.Output) == 0 {
			return nil, nil
		}
		if err := json.Unmarshal(response.Output, &output); err != nil {
			return nil, &remoteFailure{code:"PROTOCOL_ERROR", message:"delegated result is not valid JSON"}
		}
		return output, nil
	}
}

func (s *server) finishCall(msg protocol.Message) {
	s.callMu.Lock()
	waiter, ok := s.pending[msg.ID]
	s.callMu.Unlock()
	if !ok {
		return
	}
	select {
	case waiter.result <- msg:
	default:
	}
}

func (s *server) cancel(id string) {
	s.invokeMu.Lock()
	cancel := s.cancels[id]
	s.invokeMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *server) cancelAll() {
	s.invokeMu.Lock()
	defer s.invokeMu.Unlock()
	for _, cancel := range s.cancels {
		cancel()
	}
}

func (s *server) write(msg protocol.Message) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = protocol.Encode(os.Stdout, msg)
}

type remoteFailure struct {
	code    string
	message string
}

func (e *remoteFailure) Error() string { return e.code + ": " + e.message }

func has(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
