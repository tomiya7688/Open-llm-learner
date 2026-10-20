# GUI Guide

Status: **user-facing draft**

The Open LLM Learner GUI is a visual control panel for the same operations available through the Shell/CLI. It does not use a separate execution model.

## Projects
Use New Project / Open Project to create or open a Project. Once open, the main panels are Weight Builder, Agent Builder, Flow, Models, MODs, Runs, Export, and Settings.

## Weight Builder
Select the Base Model, Dataset, training method, recipe, and memory policy, run Preflight, then start training.

The GUI may show:
- Full Fine-tuning / LoRA / Streamed LoRA
- Memory: Safe Auto
- Step / Loss / RAM / VRAM / Memory Mode
- Pause / Stop

## Agent Builder
Choose a preset, model roles, Capabilities, and permissions.

Planned presets include Coding Agent, Local Lite Agent, Decision + Generator Agent, and WikiSkill variants.

## Running an Agent
Provide a goal once and the agent can continue through multiple steps autonomously. The GUI shows Current Step, Recent Operations, RAM/VRAM usage, and Runtime Mode.

Pause, Stop, and Steer remain available while the agent is running.

### Emergency Stop
Stop prevents new actions, cancels active MOD calls, asks child processes to terminate, force-kills the process tree if necessary, and saves the latest safe Task State. It does not wait for model generation to finish.

## Flow Editor
Edit Flows as nodes and edges. Node settings include inputs, outputs, Capability binding, retry, timeout, and permission requirements. Flows are stored as JSON.

## MODs
View installed and Project-local MODs with version, provided Capabilities, permissions, dependencies, and package location.

## Models
Manage Safetensors, PyTorch .pt/.pth, ONNX, GGUF, AirLLM-prepared artifacts, and other supported formats. Canonical Weights are shown separately from runtime-derived artifacts.

## Runtime
Choose Auto or a specific Runtime. The GUI shows the resolved Runtime and Memory Mode. Options may include the official runtime, ONNX Runtime, AirLLM, and Community runtimes.

## Resources / Safe Auto
Safe Auto does not target 100% RAM or VRAM utilization. Advanced settings may expose RAM headroom, VRAM headroom, disk cache, heavy-task concurrency, priority, and memory mode.

A compatible Runtime may fall back through Resident → CPU Offload → Layer Streaming → Expert Streaming.

## Permissions
Requests such as filesystem read/write, network, and process execution are shown with choices such as Allow once, Allow for this Project, and Deny.

## Model Import
Safetensors is treated as a primary weight format. PyTorch .pt/.pth normally uses a weights-only safe import path. ONNX preflight shows opset, external data, and Custom Op dependencies.

## Export
Export is split into three products.

### Weight Export
Exports portable weights or adapters such as Safetensors, LoRA, and merged models.

### Runtime Export
Produces target-specific artifacts for llama.cpp, LM Studio, Ollama, AirLLM, ONNX Runtime, or Community exporters.

### Package Export
Builds an AI Package containing model/runtime bindings, Flow, tools, scripts, filters, and permission defaults. Planned profiles include Portable, Optimized, and Hybrid.

## WikiSkill
WikiSkill-derived presets expose Raw traces, Wiki, Skills, validation scores, and accepted/rejected skill updates.

## Runs
The Runs panel shows status, timing, Flow, models, MODs, timeline, tool calls, results, errors, retry, and verification. Agents record observable actions/results/state rather than hidden chain-of-thought.