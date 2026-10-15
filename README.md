# Open LLM Learner

Open LLM Learner is an open-source project for building your own models and AI systems without locking users into one training stack, runtime, or agent framework.

The application/Core is planned in **Go**. Flow files are **JSON**. Local MOD communication uses **stdio + JSON**.

The project is built around a small core that implements **flows**. Actual capabilities are supplied by **MODs**.

## Goals

Open LLM Learner aims to make it practical for individuals and organizations to:

- create model weights from their own datasets;
- perform native fine-tuning, LoRA training, and other training methods through replaceable MODs;
- export weights for use in existing runtimes and applications;
- build AI packages that combine a model with tools, scripts, prompts, and behavior;
- create anything from character LLMs to small task-specialized coding agents;
- use generative models, decision/judgment models, and non-Transformer model architectures through the same Flow/MOD foundation;
- primarily target text-to-text models while allowing secondary model capabilities where either the input or output is text;
- replace official functionality with community or private implementations when more advanced behavior is needed.

Large-model workflows should be possible by design, even when the reference project cannot quality-assure every large-scale hardware configuration.

Models whose input and output are both non-text are outside the primary project scope. Image/image, audio/audio, and other fully non-text pipelines should normally be delegated to external tools or adapters. Text-to-image, image-to-text, multimodal-to-text, and similar text-connected capabilities may be supported as secondary features.

## Core idea

The Core implements the **flow**.

It does not implement training, inference, image generation, file access, filters, or language runtimes directly.

```text
Flow
 ├─ decides what runs
 ├─ connects inputs and outputs
 ├─ controls branching / loops / state
 └─ invokes MOD APIs

MODs
 ├─ training
 ├─ inference
 ├─ file operations
 ├─ command execution
 ├─ image generation / interpretation
 ├─ filters
 ├─ external adapters
 └─ script runners
```

### MOD

A MOD exposes an API and communicates through the MOD protocol.

Official MODs have no conceptual privilege over user MODs. The official implementations are starter/reference implementations and should be replaceable.

### Runner

A Runner is a MOD that makes it easy to implement script-based MODs.

Planned official runners:

- Luau
- Python
- Ruby

The Core itself does not need to know these languages.

Official features must not require users to prepare a system Python environment. If an official MOD is implemented in Python, it should be distributed as a self-contained built package. The official Python Runner should likewise provide/manage its own runtime rather than requiring a preinstalled system Python.

## Two separate products on the same foundation

### Weight generation

Create weights that can be taken outside Open LLM Learner and used with existing runtimes, applications, or custom software.

Examples include:

- native/full fine-tuned weights;
- LoRA / adapters;
- dense and MoE model workflows;
- low-memory model execution and adapter training through AirLLM-inspired layer/expert streaming strategies;
- merged or converted model artifacts;
- runtime-oriented exports through exporter MODs.

### AI package generation

Build an AI system around a model.

An AI package may combine:

- a model or model reference;
- prompts / behavior;
- decision flows;
- standard operations;
- user-defined operations and scripts;
- optional input/output filters;
- external service adapters.

The model weights do not have to be produced by Open LLM Learner.

## Agent presets

The Agent Builder can provide research- and product-derived presets that expand into ordinary editable Projects/Flows.

Initial WikiSkill-derived presets:

- **WikiSkill Standard** — persistent Raw/Wiki/Skills layers with validation-gated skill evolution;
- **WikiSkill Local Lite** — lower-cost local-first adaptation;
- **WikiSkill Teacher to Local** — evolve skills with a stronger teacher and execute them with a smaller/local model.

See [docs/agent-presets-v1-draft.md](docs/agent-presets-v1-draft.md).

## Planned official MODs

Initial/reference MOD ideas include:

- file read / write;
- text read / write;
- JSON Schema structured output;
- binary header inspection;
- CRC checking;
- command execution;
- LoRA adapter attachment;
- image semantic interpretation;
- simple input/output content filters;
- simple SDXL image generation;
- ComfyUI adapter;
- Luau, Python, and Ruby runners.

These are intentionally reference-level capabilities. More advanced implementations can come from users, organizations, or the wider ecosystem.

## Design principle

AI requirements change quickly. The Core should therefore remain small and stable while capabilities evolve outside it.

**Core = flow. MOD = capability.**

## Status

Early design stage. Specifications are being drafted before implementation.

See [docs/architecture.md](docs/architecture.md), [docs/flow-v1-draft.md](docs/flow-v1-draft.md), [docs/mod-protocol-v1.md](docs/mod-protocol-v1.md), [docs/mod-package-v1.md](docs/mod-package-v1.md), [docs/project-format-v1.md](docs/project-format-v1.md), [docs/moe-design-draft.md](docs/moe-design-draft.md), [docs/model-capabilities-draft.md](docs/model-capabilities-draft.md), [docs/agent-presets-v1-draft.md](docs/agent-presets-v1-draft.md), [docs/low-memory-streaming-draft.md](docs/low-memory-streaming-draft.md), and [docs/milestones.md](docs/milestones.md).

## License

Open LLM Learner project-owned source code is licensed under the [Apache License 2.0](LICENSE).

Third-party models, datasets, MODs, APIs, and derived artifacts remain subject to their own licenses and terms. Open LLM Learner does not grant or guarantee rights to third-party content. See [docs/licensing.md](docs/licensing.md).
