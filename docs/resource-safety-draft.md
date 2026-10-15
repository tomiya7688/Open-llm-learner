# Resource Safety and Emergency Stop draft

Status: **draft**

Open LLM Learner is a local-first application. Two product guarantees are therefore important:

1. an autonomous Agent can be stopped easily at any time;
2. normal operation should avoid exhausting all host RAM or accelerator VRAM.

These are separate but related responsibilities.

## 1. Emergency Stop

The application UI must always expose a prominent Stop control while a Flow/Agent is running.

Stop applies to the entire active execution tree:

```text
Agent Run
  -> Flow
     -> MOD invocation
        -> command/process
           -> child processes
```

A stop request propagates from the root downward.

### Stop sequence

1. mark the Run as `cancelling`;
2. stop scheduling new Nodes/actions;
3. send protocol `cancel` to active MOD invocations;
4. ask command/process MODs to terminate child process groups;
5. wait a short configurable grace period;
6. force-terminate unresponsive MOD/process trees;
7. release runtime resources;
8. persist the latest safe Task/Run state;
9. mark the Run as `cancelled` unless a stronger terminal error occurred.

The user must not need to wait for the model to finish generating before requesting cancellation.

### Immediate UI semantics

The Stop button acknowledges locally as soon as cancellation begins.

The UI distinguishes:

- `Cancelling...`
- `Cancelled`
- `Force stopping...`

A second explicit **Force Stop** action may skip or shorten the cooperative grace period.

### Resume

Autonomous Agent state should be persisted at safe boundaries, especially:

- before an action begins;
- after an observation/result is received;
- after plan/task-state updates;
- after verification.

A cancelled task may be resumed if its active MODs and workspace state permit it.

Resume is a new execution continuation, not an attempt to resurrect a killed process.

## 2. Resource Governor

The Core owns a host-level **Resource Governor**.

Its job is not to implement model streaming. Its job is to decide whether work may start/continue under configured host resource budgets.

Runtime/Training MODs remain responsible for memory-saving implementation strategies such as:

- resident execution;
- CPU offload;
- layer streaming;
- expert streaming;
- partitioned training;
- quantization.

## 3. Resource budget

The project/application may define:

```json
{
  "resources": {
    "ram": {
      "mode": "auto",
      "max_fraction": 0.90,
      "min_free_bytes": null
    },
    "vram": {
      "mode": "auto",
      "max_fraction": 0.90,
      "min_free_bytes": null
    },
    "disk_cache": {
      "max_bytes": null
    }
  }
}
```

The exact default fractions may be tuned later.

The invariant is more important than a hard-coded number:

> Auto mode intentionally preserves headroom and does not target 100% RAM/VRAM utilization.

A user may override limits in Advanced settings.

## 4. Admission control

Before starting a memory-heavy Node, the Core asks the selected MOD for a resource estimate/preflight where supported.

The Resource Governor compares:

- current free RAM/VRAM;
- configured reserve/headroom;
- currently admitted Runs/Nodes;
- estimated peak demand;
- runtime memory strategy;
- disk/cache needs.

It may then:

- admit the Node;
- delay it until resources are free;
- reduce concurrency;
- request a lower-memory runtime plan;
- reject it with a clear resource error.

## 5. Memory pressure

During execution, the Core monitors host memory pressure.

Suggested pressure levels:

- `normal`
- `soft_limit`
- `hard_limit`
- `critical`

At a soft limit, the system may:

- stop starting new memory-heavy work;
- ask compatible runtimes to evict caches;
- lower prefetch depth;
- reduce concurrent model workers;
- switch future calls to a streaming/offload plan.

At a hard/critical limit, the system may:

- pause compatible tasks at safe boundaries;
- cancel lowest-priority work;
- force-stop a runaway child process when needed to protect the host.

## 6. Runtime memory-pressure protocol

Memory-aware MODs may expose generic actions/capabilities such as:

- `resource.estimate`
- `resource.trim`
- `resource.replan`

Example replan request:

```json
{
  "constraint": {
    "max_vram_bytes": 10737418240,
    "max_ram_bytes": 25769803776
  },
  "current_mode": "resident"
}
```

Possible response:

```json
{
  "mode": "layer_stream",
  "estimated_peak_vram_bytes": 4294967296,
  "estimated_peak_ram_bytes": 8589934592,
  "requires_prepare": true
}
```

The Core does not understand the model-specific strategy. It only enforces the declared budget and chooses among plans exposed by the MOD.

## 7. Agent-specific scheduling

Autonomous Agents must respect resource budgets across their whole task.

For example, if an Agent uses:

- a planner model;
- a generator model;
- a verifier model;

the scheduler does not assume all three may remain resident simultaneously.

The Agent may:

- reuse one model for multiple roles;
- unload one model before loading another;
- serialize model calls;
- keep a small decision model resident while streaming a larger generator;
- select a lower-memory Runtime Export/Package profile.

This is a major advantage of a local Package: the execution policy can be optimized for the actual machine.

## 8. AirLLM-like behavior

When a model does not fit within the configured headroom:

```text
resident
  -> CPU offload
  -> layer streaming
  -> expert streaming (supported sparse MoE)
```

is a valid degradation path.

This is implemented by Runtime MODs such as the native streaming runtime, while the Resource Governor supplies the constraint.

The same principle applies to training through partitioned streaming.

## 9. Hard enforcement

Cross-platform hard resource enforcement differs by operating system.

Open LLM Learner should use the strongest available mechanism where practical:

- process groups / process-tree termination;
- OS job/resource controls;
- cgroup-like controls where available;
- runtime-specific VRAM limits where supported.

A MOD that cannot be safely constrained must report that limitation.

If a process exceeds a critical safety boundary and cannot trim, the Core may terminate it.

## 10. UI

The normal UI should remain simple:

```text
Resources
  Memory: Safe Auto
  [Advanced...]

Running Agent
  RAM   21.4 / 32 GB
  VRAM   8.2 / 12 GB
  Mode   Layer Stream

  [Pause] [Stop]
```

Advanced settings may expose:

- RAM headroom;
- VRAM headroom;
- max disk cache;
- concurrent heavy tasks;
- runtime memory mode;
- prefetch depth;
- priority.

The default should favor system responsiveness over maximum benchmark utilization.
