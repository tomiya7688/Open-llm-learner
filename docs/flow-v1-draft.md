# Flow v1 Draft

Status: **draft**

The Flow specification is intentionally domain-neutral. Terms such as LLM, training, image generation, and agent do not belong to the Core Flow model.

The canonical Flow serialization is **JSON only**.

## 1. Core concepts

Flow v1 is built from five concepts:

- Node
- Port
- Edge
- State
- Control

## 2. Node

A Node is an executable unit in a Flow.

Two categories are currently planned:

### MOD Node

Invokes an API exposed by a MOD.

Example:

```json
{
  "id": "read_source",
  "type": "mod",
  "mod": "official.file",
  "action": "read_text"
}
```

### Core control nodes

The initial control set is intentionally small:

- start;
- end;
- branch;
- loop;
- set;
- get;
- flow.call.

Domain-specific nodes should not be added to the Core.

## 3. Port

A Node exposes input and output ports.

Example:

```json
{
  "inputs": {
    "path": { "type": "string" },
    "encoding": { "type": "string", "optional": true }
  },
  "outputs": {
    "text": { "type": "string" }
  }
}
```

### Basic values

Flow v1 should support JSON-compatible values:

- null;
- boolean;
- integer;
- number;
- string;
- array;
- object.

A reference value should also exist for large or externally managed artifacts.

Example:

```json
{
  "$ref": "artifact://example-id"
}
```

The Core transports the reference but does not need to understand whether it points to a model, image, dataset, or another resource.

## 4. Edge

An Edge connects outputs to inputs.

```json
{
  "edges": [
    {
      "from": "read_file.text",
      "to": "process.input"
    }
  ]
}
```

Data dependencies also imply execution dependencies.

A Node becomes runnable when all required inputs and explicit dependencies are ready.

For ordering without a data dependency, a Node may declare:

```json
{
  "after": ["previous_node"]
}
```

## 5. Parallel execution

No dedicated parallel node is required in v1.

If two nodes become runnable at the same time and have no dependency between them, the Core may execute them concurrently.

```text
      A
     / \
    B   C
     \ /
      D
```

After A completes, B and C may run concurrently. D runs when its dependencies are complete.

## 6. Branch

The Core should support basic conditional branching.

Initial comparison operators:

- ==;
- !=;
- >;
- >=;
- <;
- <=;
- exists;
- empty;
- truthy.

Logical composition:

- AND;
- OR;
- NOT.

Complex decision logic should be implemented by a MOD rather than by growing a large expression language inside the Core.

## 7. Loop

Loops are required for iterative workflows, including agent-like flows.

A loop should require a finite guard such as `max_iterations`.

Example:

```json
{
  "type": "loop",
  "until": {
    "value": "$state.finished",
    "op": "==",
    "compare": true
  },
  "max_iterations": 20
}
```

Infinite loops are not part of the v1 design.

Long-running or persistent services should later be modeled separately instead of pretending to be infinite Flow loops.

## 8. State

Flow v1 has run-scoped state.

State exists only to coordinate one Flow execution.

Examples:

```text
state.retry_count
state.last_result
state.finished
```

Persistent memory is not a Core responsibility and should be provided by a MOD.

## 9. Execution state

A Node invocation should have a small shared lifecycle:

- pending;
- running;
- succeeded;
- failed;
- cancelled;
- timed_out.

## 10. Progress and events

A running MOD may emit execution events such as:

- progress;
- metric;
- log;
- preview / informational events.

Example:

```json
{
  "type": "progress",
  "current": 18,
  "total": 30
}
```

In v1, events are execution telemetry and are not normal Flow data edges.

Formal node outputs are delivered through the final result.

Streaming data flows may be added in a later specification.

## 11. Errors

A failed MOD call returns a structured error.

Example:

```json
{
  "code": "OUT_OF_MEMORY",
  "message": "Not enough GPU memory",
  "details": {}
}
```

The Core transports the error without needing to understand MOD-specific error codes.

If a Flow defines an error route, execution follows it.

Otherwise, an unhandled node failure fails the Flow.

## 12. Retry

Retry belongs to Flow control.

Example:

```json
{
  "retry": {
    "max": 3,
    "delay_ms": 1000
  }
}
```

Domain-specific recovery logic belongs in a MOD or explicit Flow.

For example, automatically changing a training batch size after an out-of-memory error is not a generic Core behavior.

## 13. Timeout and cancellation

A Node may specify a timeout.

A running Flow may be cancelled.

The MOD protocol must provide a way for the Core to request cancellation of an active invocation.

## 14. Subflows

A Flow may expose inputs and outputs and be invoked from another Flow.

This allows reusable Flow components and prevents large workflows from becoming a single monolithic graph.

Example:

```text
CI Fixer
├─ Gather Context.flow
├─ Diagnose.flow
├─ Apply Fix.flow
└─ Verify.flow
```

## 15. What is intentionally not part of Flow v1

The initial version should avoid:

- realtime event-to-edge streaming;
- distributed Flow execution;
- remote workers;
- cron / scheduler behavior;
- persistent memory;
- transactions;
- a complex embedded expression language;
- unrestricted custom GUI rendering by MODs;
- dynamic self-rewriting Flow graphs;
- multi-agent-specific Core primitives.

These can be added later if real use cases require them.

## 16. Guiding boundary

**Core / Flow controls how work proceeds.**

**MODs perform the work.**

Examples:

```text
Dataset -> Training MOD -> Evaluation MOD -> Export MOD
```

and:

```text
Input -> LLM MOD -> Action -> Tool MOD -> Result -> Loop
```

are both ordinary Flows from the Core's point of view.
