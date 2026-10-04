# Milestones

## v0.1 — Core proof

v0.1 is intentionally loose. Its purpose is to prove the architecture, not to be a useful LLM product yet.

Minimum success criteria:

- Go application/Core can create/open/save a project directory;
- `project.json` and `*.flow.json` load successfully;
- Core can discover a local MOD package from `mod.json`;
- Core can launch a process MOD;
- MOD Protocol handshake works over stdio + NDJSON;
- Core can invoke an action and receive events, result, error, cancellation;
- one MOD can call another MOD through the Core;
- permission declaration and user grant/deny path exists;
- basic Flow execution works:
  - start/end;
  - MOD nodes;
  - data edges;
  - branch;
  - run-scoped state;
  - basic loop with max iterations;
  - timeout/cancel;
- small JSON values and opaque large-resource references can pass between nodes;
- at least two trivial/reference MODs exist so a multi-node Flow can be demonstrated.

Not required for v0.1:

- real LLM training;
- polished GUI;
- MOD registry;
- strong hostile-code sandbox;
- remote/distributed execution;
- all planned official MODs.

## v0.x — Build toward product

During 0.x, add practical official MODs and harden the protocol without promising full compatibility until the v1 line is frozen.

Priority areas:

- self-contained Python Runner;
- official training MOD(s);
- file/text operations;
- command execution;
- model loading/inference;
- LoRA adapter handling;
- project UI/Flow editor;
- export adapters;
- ComfyUI adapter;
- additional script runners.

## v1.0 — End-to-end useful LLM system

v1.0 means a user can actually create a model and use it to perform a real task.

Minimum product criteria:

### Weight generation

- user can select a supported base model;
- user can provide training data;
- at least one reference model family is supported end-to-end;
- native/full fine-tuning works for a supported path;
- LoRA fine-tuning works for a supported path;
- training settings have usable defaults but remain editable;
- training progress and errors are visible;
- resulting weights/adapters are saved as portable artifacts;
- user can test the trained model;
- export path exists for practical external use.

### AI / task execution

- an existing model or newly trained model can be bound into an AI project;
- user input can reach the model;
- model output can drive at least one standard operation through a Flow;
- operation result can return to the model/Flow;
- final output can return to the user;
- user-defined MOD/script functionality can participate through the same public interface;
- input/output restrictions can be inserted as replaceable MODs, even if the official filters remain basic.

A CI-fix-style loop or similarly concrete task should be possible:

```text
input/context
  -> model
  -> choose action
  -> file/command operation
  -> result
  -> model
  -> final output
```

### Distribution

- normal official functionality does not require the user to manually install/configure Python;
- official Python-based components are distributed self-contained;
- projects can be saved and reopened with their selected MOD versions/settings.

### Scope statement

v1.0 does not require the project to quality-assure every large-model hardware configuration.

The architecture must allow large models and MoE workflows, while reference quality assurance may remain limited to hardware actually available to the maintainers.
