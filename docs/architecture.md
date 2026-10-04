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
