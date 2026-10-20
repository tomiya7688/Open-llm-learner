package modproc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tomiya7688/Open-llm-learner/internal/protocol"
)

var ErrTerminated = errors.New("MOD process terminated")

type RemoteError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *RemoteError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type CallHandler func(context.Context, protocol.CallRequest) (any, error)

type Config struct {
	Executable         string
	Args               []string
	WorkDir            string
	Env                []string
	ExpectedModID      string
	ExpectedModVersion string
	CoreVersion        string
	SupportedProtocols []string
	HandshakeTimeout   time.Duration
	CancellationGrace  time.Duration
	ShutdownGrace      time.Duration
	MaxMessageBytes    int
	Stderr             io.Writer
	CallHandler        CallHandler
}

type terminal struct {
	msg protocol.Message
	err error
}

type pendingInvocation struct {
	ctx      context.Context
	terminal chan terminal
	onEvent  func(protocol.Event)
}

type Client struct {
	cfg   Config
	cmd   *exec.Cmd
	stdin io.WriteCloser

	sendMu  sync.Mutex
	mu      sync.Mutex
	pending map[string]*pendingInvocation

	helloCh         chan protocol.Message
	shutdownAck     chan struct{}
	done            chan struct{}
	finishOnce      sync.Once
	shutdownAckOnce sync.Once
	procErr         error

	counter atomic.Uint64
}

func Start(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Executable == "" {
		return nil, errors.New("MOD executable is required")
	}
	if len(cfg.SupportedProtocols) == 0 {
		cfg.SupportedProtocols = []string{protocol.Version}
	}
	if cfg.CoreVersion == "" {
		cfg.CoreVersion = "0.1.0"
	}
	if cfg.HandshakeTimeout <= 0 {
		cfg.HandshakeTimeout = 30 * time.Second
	}
	if cfg.CancellationGrace <= 0 {
		cfg.CancellationGrace = 2 * time.Second
	}
	if cfg.ShutdownGrace <= 0 {
		cfg.ShutdownGrace = 2 * time.Second
	}

	cmd := exec.Command(cfg.Executable, cfg.Args...)
	cmd.Dir = cfg.WorkDir
	if cfg.Env != nil {
		cmd.Env = cfg.Env
	}
	configureProcess(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("MOD stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("MOD stdout: %w", err)
	}
	if cfg.Stderr != nil {
		cmd.Stderr = cfg.Stderr
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start MOD: %w", err)
	}

	c := &Client{
		cfg:         cfg,
		cmd:         cmd,
		stdin:       stdin,
		pending:     make(map[string]*pendingInvocation),
		helloCh:     make(chan protocol.Message, 1),
		shutdownAck: make(chan struct{}),
		done:        make(chan struct{}),
	}
	go c.readLoop(stdout)
	go c.waitLoop()

	if err := c.send(protocol.Message{
		Type:      "hello",
		Protocols: cfg.SupportedProtocols,
		Core:      &protocol.CoreInfo{Version: cfg.CoreVersion},
	}); err != nil {
		_ = c.ForceStop()
		return nil, err
	}

	timer := time.NewTimer(cfg.HandshakeTimeout)
	defer timer.Stop()
	select {
	case hello := <-c.helloCh:
		if hello.Protocol == "" || hello.Mod == nil {
			_ = c.ForceStop()
			return nil, errors.New("invalid MOD hello response")
		}
		if !contains(cfg.SupportedProtocols, hello.Protocol) {
			_ = c.ForceStop()
			return nil, fmt.Errorf("MOD selected unsupported protocol %q", hello.Protocol)
		}
		if cfg.ExpectedModID != "" && hello.Mod.ID != cfg.ExpectedModID {
			_ = c.ForceStop()
			return nil, fmt.Errorf("MOD id mismatch: expected %q got %q", cfg.ExpectedModID, hello.Mod.ID)
		}
		if cfg.ExpectedModVersion != "" && hello.Mod.Version != cfg.ExpectedModVersion {
			_ = c.ForceStop()
			return nil, fmt.Errorf("MOD version mismatch: expected %q got %q", cfg.ExpectedModVersion, hello.Mod.Version)
		}
		return c, nil
	case <-ctx.Done():
		_ = c.ForceStop()
		return nil, ctx.Err()
	case <-timer.C:
		_ = c.ForceStop()
		return nil, errors.New("MOD handshake timed out")
	case <-c.done:
		return nil, c.processError()
	}
}

func (c *Client) Invoke(ctx context.Context, action string, input any, onEvent func(protocol.Event)) (json.RawMessage, error) {
	if action == "" {
		return nil, errors.New("action is required")
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal invocation input: %w", err)
	}
	id := fmt.Sprintf("inv-%d", c.counter.Add(1))
	p := &pendingInvocation{ctx: ctx, terminal: make(chan terminal, 1), onEvent: onEvent}
	c.mu.Lock()
	select {
	case <-c.done:
		c.mu.Unlock()
		return nil, c.processError()
	default:
	}
	c.pending[id] = p
	c.mu.Unlock()

	if err := c.send(protocol.Message{Type: "invoke", ID: id, Action: action, Input: payload}); err != nil {
		c.removePending(id)
		return nil, err
	}

	select {
	case t := <-p.terminal:
		return terminalResult(t)
	case <-ctx.Done():
		return c.cancelAndWait(id, p, ctx.Err())
	case <-c.done:
		c.removePending(id)
		return nil, c.processError()
	}
}

func (c *Client) cancelAndWait(id string, p *pendingInvocation, cause error) (json.RawMessage, error) {
	_ = c.send(protocol.Message{Type: "cancel", ID: id})
	timer := time.NewTimer(c.cfg.CancellationGrace)
	defer timer.Stop()
	select {
	case <-p.terminal:
		if errors.Is(cause, context.DeadlineExceeded) {
			return nil, &RemoteError{Code: "TIMEOUT", Message: cause.Error()}
		}
		return nil, &RemoteError{Code: "CANCELLED", Message: cause.Error()}
	case <-timer.C:
		_ = c.ForceStop()
		if errors.Is(cause, context.DeadlineExceeded) {
			return nil, &RemoteError{Code: "TIMEOUT", Message: "invocation exceeded deadline and MOD ignored cancellation"}
		}
		return nil, &RemoteError{Code: "CANCELLED", Message: "invocation cancelled; MOD process was force-stopped"}
	case <-c.done:
		c.removePending(id)
		if errors.Is(cause, context.DeadlineExceeded) {
			return nil, &RemoteError{Code: "TIMEOUT", Message: cause.Error()}
		}
		return nil, &RemoteError{Code: "CANCELLED", Message: cause.Error()}
	}
}

func terminalResult(t terminal) (json.RawMessage, error) {
	if t.err != nil {
		return nil, t.err
	}
	switch t.msg.Type {
	case "result":
		return t.msg.Output, nil
	case "error":
		if t.msg.Error == nil {
			return nil, &RemoteError{Code: "PROTOCOL_ERROR", Message: "error message missing payload"}
		}
		return nil, &RemoteError{Code: t.msg.Error.Code, Message: t.msg.Error.Message, Details: t.msg.Error.Details}
	case "cancelled":
		return nil, &RemoteError{Code: "CANCELLED", Message: "invocation cancelled by MOD"}
	default:
		return nil, &RemoteError{Code: "PROTOCOL_ERROR", Message: "unexpected terminal message " + t.msg.Type}
	}
}

func (c *Client) Shutdown(ctx context.Context) error {
	select {
	case <-c.done:
		return nil
	default:
	}
	if err := c.send(protocol.Message{Type: "shutdown"}); err != nil {
		_ = c.ForceStop()
		return err
	}
	ackTimer := time.NewTimer(c.cfg.ShutdownGrace)
	defer ackTimer.Stop()
	select {
	case <-c.shutdownAck:
	case <-ctx.Done():
		_ = c.ForceStop()
		return ctx.Err()
	case <-ackTimer.C:
		_ = c.ForceStop()
		return errors.New("MOD shutdown acknowledgement timed out")
	case <-c.done:
		return nil
	}

	exitTimer := time.NewTimer(c.cfg.ShutdownGrace)
	defer exitTimer.Stop()
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		_ = c.ForceStop()
		return ctx.Err()
	case <-exitTimer.C:
		_ = c.ForceStop()
		return errors.New("MOD did not exit after shutdown acknowledgement")
	}
}

func (c *Client) ForceStop() error {
	return killProcessTree(c.cmd)
}

func (c *Client) readLoop(stdout io.Reader) {
	dec := protocol.NewDecoder(stdout, c.cfg.MaxMessageBytes)
	for {
		var msg protocol.Message
		if err := dec.Decode(&msg); err != nil {
			if !errors.Is(err, io.EOF) {
				c.finish(fmt.Errorf("MOD protocol read failed: %w", err))
			}
			return
		}
		switch msg.Type {
		case "hello":
			select {
			case c.helloCh <- msg:
			default:
				c.finish(errors.New("MOD sent duplicate hello"))
				_ = c.ForceStop()
				return
			}
		case "event":
			c.dispatchEvent(msg)
		case "result", "error", "cancelled":
			c.finishInvocation(msg.ID, terminal{msg: msg})
		case "call":
			go c.handleCall(msg, c.parentContext(msg.Parent))
		case "shutdown_ack":
			c.shutdownAckOnce.Do(func() { close(c.shutdownAck) })
		default:
			c.finish(fmt.Errorf("MOD sent unsupported message type %q", msg.Type))
			_ = c.ForceStop()
			return
		}
	}
}

func (c *Client) waitLoop() {
	err := c.cmd.Wait()
	if err != nil {
		c.finish(fmt.Errorf("%w: %v", ErrTerminated, err))
		return
	}
	c.finish(nil)
}

func (c *Client) handleCall(msg protocol.Message, parentCtx context.Context) {
	if msg.ID == "" || msg.Target == nil || msg.Parent == "" {
		_ = c.send(protocol.Message{Type: "call_error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: "INVALID_REQUEST", Message: "call requires id, parent, and target"}})
		return
	}
	if parentCtx == nil {
		_ = c.send(protocol.Message{Type: "call_error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: "INVALID_REQUEST", Message: "call references an unknown parent invocation"}})
		return
	}
	if c.cfg.CallHandler == nil {
		_ = c.send(protocol.Message{Type: "call_error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: "CAPABILITY_NOT_FOUND", Message: "no Core call handler is configured"}})
		return
	}
	out, err := c.cfg.CallHandler(parentCtx, protocol.CallRequest{ID: msg.ID, Parent: msg.Parent, Target: *msg.Target, Input: msg.Input})
	if err != nil {
		code := "INVALID_REQUEST"
		message := err.Error()
		var remote *RemoteError
		if errors.As(err, &remote) {
			code = remote.Code
			message = remote.Message
		}
		_ = c.send(protocol.Message{Type: "call_error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: code, Message: message}})
		return
	}
	payload, err := json.Marshal(out)
	if err != nil {
		_ = c.send(protocol.Message{Type: "call_error", ID: msg.ID, Error: &protocol.ErrorPayload{Code: "PROTOCOL_ERROR", Message: "Core call handler returned non-JSON output"}})
		return
	}
	_ = c.send(protocol.Message{Type: "call_result", ID: msg.ID, Output: payload})
}

func (c *Client) parentContext(parentID string) context.Context {
	if parentID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if p := c.pending[parentID]; p != nil {
		return p.ctx
	}
	return nil
}

func (c *Client) dispatchEvent(msg protocol.Message) {
	c.mu.Lock()
	p := c.pending[msg.ID]
	c.mu.Unlock()
	if p == nil || p.onEvent == nil {
		return
	}
	e := protocol.Event{InvocationID: msg.ID, Name: msg.Event, Data: msg.Data}
	go p.onEvent(e)
}

func (c *Client) finishInvocation(id string, t terminal) {
	c.mu.Lock()
	p := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if p != nil {
		p.terminal <- t
	}
}

func (c *Client) removePending(id string) {
	c.mu.Lock()
	delete(c.pending, id)
	c.mu.Unlock()
}

func (c *Client) send(msg protocol.Message) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	select {
	case <-c.done:
		return c.processError()
	default:
	}
	if err := protocol.Encode(c.stdin, msg); err != nil {
		return fmt.Errorf("write MOD protocol message: %w", err)
	}
	return nil
}

func (c *Client) finish(err error) {
	c.finishOnce.Do(func() {
		terminalErr := err
		if terminalErr == nil {
			terminalErr = ErrTerminated
		}
		c.mu.Lock()
		c.procErr = err
		pending := c.pending
		c.pending = make(map[string]*pendingInvocation)
		c.mu.Unlock()
		for _, p := range pending {
			p.terminal <- terminal{err: terminalErr}
		}
		close(c.done)
	})
}

func (c *Client) processError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.procErr != nil {
		return c.procErr
	}
	return ErrTerminated
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
