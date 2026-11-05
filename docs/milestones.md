# Milestones

## v0.1 — Core proof

v0.1 is intentionally loose. Its purpose is to prove the architecture, not to be a complete model-building product yet.

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

- real model/LLM training;
- polished GUI;
- MOD registry;
- strong hostile-code sandbox;
- remote/distributed execution;
- all planned official MODs.

## v0.x — Build toward product

During 0.x, add practical official MODs and harden the protocol without promising full compatibility until the v1 line is frozen.

Priority areas:

- Task / Dataset Contract / Model Recipe definitions;
- task-first Model Builder;
- self-contained Python Runner;
- official training MOD(s);
- at least one small specialized Transformer recipe;
- official LLM training recipes;
- file/text operations;
- command execution;
- model loading/inference;
- LoRA adapter handling;
- project UI/Flow editor;
- export adapters;
- ComfyUI adapter;
- additional script runners.

## v1.0 — End-to-end useful model and AI system

v1.0 means a user can create a task-specific model or LLM, evaluate/export it, and use models to perform real tasks.

Minimum product criteria:

### Model creation

- user starts from a Task rather than being forced to choose a CausalLM;
- Dataset is mapped to a Task-specific Dataset Contract;
- Model Builder can recommend compatible Model Recipes;
- at least one specialized non-generative/task model path is supported end-to-end;
- at least one LLM/generative model family is supported end-to-end;
- native/full fine-tuning works for a supported path;
- LoRA/Adapter training works for a supported path where appropriate;
- training from scratch or head-only/partial training is supported by at least one task recipe;
- task-specific evaluation runs before export;
- training settings have usable defaults but remain editable;
- training progress and errors are visible;
- resulting weights/adapters are saved as portable artifacts;
- user can test the trained model;
- Weight Export and Runtime Export exist for practical external use.

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


## v1.0 autonomous-agent requirement

The official autonomous agent is a mandatory v1.0 feature.

Acceptance requires a multi-step task to proceed without a new user prompt between ordinary steps:

- receive a goal;
- plan/select actions;
- inspect workspace/context;
- call capabilities;
- observe results;
- revise/retry after failures;
- verify completion;
- produce a final result.

The agent must support pause/resume/cancel/user steering and must obey project permission and execution-budget limits.

A coding/CI-fix reference task is the primary acceptance case.

See `docs/official-autonomous-agent-draft.md`.


## v1.0 model-direction acceptance

The v1.0 product must prove that Open LLM Learner is not only an LLM fine-tuning UI.

At minimum, acceptance includes:

1. one LLM/generative path;
2. one smaller specialized task-model path whose primary runtime capability is not generic free-form generation;
3. both created through the same Task -> Dataset Contract -> Recipe -> Training -> Evaluation -> Export model;
4. no Core changes required to add another task/recipe family.

Reference specialized candidates include code transformation, classification/judgment, ranking, or extraction.

See `docs/design/`.
