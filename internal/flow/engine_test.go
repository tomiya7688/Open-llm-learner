package flow

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeInvoker struct {
	fn        func(context.Context, InvocationRequest) (any, error)
	active    atomic.Int32
	maxActive atomic.Int32
}

func (f *fakeInvoker) Invoke(ctx context.Context, req InvocationRequest, onEvent func(InvocationEvent)) (any, error) {
	active := f.active.Add(1)
	defer f.active.Add(-1)
	for {
		old := f.maxActive.Load()
		if active <= old || f.maxActive.CompareAndSwap(old, active) {
			break
		}
	}
	if onEvent != nil {
		onEvent(InvocationEvent{Name:"status", Data:"running"})
	}
	if f.fn != nil {
		return f.fn(ctx, req)
	}
	return req.Input, nil
}

func baseDocument(nodes []Node, edges []Edge) *Document {
	return &Document{
		Format: Format,
		FormatVersion: FormatVersion,
		ID: "test",
		Nodes: nodes,
		Edges: edges,
	}
}

func TestEnginePassesDataAcrossModNodes(t *testing.T) {
	invoker := &fakeInvoker{fn: func(ctx context.Context, req InvocationRequest) (any, error) {
		value, _ := asFloat64(req.Input["value"])
		return map[string]any{"value": value + 1}, nil
	}}
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"a", Type:"mod", Mod:"test.math", Action:"inc"},
		{ID:"b", Type:"mod", Mod:"test.math", Action:"inc"},
		{ID:"end", Type:"end"},
	}, []Edge{
		{From:"start.value", To:"a.value"},
		{From:"a.value", To:"b.value"},
		{From:"b.value", To:"end.value"},
	})
	result, err := (&Engine{Invoker:invoker}).Run(context.Background(), doc, map[string]any{"value":1})
	if err != nil {
		t.Fatal(err)
	}
	if got := result.EndOutputs["end"]["value"]; got != float64(3) {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestIndependentNodesRunConcurrently(t *testing.T) {
	invoker := &fakeInvoker{fn: func(ctx context.Context, req InvocationRequest) (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(60*time.Millisecond):
			return map[string]any{"value": req.Input["value"]}, nil
		}
	}}
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"a", Type:"mod", Mod:"test", Action:"sleep"},
		{ID:"b", Type:"mod", Mod:"test", Action:"sleep"},
		{ID:"endA", Type:"end"},
		{ID:"endB", Type:"end"},
	}, []Edge{
		{From:"start.value", To:"a.value"},
		{From:"start.value", To:"b.value"},
		{From:"a.value", To:"endA.value"},
		{From:"b.value", To:"endB.value"},
	})
	if _, err := (&Engine{Invoker:invoker}).Run(context.Background(), doc, map[string]any{"value":"x"}); err != nil {
		t.Fatal(err)
	}
	if invoker.maxActive.Load() < 2 {
		t.Fatalf("expected independent MOD nodes to overlap, max active=%d", invoker.maxActive.Load())
	}
}

func TestBranchUsesAlternativeEdgesOnSameInputPort(t *testing.T) {
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"flag", Type:"set", Key:"flag", Value:true},
		{ID:"branch", Type:"branch", Condition:map[string]any{
			"and":[]any{
				map[string]any{"value":"$state.flag","op":"truthy"},
				map[string]any{"not":map[string]any{"value":"$state.missing","op":"exists"}},
			},
		}},
		{ID:"yes", Type:"set", Key:"route", Value:"yes"},
		{ID:"no", Type:"set", Key:"route", Value:"no"},
		{ID:"end", Type:"end"},
	}, []Edge{
		{From:"start.go", To:"flag._control"},
		{From:"flag.value", To:"branch._control"},
		{From:"branch.true", To:"yes._control"},
		{From:"branch.false", To:"no._control"},
		{From:"yes.value", To:"end.value"},
		{From:"no.value", To:"end.value"},
	})
	result, err := (&Engine{}).Run(context.Background(), doc, map[string]any{"go":true})
	if err != nil {
		t.Fatal(err)
	}
	if result.State["route"] != "yes" {
		t.Fatalf("unexpected branch state: %#v", result.State)
	}
	if result.NodeStatus["no"] != StatusSkipped {
		t.Fatalf("expected false branch to be skipped, got %s", result.NodeStatus["no"])
	}
	if result.EndOutputs["end"]["value"] != "yes" {
		t.Fatalf("unexpected end output: %#v", result.EndOutputs)
	}
}

func TestLoopRepeatsCycleUntilCondition(t *testing.T) {
	invoker := &fakeInvoker{fn: func(ctx context.Context, req InvocationRequest) (any, error) {
		value, _ := asFloat64(req.Input["value"])
		return map[string]any{"value": value + 1}, nil
	}}
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"init", Type:"set", Key:"count", Value:0},
		{ID:"loop", Type:"loop", Until:map[string]any{"value":"$state.count","op":">=","compare":3}, MaxIterations:3},
		{ID:"get", Type:"get", Key:"count"},
		{ID:"inc", Type:"mod", Mod:"test.math", Action:"inc"},
		{ID:"save", Type:"set", Key:"count"},
		{ID:"end", Type:"end"},
	}, []Edge{
		{From:"start.go", To:"init._control"},
		{From:"init.value", To:"loop.tick"},
		{From:"loop.continue", To:"get._control"},
		{From:"get.value", To:"inc.value"},
		{From:"inc.value", To:"save.value"},
		{From:"save.value", To:"loop.tick"},
		{From:"loop.done", To:"end._control"},
	})
	result, err := (&Engine{Invoker:invoker}).Run(context.Background(), doc, map[string]any{"go":true})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := asFloat64(result.State["count"]); got != 3 {
		t.Fatalf("expected count=3, state=%#v", result.State)
	}
	if result.NodeStatus["end"] != StatusSucceeded {
		t.Fatalf("end did not run: %#v", result.NodeStatus)
	}
}

func TestLoopLimitStopsNonTerminatingCycle(t *testing.T) {
	invoker := &fakeInvoker{fn: func(ctx context.Context, req InvocationRequest) (any, error) {
		return map[string]any{"value":true}, nil
	}}
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"loop", Type:"loop", Until:map[string]any{"value":false,"op":"truthy"}, MaxIterations:2},
		{ID:"body", Type:"mod", Mod:"test", Action:"noop"},
	}, []Edge{
		{From:"start.go", To:"loop.tick"},
		{From:"loop.continue", To:"body._control"},
		{From:"body.value", To:"loop.tick"},
	})
	_, err := (&Engine{Invoker:invoker}).Run(context.Background(), doc, map[string]any{"go":true})
	if err == nil || !containsText(err.Error(), "LOOP_LIMIT_EXCEEDED") {
		t.Fatalf("expected loop limit error, got %v", err)
	}
}

func TestRetryAndTimeout(t *testing.T) {
	var attempts atomic.Int32
	invoker := &fakeInvoker{fn: func(ctx context.Context, req InvocationRequest) (any, error) {
		if attempts.Add(1) < 3 {
			return nil, errors.New("temporary")
		}
		return map[string]any{"value":"ok"}, nil
	}}
	doc := baseDocument([]Node{
		{ID:"start", Type:"start"},
		{ID:"work", Type:"mod", Mod:"test", Action:"retry", Retry:&Retry{Max:2}},
		{ID:"end", Type:"end"},
	}, []Edge{
		{From:"start.go", To:"work._control"},
		{From:"work.value", To:"end.value"},
	})
	result, err := (&Engine{Invoker:invoker}).Run(context.Background(), doc, map[string]any{"go":true})
	if err != nil {
		t.Fatal(err)
	}
	if attempts.Load() != 3 || result.EndOutputs["end"]["value"] != "ok" {
		t.Fatalf("retry failed: attempts=%d result=%#v", attempts.Load(), result.EndOutputs)
	}
}

func TestValidateRejectsUnboundedCycle(t *testing.T) {
	doc := baseDocument([]Node{
		{ID:"a", Type:"get", Key:"x"},
		{ID:"b", Type:"set", Key:"x"},
	}, []Edge{
		{From:"a.value", To:"b.value"},
		{From:"b.value", To:"a._control"},
	})
	if err := doc.Validate(); err == nil {
		t.Fatal("expected unbounded cycle validation error")
	}
}

func containsText(value, want string) bool {
	return len(value) >= len(want) && (value == want || indexText(value, want) >= 0)
}

func indexText(value, want string) int {
	for i := 0; i+len(want) <= len(value); i++ {
		if value[i:i+len(want)] == want { return i }
	}
	return -1
}

var _ sync.Locker
