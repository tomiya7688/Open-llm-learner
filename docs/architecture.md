# Architecture

Status: **agreed direction**

This document captures the architectural decisions already agreed for Open LLM Learner.

## 1. Core responsibility

The Core implements **flow execution**.

The Core is responsible for:

- loading a flow definition;
- invoking MOD APIs;
- connecting node inputs and outputs;
- execution ordering;
- branching;
- loops;
- run-scoped state;
- retries, timeouts, cancellation, and error transitions;
- tracking execution status.

The Core should not contain domain-specific implementations such as model training, inference, image generation, or file manipulation.

## 2. MOD

A MOD is a capability provider that exposes an API and communicates with the Core.

A MOD may implement anything, including:

- model training;
- model inference;
- evaluation;
- export / conversion;
- file access;
- command execution;
- filters;
- image generation;
- image interpretation;
- external service integration.

The Core does not need to know the implementation language or internal architecture of a MOD.

### Official MODs are ordinary MODs

Official functionality is provided as reference / starter MODs.

Official MODs should use the same public MOD mechanism as third-party MODs. Users should be free to replace official implementations with their own.

This is intentional: requirements in the AI ecosystem change quickly, while the Core should remain small and stable.

## 3. Runner MODs

A Runner is a MOD that adapts a scripting language to the MOD interface.

Planned official runners:

- Luau;
- Python;
- Ruby.

A Runner exists for convenience. MODs are not restricted to these languages.

A developer may implement the MOD protocol directly in any language or runtime capable of exposing the required API.

## 4. Weight generation and AI package generation are separate

Open LLM Learner treats these as distinct workflows.

### Weight generation

Purpose: create portable model artifacts.

A user may only want to create weights and then use them with:

- an existing local runtime;
- another application;
- custom software;
- an external model service.

The project must not require users to build an AI package after training.

### AI package generation

Purpose: combine a model with behavior and operations.

An AI package may use:

- weights created by Open LLM Learner;
- existing local weights;
- an external model runtime;
- a remote model API.

An AI package may additionally contain or reference:

- prompts;
- flow definitions;
- tools / operations;
- user scripts;
- adapters;
- filters;
- runtime configuration.

Weights are only one part of an AI system.

## 5. Model size

The architecture must not impose a small-model-only limit.

The reference project may primarily test configurations available to its maintainers, but large models, multi-GPU setups, and other larger environments should remain possible through MOD implementations.

Quality assurance scope and architectural capability are separate concerns.

## 6. Initial use cases

Important target use cases include:

- character LLMs;
- small task-specific local models;
- coding agents specialized for narrow tasks such as CI repair;
- user-defined local AI systems;
- larger custom models where the user provides the required compute environment.

## 7. Initial official MOD direction

Reference MODs are expected to include:

### File and data

- file read;
- text read;
- file write;
- text write;
- JSON Schema structured writing;
- binary header inspection;
- CRC checking.

### Execution

- command execution.

### Models

- LoRA adapter attachment.

### Images

- simple SDXL generation;
- image semantic interpretation;
- ComfyUI adapter for more advanced image workflows.

### Filters

- simple input content filter;
- simple output content filter.

Official filters are convenience implementations, not intended to replace specialized user or third-party filtering systems.

### Script runners

- Luau;
- Python;
- Ruby.

## 8. External services

Complex capabilities should be delegable to external software and APIs rather than reimplemented inside the project.

ComfyUI is the first planned reference adapter for this approach.

The same architecture should allow future adapters for other image, audio, video, model, or automation services without changing the Core.

## 9. Architectural rule

When deciding whether functionality belongs in the Core, use this test:

> Does this functionality define how a flow is executed, or does it perform work inside that flow?

If it defines flow execution, it may belong in the Core.

If it performs work, it should normally be a MOD.


## 10. Official distribution must be self-contained

Official functionality must not require end users to manually install or configure a system Python environment.

This applies even when an official MOD is implemented in Python.

The intended distribution boundary is:

- project developers may freely use Python internally for training and other official MOD implementations;
- official Python-based MODs should be built/packaged into self-contained distributable artifacts;
- the official Python Runner should provide or manage its own Python runtime;
- the system Python installation, if any, must not be a prerequisite for normal use of official functionality;
- user-authored Python MODs should execute through the managed Runner model rather than relying on an arbitrary global Python installation.

The exact packaging technology is an implementation detail and is not fixed by the architecture.


## 11. Implementation language and wire format

The application/Core is implemented in **Go**.

The Core-to-local-MOD protocol uses **standard input / standard output with JSON messages**.

The protocol must remain language-neutral. A MOD may be implemented in any language as long as it can participate in the protocol.

Flow definitions are stored as **JSON**. YAML is not part of the Flow format.

Human readability should be achieved through a small, explicit JSON schema and good tooling rather than by introducing a second canonical serialization format.


## 12. Project file is the source of reproducibility

Open LLM Learner does not require a separate reproducibility subsystem.

The project file is the canonical saved state of a project and should contain or reference everything needed to reconstruct its intended behavior, including:

- Flow definitions;
- selected MODs and their versions;
- user configuration and resolved settings;
- model / dataset / artifact references;
- script and package references;
- relevant seeds or deterministic settings when used.

Execution logs, metrics, caches, and generated outputs are supplementary artifacts. They may help debugging or analysis, but they are not the primary source of project reproducibility.

Saving the project means saving the reproducible project state.


## 13. Artifact transport

Flow values use two broad classes:

- small values are passed directly as JSON-compatible values;
- large or externally managed data is passed by reference.

The Core must not copy large model weights, datasets, images, directories, or other large resources between nodes merely to satisfy Flow transport.

A reference is opaque to the Core. The Core transports it while the producing and consuming MODs determine how to interpret it.

The project format may store stable references to artifacts when they are part of the saved project state.

## 14. Permission model

MODs declare the capabilities / permissions they require.

The initial security model is:

1. a MOD declares requested permissions;
2. the application presents relevant permission requests to the user;
3. approved permissions are recorded in project/application state as appropriate;
4. the Core enforces the execution boundary and does not rely solely on the MOD to self-police.

The first implementation does not need to promise a perfect hostile-code sandbox. Stronger isolation mechanisms may be added later without changing the basic permission declaration model.

## 15. Project directory and JSON file organization

A project is represented as a directory with `project.json` as its entry point.

Different JSON document types should use separate directories and filename conventions rather than being mixed into one large file.

Initial convention:

```text
my-project/
├─ project.json
├─ flows/
│  ├─ train.flow.json
│  └─ agent.flow.json
├─ configs/
│  ├─ training.config.json
│  └─ runtime.config.json
├─ refs/
│  ├─ models.ref.json
│  └─ datasets.ref.json
├─ schemas/
│  └─ action.schema.json
├─ scripts/
│  └─ ...
└─ packages/
   └─ ...
```

The exact set of directories may grow as features are added. The important rule is that document type is obvious from both its location and filename suffix.

Suggested suffixes:

- `.flow.json` for Flow definitions;
- `.config.json` for configuration;
- `.ref.json` for external or artifact references;
- `.schema.json` for JSON Schema documents.

`project.json` is the canonical project manifest and references the other project documents.

Usability takes priority over inventing a highly abstract storage format. The format should remain easy to inspect, diff, copy, and edit with ordinary tools.


## 16. Models are capability providers, not necessarily text generators

Open LLM Learner must not assume a model is a Transformer, autoregressive model, or text generator.

Model architecture is an implementation detail owned by Runtime, Training, and other model-related MODs.

The Core and Flow interact with semantic capabilities such as text generation, typed decision/judgment, embedding, or scoring.

This allows judgment models, state-space/recurrent/hybrid language models, custom research architectures, and future model types to participate without changing the Core.

A single AI Package may use multiple specialized models: for example a judgment model for action selection and a generative model for writing.

See `docs/model-capabilities-draft.md`.


## 17. Model scope is text-centered

Open LLM Learner is primarily a **text-to-text model creation and AI construction tool**.

The model abstraction remains architecture-neutral, but project scope is intentionally narrower than "all machine learning models."

### Primary scope

- text -> text generation;
- text/chat -> structured decision or score;
- text-bearing multimodal input -> text;
- other model tasks closely connected to language-model workflows.

### Secondary / optional scope

Models where at least one side of the model interface is text may be supported when useful to an LLM workflow.

Examples:

- text -> image generation;
- image -> text interpretation;
- image + text -> text;
- text -> embedding/vector;
- text -> typed judgment.

These are supported as extensions, not as the main product target.

LLM-based image generation may receive lightweight/reference support when it can fit naturally into the MOD model.

### Out of scope

Models whose **input and output are both non-text** are not a project target.

Examples include:

- image -> image model pipelines;
- audio -> audio models;
- image -> audio;
- pure vision-to-vector systems with no text-facing interface;
- general-purpose non-language ML pipelines.

Such capabilities may still be reachable through external adapters or generic MODs, but Open LLM Learner does not need first-class Weight Builder, training UX, or quality guarantees for them.

This boundary prevents the project from growing into a general ML studio while keeping language-model-adjacent multimodal features possible.


## 18. Weight artifacts and Packages are independent distribution products

Open LLM Learner supports distributing a model as a **weight artifact alone** or as part of a **Package**.

A weight artifact is intentionally portable and may be used entirely outside Open LLM Learner.

A Package may bind that model to:

- a Runtime MOD;
- runtime/memory configuration;
- derived runtime-prepared artifacts;
- capability bindings;
- Flow;
- tools/operations;
- scripts;
- filters.

Runtime-specific preparation such as AirLLM-style layer/expert sharding must not replace or redefine the canonical trained weights. It is a derived execution artifact that can be regenerated from the canonical model.

This keeps model ownership/export independent from Open LLM Learner while still allowing highly optimized self-contained Packages.
