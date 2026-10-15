# Official Autonomous Agent draft

Status: **draft**

Open LLM Learner must ship an official autonomous agent implementation.

The goal is not to add an agent framework into the Core. The official agent must be built using the same public Flow, MOD, Capability, Permission, Project, and Package mechanisms available to users.

It should provide a Codex/Claude-Code-like user experience: give the agent a goal, allow it to inspect and modify its workspace through approved operations, verify its work, recover from failures, and continue until it is done, blocked, cancelled, or out of budget.

## Product requirement

An ordinary user should be able to:

1. choose a compatible model;
2. choose or create a workspace/project;
3. describe a task in natural language;
4. grant the required capabilities/permissions;
5. press Run;
6. let the agent continue autonomously;
7. interrupt, steer, pause, resume, or cancel it;
8. inspect the final result and all observable operations.

## Architecture

Conceptually:

```text
Goal
  |
  v
Task State
  |
  v
Planner / Decision Model
  |
  v
Action
  |
  +----> Capability / MOD
  |          |
  |          v
  |      Observation
  |          |
  +----------+
  |
  v
Verifier / Evaluator
  |
  +-- not done --> plan/update/retry
  |
  +-- done ------> Final Result
```

This is implemented as ordinary Flow + MOD behavior.

## Model roles

The official agent supports multiple model configurations.

### Single-model

One model handles:

- planning;
- action selection;
- synthesis;
- final response.

This is the default for simple/local setups.

### Decision + Generator

A decision/judgment model selects actions and a generative model handles free-form text/code.

Example:

```text
Local Judgment Model
  -> choose operation / next state

Text/Code Model
  -> generate patch / explanation
```

### Planner + Executor + Verifier

Advanced users may bind separate models for:

- planning;
- execution/action selection;
- verification.

The agent does not require this split.

## Task state

The agent maintains explicit task state rather than relying only on chat history.

Suggested state:

```json
{
  "goal": "...",
  "status": "running",
  "plan": [],
  "current_step": null,
  "observations": [],
  "artifacts": [],
  "attempts": {},
  "blocked_reason": null,
  "final_result": null
}
```

The exact internal schema may evolve, but state must be inspectable and serializable.

## Action protocol

The official agent uses a structured action schema.

Initial action kinds:

- `call` — invoke a Capability/MOD;
- `plan_update` — update explicit plan/task state;
- `ask_user` — request missing information or approval;
- `final` — finish successfully;
- `blocked` — stop because required progress is impossible.

Example:

```json
{
  "kind": "call",
  "capability": "filesystem.read_text",
  "input": {
    "path": "src/main.go"
  },
  "reason": "Inspect the failing implementation"
}
```

The `reason` field is an observable short justification, not hidden chain-of-thought.

## Autonomous loop

The agent continues while:

- the goal is not complete;
- it is not blocked;
- the user has not paused/cancelled;
- execution budgets remain;
- permission policy allows the next operation.

The loop always has hard bounds.

Default budget categories:

- max iterations;
- max tool/capability calls;
- wall-clock time;
- optional token budget;
- optional cost budget for remote providers.

An unlimited autonomous loop is not a default mode.

## Verification

The agent should not finish only because the model says it is done.

Verification may use:

- tests;
- assertions;
- compilation/build result;
- diff inspection;
- deterministic evaluator;
- domain-specific `evaluation.score`;
- explicitly configured model verifier.

For coding tasks, the default should prefer executable verification such as tests/builds when available.

## Failure recovery

A failed action becomes an observation.

The agent may then:

- retry with changed arguments;
- inspect additional context;
- choose another capability;
- revise the plan;
- ask the user;
- mark itself blocked.

Repeated identical failure should be detectable to avoid useless loops.

## User control

The official agent supports:

- pause;
- resume;
- cancel;
- steer/inject user instruction;
- approve/deny permission escalation;
- inspect current task state;
- inspect operation history.

A user interruption becomes an explicit observation/state update rather than silently resetting the agent.

## Permission modes

Suggested presets:

### Ask

Sensitive capabilities require approval as they are encountered.

### Workspace

Automatically use already granted capabilities inside the configured workspace boundaries.

### Dry Run

Plan and propose actions but do not perform mutating operations.

These modes configure the existing Core permission system; they are not separate security mechanisms.

## Observable history

The agent records:

- user goal/instructions;
- model-visible messages where appropriate;
- structured actions;
- capability calls;
- capability results;
- errors;
- plan/task-state changes;
- evaluator/verifier results;
- final result.

Hidden chain-of-thought is not required or stored.

## Workspace operations

The official coding/work agent should be able to bind common capabilities such as:

- filesystem.read_text;
- filesystem.write_text;
- filesystem read/write;
- command execution;
- JSON/schema operations;
- search capabilities when available;
- Git capabilities when available;
- user Script MODs.

No capability is privileged simply because it is official.

## Presets

Initial official agent presets should include:

### General Work Agent

General-purpose autonomous Flow using whichever capabilities the user attaches.

### Coding Agent

Defaults to file, command, test/build, diff, and source-control oriented capabilities.

### Local Lite Agent

Optimized for a single smaller local model:

- short explicit state;
- compact tool schema;
- bounded context;
- deterministic capability results;
- conservative loop budget.

### Decision + Generator Agent

Uses a small decision model for action selection and a text/code model for synthesis.

### WikiSkill Agent

Uses the existing WikiSkill-derived persistent knowledge / skill evolution presets on top of the autonomous agent execution loop.

## Package behavior

The official autonomous agent can itself be exported as an AI Package.

A Package may include or reference:

- model(s);
- agent Flow;
- Agent Controller MOD;
- capabilities;
- scripts;
- permission defaults;
- verifier/evaluator;
- optional memory/WikiSkill state.

This makes autonomous agents distributable products rather than only features of the Open LLM Learner application.

## v1.0 requirement

Open LLM Learner v1.0 is not complete until the official autonomous agent can perform a real multi-step task end-to-end with no manual prompt required between ordinary steps.

A coding reference task should demonstrate:

```text
goal
 -> inspect repository
 -> run tests
 -> inspect failure
 -> modify file
 -> rerun tests
 -> recover/revise if needed
 -> verify success
 -> return final result
```

The user may intervene, but ordinary progress must not require intervention.
