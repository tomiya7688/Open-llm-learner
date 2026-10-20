package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	Format        = "open-llm-learner.flow"
	FormatVersion = 1
)

type Retry struct {
	Max     int `json:"max"`
	DelayMS int `json:"delay_ms,omitempty"`
}

type Node struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Mod           string         `json:"mod,omitempty"`
	Action        string         `json:"action,omitempty"`
	Capability    string         `json:"capability,omitempty"`
	Flow          string         `json:"flow,omitempty"`
	Inputs        map[string]any `json:"inputs,omitempty"`
	After         []string       `json:"after,omitempty"`
	TimeoutMS     int            `json:"timeout_ms,omitempty"`
	Retry         *Retry         `json:"retry,omitempty"`
	Condition     map[string]any `json:"condition,omitempty"`
	Until         map[string]any `json:"until,omitempty"`
	MaxIterations int            `json:"max_iterations,omitempty"`
	Key           string         `json:"key,omitempty"`
	Value         any            `json:"value,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Document struct {
	Format        string         `json:"format"`
	FormatVersion int            `json:"format_version"`
	ID            string         `json:"id"`
	Name          string         `json:"name,omitempty"`
	Inputs        map[string]any `json:"inputs,omitempty"`
	Outputs       map[string]any `json:"outputs,omitempty"`
	Nodes         []Node         `json:"nodes"`
	Edges         []Edge         `json:"edges,omitempty"`
}

func Decode(data []byte) (*Document, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse flow: %w", err)
	}
	if err := doc.Validate(); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (d *Document) Validate() error {
	if d == nil {
		return errors.New("flow is nil")
	}
	if d.Format != Format {
		return fmt.Errorf("unsupported flow format %q", d.Format)
	}
	if d.FormatVersion != FormatVersion {
		return fmt.Errorf("unsupported flow format_version %d", d.FormatVersion)
	}
	if strings.TrimSpace(d.ID) == "" {
		return errors.New("flow id is required")
	}
	if d.Nodes == nil {
		return errors.New("flow nodes is required")
	}
	nodes := make(map[string]Node, len(d.Nodes))
	for _, node := range d.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			return errors.New("flow node id is required")
		}
		if _, exists := nodes[node.ID]; exists {
			return fmt.Errorf("duplicate flow node id %q", node.ID)
		}
		switch node.Type {
		case "start", "end":
		case "mod":
			hasDirect := node.Mod != "" && node.Action != ""
			hasCapability := node.Capability != ""
			if !hasDirect && !hasCapability {
				return fmt.Errorf("mod node %q requires mod+action or capability", node.ID)
			}
			if hasDirect && hasCapability {
				return fmt.Errorf("mod node %q cannot specify both direct target and capability", node.ID)
			}
		case "branch":
			if len(node.Condition) == 0 {
				return fmt.Errorf("branch node %q requires condition", node.ID)
			}
		case "loop":
			if len(node.Until) == 0 {
				return fmt.Errorf("loop node %q requires until condition", node.ID)
			}
			if node.MaxIterations < 1 {
				return fmt.Errorf("loop node %q requires max_iterations >= 1", node.ID)
			}
		case "set":
			if strings.TrimSpace(node.Key) == "" {
				return fmt.Errorf("set node %q requires key", node.ID)
			}
		case "get":
			if strings.TrimSpace(node.Key) == "" {
				return fmt.Errorf("get node %q requires key", node.ID)
			}
		case "flow.call":
			if strings.TrimSpace(node.Flow) == "" {
				return fmt.Errorf("flow.call node %q requires flow", node.ID)
			}
		default:
			return fmt.Errorf("node %q has unsupported type %q", node.ID, node.Type)
		}
		if node.TimeoutMS < 0 {
			return fmt.Errorf("node %q timeout_ms cannot be negative", node.ID)
		}
		if node.Retry != nil && (node.Retry.Max < 0 || node.Retry.DelayMS < 0) {
			return fmt.Errorf("node %q retry values cannot be negative", node.ID)
		}
		nodes[node.ID] = node
	}
	for _, node := range d.Nodes {
		for _, dep := range node.After {
			if dep == node.ID {
				return fmt.Errorf("node %q cannot depend on itself via after", node.ID)
			}
			if _, ok := nodes[dep]; !ok {
				return fmt.Errorf("node %q after references unknown node %q", node.ID, dep)
			}
		}
	}
	for _, edge := range d.Edges {
		fromNode, fromPort, err := splitEndpoint(edge.From)
		if err != nil {
			return fmt.Errorf("edge from %q: %w", edge.From, err)
		}
		toNode, toPort, err := splitEndpoint(edge.To)
		if err != nil {
			return fmt.Errorf("edge to %q: %w", edge.To, err)
		}
		if _, ok := nodes[fromNode]; !ok {
			return fmt.Errorf("edge references unknown source node %q", fromNode)
		}
		if _, ok := nodes[toNode]; !ok {
			return fmt.Errorf("edge references unknown target node %q", toNode)
		}
		if fromPort == "" || toPort == "" {
			return fmt.Errorf("edge endpoints require non-empty ports")
		}
	}
	if err := validateCycles(d); err != nil {
		return err
	}
	return nil
}

func splitEndpoint(value string) (string, string, error) {
	i := strings.LastIndex(value, ".")
	if i <= 0 || i == len(value)-1 {
		return "", "", fmt.Errorf("endpoint must be node.port")
	}
	return value[:i], value[i+1:], nil
}

func validateCycles(d *Document) error {
	components := stronglyConnected(d)
	for _, component := range components {
		cyclic := len(component) > 1
		if len(component) == 1 {
			id := component[0]
			for _, edge := range d.Edges {
				from, _, _ := splitEndpoint(edge.From)
				to, _, _ := splitEndpoint(edge.To)
				if from == id && to == id {
					cyclic = true
					break
				}
			}
		}
		if !cyclic {
			continue
		}
		hasLoop := false
		for _, id := range component {
			for _, node := range d.Nodes {
				if node.ID == id && node.Type == "loop" && node.MaxIterations > 0 {
					hasLoop = true
					break
				}
			}
		}
		if !hasLoop {
			return fmt.Errorf("flow cycle %v has no bounded loop node", component)
		}
	}
	return nil
}

func stronglyConnected(d *Document) [][]string {
	adj := map[string][]string{}
	for _, node := range d.Nodes {
		adj[node.ID] = nil
	}
	for _, edge := range d.Edges {
		from, _, err1 := splitEndpoint(edge.From)
		to, _, err2 := splitEndpoint(edge.To)
		if err1 == nil && err2 == nil {
			adj[from] = append(adj[from], to)
		}
	}

	index := 0
	indices := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var result [][]string
	var visit func(string)
	visit = func(v string) {
		indices[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range adj[v] {
			if _, seen := indices[w]; !seen {
				visit(w)
				if low[w] < low[v] { low[v] = low[w] }
			} else if onStack[w] && indices[w] < low[v] {
				low[v] = indices[w]
			}
		}
		if low[v] == indices[v] {
			var component []string
			for {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[n] = false
				component = append(component, n)
				if n == v { break }
			}
			result = append(result, component)
		}
	}
	for id := range adj {
		if _, seen := indices[id]; !seen {
			visit(id)
		}
	}
	return result
}

func repeatableNodes(d *Document) map[string]bool {
	result := map[string]bool{}
	for _, component := range stronglyConnected(d) {
		if len(component) > 1 {
			for _, id := range component { result[id] = true }
			continue
		}
		id := component[0]
		for _, edge := range d.Edges {
			from, _, _ := splitEndpoint(edge.From)
			to, _, _ := splitEndpoint(edge.To)
			if from == id && to == id {
				result[id] = true
				break
			}
		}
	}
	return result
}
