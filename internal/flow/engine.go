package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

type InvocationRequest struct {
	NodeID     string
	Mod        string
	Action     string
	Capability string
	Input      map[string]any
}

type InvocationEvent struct {
	Name string
	Data any
}

type Invoker interface {
	Invoke(context.Context, InvocationRequest, func(InvocationEvent)) (any, error)
}

type NodeStatus string

const (
	StatusPending   NodeStatus = "pending"
	StatusRunning   NodeStatus = "running"
	StatusSucceeded NodeStatus = "succeeded"
	StatusFailed    NodeStatus = "failed"
	StatusSkipped   NodeStatus = "skipped"
)

type NodeEvent struct {
	NodeID string
	Event  InvocationEvent
}

type Engine struct {
	Invoker Invoker
	OnEvent func(NodeEvent)
}

type Result struct {
	State       map[string]any
	NodeOutputs map[string]map[string]any
	NodeStatus  map[string]NodeStatus
	EndOutputs  map[string]map[string]any
}

type edgeRuntime struct {
	fromNode string
	fromPort string
	toNode   string
	toPort   string
	queue    []any
}

type nodeRuntime struct {
	node       Node
	runs       int
	running    bool
	repeatable bool
	status     NodeStatus
}

type completion struct {
	nodeID    string
	output    map[string]any
	setKey    string
	setValue  any
	hasSet    bool
	err       error
}

type executionEnv struct {
	flowInput map[string]any
	nodeInput map[string]any
	state     map[string]any
	outputs   map[string]map[string]any
}

func (e *Engine) Run(ctx context.Context, doc *Document, input map[string]any) (*Result, error) {
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	repeatable := repeatableNodes(doc)
	nodes := make(map[string]*nodeRuntime, len(doc.Nodes))
	order := make([]string, 0, len(doc.Nodes))
	for _, node := range doc.Nodes {
		nodes[node.ID] = &nodeRuntime{node: node, repeatable: repeatable[node.ID], status: StatusPending}
		order = append(order, node.ID)
	}

	incoming := map[string]map[string][]*edgeRuntime{}
	outgoing := map[string][]*edgeRuntime{}
	for _, edge := range doc.Edges {
		fromNode, fromPort, _ := splitEndpoint(edge.From)
		toNode, toPort, _ := splitEndpoint(edge.To)
		er := &edgeRuntime{fromNode: fromNode, fromPort: fromPort, toNode: toNode, toPort: toPort}
		if incoming[toNode] == nil {
			incoming[toNode] = map[string][]*edgeRuntime{}
		}
		incoming[toNode][toPort] = append(incoming[toNode][toPort], er)
		outgoing[fromNode] = append(outgoing[fromNode], er)
	}

	state := map[string]any{}
	outputs := map[string]map[string]any{}
	endOutputs := map[string]map[string]any{}
	completions := make(chan completion, maxInt(1, len(nodes)))
	running := 0
	var firstErr error

	for {
		started := false
		if firstErr == nil {
			for _, id := range order {
				rt := nodes[id]
				if !isRunnable(rt, nodes, incoming[id]) {
					continue
				}
				nodeInput := consumeInputs(incoming[id])
				rt.running = true
				rt.status = StatusRunning
				running++
				started = true

				env := executionEnv{
					flowInput: cloneMap(input),
					nodeInput: cloneMap(nodeInput),
					state:     cloneMap(state),
					outputs:   cloneNestedMap(outputs),
				}
				runCount := rt.runs
				go func(node Node, env executionEnv, runCount int) {
					completions <- e.executeNode(runCtx, node, env, runCount)
				}(rt.node, env, runCount)
			}
		}

		if running == 0 {
			if firstErr != nil {
				markPendingSkipped(nodes)
				return resultFrom(state, outputs, nodes, endOutputs), firstErr
			}
			select {
			case <-ctx.Done():
				markPendingSkipped(nodes)
				return resultFrom(state, outputs, nodes, endOutputs), ctx.Err()
			default:
			}
			if !started {
				markPendingSkipped(nodes)
				return resultFrom(state, outputs, nodes, endOutputs), nil
			}
		}

		select {
		case c := <-completions:
			rt := nodes[c.nodeID]
			rt.running = false
			running--
			if c.err != nil {
				rt.status = StatusFailed
				if firstErr == nil {
					firstErr = fmt.Errorf("node %s failed: %w", c.nodeID, c.err)
					cancel()
				}
				continue
			}
			rt.runs++
			rt.status = StatusSucceeded
			outputs[c.nodeID] = cloneMap(c.output)
			if c.hasSet {
				setPath(state, c.setKey, c.setValue)
			}
			if rt.node.Type == "end" {
				endOutputs[c.nodeID] = withoutControl(c.output)
			}
			for _, edge := range outgoing[c.nodeID] {
				value, ok := c.output[edge.fromPort]
				if !ok {
					if rt.node.Type == "branch" || rt.node.Type == "loop" {
						continue
					}
					if firstErr == nil {
						firstErr = fmt.Errorf("node %s did not produce output port %q required by edge to %s.%s", c.nodeID, edge.fromPort, edge.toNode, edge.toPort)
						cancel()
					}
					continue
				}
				edge.queue = append(edge.queue, value)
			}
		case <-ctx.Done():
			if firstErr == nil {
				firstErr = ctx.Err()
				cancel()
			}
		}
	}
}

func isRunnable(rt *nodeRuntime, all map[string]*nodeRuntime, groups map[string][]*edgeRuntime) bool {
	if rt.running {
		return false
	}
	if !rt.repeatable && rt.runs > 0 {
		return false
	}
	for _, dep := range rt.node.After {
		depRT := all[dep]
		if depRT == nil || depRT.runs == 0 || depRT.status == StatusFailed {
			return false
		}
	}
	if len(groups) == 0 {
		return rt.runs == 0
	}
	for _, alternatives := range groups {
		ready := false
		for _, edge := range alternatives {
			if len(edge.queue) > 0 {
				ready = true
				break
			}
		}
		if !ready {
			return false
		}
	}
	return true
}

func consumeInputs(groups map[string][]*edgeRuntime) map[string]any {
	input := map[string]any{}
	ports := make([]string, 0, len(groups))
	for port := range groups { ports = append(ports, port) }
	sort.Strings(ports)
	for _, port := range ports {
		for _, edge := range groups[port] {
			if len(edge.queue) == 0 {
				continue
			}
			input[port] = edge.queue[0]
			edge.queue = edge.queue[1:]
			break
		}
	}
	return input
}

func (e *Engine) executeNode(ctx context.Context, node Node, env executionEnv, runCount int) completion {
	c := completion{nodeID: node.ID}
	nodeCtx := ctx
	cancel := func() {}
	if node.TimeoutMS > 0 {
		nodeCtx, cancel = context.WithTimeout(ctx, time.Duration(node.TimeoutMS)*time.Millisecond)
	}
	defer cancel()

	staticInput := map[string]any{}
	for key, value := range node.Inputs {
		resolved, _, err := resolveDeep(value, env)
		if err != nil {
			c.err = err
			return c
		}
		staticInput[key] = resolved
	}
	for key, value := range env.nodeInput {
		staticInput[key] = value
	}
	env.nodeInput = staticInput

	switch node.Type {
	case "start":
		c.output = cloneMap(env.flowInput)
	case "end":
		c.output = withoutControl(staticInput)
	case "set":
		value := node.Value
		if value == nil {
			value = staticInput["value"]
		}
		resolved, _, err := resolveDeep(value, env)
		if err != nil {
			c.err = err
			return c
		}
		c.hasSet = true
		c.setKey = node.Key
		c.setValue = resolved
		c.output = map[string]any{"value": resolved}
	case "get":
		value, exists := lookupPath(env.state, node.Key)
		c.output = map[string]any{"value": value, "exists": exists}
	case "branch":
		matched, err := evalCondition(node.Condition, env)
		if err != nil {
			c.err = err
			return c
		}
		if matched {
			c.output = map[string]any{"true": true, "matched": true}
		} else {
			c.output = map[string]any{"false": true, "matched": false}
		}
	case "loop":
		done, err := evalCondition(node.Until, env)
		if err != nil {
			c.err = err
			return c
		}
		if done {
			c.output = map[string]any{"done": true, "iterations": runCount}
			return c
		}
		if runCount >= node.MaxIterations {
			c.err = fmt.Errorf("LOOP_LIMIT_EXCEEDED: max_iterations=%d", node.MaxIterations)
			return c
		}
		c.output = map[string]any{"continue": true, "iteration": runCount + 1}
	case "mod":
		if e.Invoker == nil {
			c.err = errors.New("no MOD invoker configured")
			return c
		}
		invokeInput := withoutControl(staticInput)
		request := InvocationRequest{
			NodeID: node.ID,
			Mod: node.Mod,
			Action: node.Action,
			Capability: node.Capability,
			Input: invokeInput,
		}
		var output any
		var err error
		attempts := 1
		delay := time.Duration(0)
		if node.Retry != nil {
			attempts += node.Retry.Max
			delay = time.Duration(node.Retry.DelayMS) * time.Millisecond
		}
		for attempt := 0; attempt < attempts; attempt++ {
			output, err = e.Invoker.Invoke(nodeCtx, request, func(event InvocationEvent) {
				if e.OnEvent != nil {
					e.OnEvent(NodeEvent{NodeID: node.ID, Event: event})
				}
			})
			if err == nil {
				break
			}
			if attempt+1 < attempts && delay > 0 {
				timer := time.NewTimer(delay)
				select {
				case <-nodeCtx.Done():
					timer.Stop()
					c.err = nodeCtx.Err()
					return c
				case <-timer.C:
				}
			}
		}
		if err != nil {
			c.err = err
			return c
		}
		c.output = normalizeOutput(output)
	case "flow.call":
		c.err = errors.New("flow.call execution is not implemented yet")
	default:
		c.err = fmt.Errorf("unsupported node type %q", node.Type)
	}
	if c.output == nil {
		c.output = map[string]any{}
	}
	return c
}

func normalizeOutput(value any) map[string]any {
	if value == nil {
		return map[string]any{"value": nil}
	}
	if object, ok := value.(map[string]any); ok {
		return cloneMap(object)
	}
	return map[string]any{"value": value}
}

func withoutControl(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		if key == "_control" {
			continue
		}
		out[key] = value
	}
	return out
}

func evalCondition(condition map[string]any, env executionEnv) (bool, error) {
	if len(condition) == 0 {
		return false, errors.New("condition is empty")
	}
	if raw, ok := condition["and"]; ok {
		items, ok := raw.([]any)
		if !ok {
			return false, errors.New("condition.and must be an array")
		}
		for _, item := range items {
			child, ok := item.(map[string]any)
			if !ok {
				return false, errors.New("condition.and item must be an object")
			}
			value, err := evalCondition(child, env)
			if err != nil || !value {
				return value, err
			}
		}
		return true, nil
	}
	if raw, ok := condition["or"]; ok {
		items, ok := raw.([]any)
		if !ok {
			return false, errors.New("condition.or must be an array")
		}
		for _, item := range items {
			child, ok := item.(map[string]any)
			if !ok {
				return false, errors.New("condition.or item must be an object")
			}
			value, err := evalCondition(child, env)
			if err != nil {
				return false, err
			}
			if value {
				return true, nil
			}
		}
		return false, nil
	}
	if raw, ok := condition["not"]; ok {
		child, ok := raw.(map[string]any)
		if !ok {
			return false, errors.New("condition.not must be an object")
		}
		value, err := evalCondition(child, env)
		return !value, err
	}

	op, _ := condition["op"].(string)
	if op == "" {
		return false, errors.New("condition op is required")
	}
	left, leftExists, err := resolveValue(condition["value"], env)
	if err != nil {
		return false, err
	}
	switch op {
	case "exists":
		return leftExists, nil
	case "empty":
		return !leftExists || isEmpty(left), nil
	case "truthy":
		return leftExists && isTruthy(left), nil
	}
	right, _, err := resolveValue(condition["compare"], env)
	if err != nil {
		return false, err
	}
	switch op {
	case "==":
		return valuesEqual(left, right), nil
	case "!=":
		return !valuesEqual(left, right), nil
	case ">", ">=", "<", "<=":
		cmp, err := orderedCompare(left, right)
		if err != nil {
			return false, err
		}
		switch op {
		case ">": return cmp > 0, nil
		case ">=": return cmp >= 0, nil
		case "<": return cmp < 0, nil
		default: return cmp <= 0, nil
		}
	default:
		return false, fmt.Errorf("unsupported condition operator %q", op)
	}
}

func resolveDeep(value any, env executionEnv) (any, bool, error) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			v, _, err := resolveDeep(child, env)
			if err != nil { return nil, false, err }
			out[key] = v
		}
		return out, true, nil
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			v, _, err := resolveDeep(child, env)
			if err != nil { return nil, false, err }
			out[i] = v
		}
		return out, true, nil
	default:
		return resolveValue(value, env)
	}
}

func resolveValue(value any, env executionEnv) (any, bool, error) {
	text, ok := value.(string)
	if !ok || !strings.HasPrefix(text, "$") {
		return value, true, nil
	}
	switch {
	case strings.HasPrefix(text, "$state."):
		value, exists := lookupPath(env.state, strings.TrimPrefix(text, "$state."))
		return value, exists, nil
	case strings.HasPrefix(text, "$input."):
		value, exists := lookupPath(env.nodeInput, strings.TrimPrefix(text, "$input."))
		return value, exists, nil
	case strings.HasPrefix(text, "$flow."):
		value, exists := lookupPath(env.flowInput, strings.TrimPrefix(text, "$flow."))
		return value, exists, nil
	case strings.HasPrefix(text, "$node."):
		rest := strings.TrimPrefix(text, "$node.")
		var selectedID string
		for id := range env.outputs {
			prefix := id + "."
			if strings.HasPrefix(rest, prefix) && len(id) > len(selectedID) {
				selectedID = id
			}
		}
		if selectedID == "" {
			return nil, false, nil
		}
		value, exists := lookupPath(env.outputs[selectedID], strings.TrimPrefix(rest, selectedID+"."))
		return value, exists, nil
	default:
		return nil, false, fmt.Errorf("unsupported value reference %q", text)
	}
}

func lookupPath(root map[string]any, path string) (any, bool) {
	if path == "" {
		return root, true
	}
	var current any = root
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func setPath(root map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := root
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = value
}

func valuesEqual(a, b any) bool {
	af, aok := asFloat64(a)
	bf, bok := asFloat64(b)
	if aok && bok {
		return af == bf
	}
	return reflect.DeepEqual(a, b)
}

func orderedCompare(a, b any) (int, error) {
	if af, ok := asFloat64(a); ok {
		bf, ok := asFloat64(b)
		if !ok { return 0, fmt.Errorf("cannot compare numeric value with %T", b) }
		switch {
		case af < bf: return -1, nil
		case af > bf: return 1, nil
		default: return 0, nil
		}
	}
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		switch {
		case as < bs: return -1, nil
		case as > bs: return 1, nil
		default: return 0, nil
		}
	}
	return 0, fmt.Errorf("ordered comparison requires numbers or strings, got %T and %T", a, b)
}

func asFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int: return float64(v), true
	case int8: return float64(v), true
	case int16: return float64(v), true
	case int32: return float64(v), true
	case int64: return float64(v), true
	case uint: return float64(v), true
	case uint8: return float64(v), true
	case uint16: return float64(v), true
	case uint32: return float64(v), true
	case uint64: return float64(v), true
	case float32: return float64(v), true
	case float64: return v, true
	default: return 0, false
	}
}

func isEmpty(value any) bool {
	if value == nil { return true }
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	default:
		return false
	}
}

func isTruthy(value any) bool {
	if value == nil { return false }
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	default:
		if n, ok := asFloat64(value); ok { return n != 0 }
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map:
			return rv.Len() > 0
		default:
			return true
		}
	}
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil { return map[string]any{} }
	out := make(map[string]any, len(source))
	for key, value := range source { out[key] = value }
	return out
}

func cloneNestedMap(source map[string]map[string]any) map[string]map[string]any {
	out := make(map[string]map[string]any, len(source))
	for key, value := range source { out[key] = cloneMap(value) }
	return out
}

func markPendingSkipped(nodes map[string]*nodeRuntime) {
	for _, rt := range nodes {
		if rt.status == StatusPending {
			rt.status = StatusSkipped
		}
	}
}

func resultFrom(state map[string]any, outputs map[string]map[string]any, nodes map[string]*nodeRuntime, ends map[string]map[string]any) *Result {
	status := make(map[string]NodeStatus, len(nodes))
	for id, rt := range nodes { status[id] = rt.status }
	return &Result{
		State: cloneMap(state),
		NodeOutputs: cloneNestedMap(outputs),
		NodeStatus: status,
		EndOutputs: cloneNestedMap(ends),
	}
}

func maxInt(a, b int) int {
	if a > b { return a }
	return b
}

var _ sync.Locker
