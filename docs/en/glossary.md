# User Glossary

Status: **user-facing draft**

## Project
The saved unit of work containing or referencing models, datasets, Flows, MODs, settings, permissions, scripts, and artifacts.

## Flow
The JSON-defined execution graph describing what runs and in what order.

## Node
One operation in a Flow, such as model inference, file access, command execution, branch, or loop.

## MOD
A capability provider that performs actual work. Official and user MODs use the same mechanism.

## Capability
A semantic operation name such as filesystem.read_text, model.text.generate, or evaluation.score.

## Runner
A MOD that allows scripts such as Python, Luau, or Ruby to act as MODs.

## Weight / Canonical Weight
Model weights. Canonical Weight is the source-of-truth artifact before runtime-specific conversion.

## LoRA / Adapter
A small trainable adapter that changes model behavior without retraining every base parameter.

## Runtime
The component that loads and executes a model.

## Weight Export
Portable export of model weights or adapters.

## Runtime Export
Conversion into the files/layout expected by a target runtime or application.

## AI Package
A distributable AI containing model/runtime bindings, Flow, tools, scripts, filters, and other behavior.

## Prepared Runtime Artifact
A derived runtime-specific artifact such as GGUF, AirLLM-prepared data, or streaming shards.

## Layer Streaming
Loading model layers as needed instead of keeping the entire model in VRAM.

## Expert Streaming
Loading only selected experts for supported sparse MoE models.

## Partitioned Streaming Training
Training by loading/updating model partitions instead of keeping the whole model resident.

## MoE / Active Parameters
Mixture-of-Experts model structure and the subset of parameters active for a token; not the same as total storage or memory use.

## Model Capability
What a model can do, such as text generation, judgment, embedding, or text-connected multimodal work.

## Decision / Judgment Model
A model that returns choices, booleans, scores, or other bounded decisions instead of free-form generation.

## Agent
A goal-driven AI that selects actions, observes results, verifies progress, and continues autonomously.

## Agent Preset
A template for creating an Agent Project/Flow.

## WikiSkill
A research-derived preset family that separates raw experience, persistent Wiki knowledge, and validated executable Skills.

## Task State
Explicit Agent state such as goal, plan, current step, observations, attempts, and final result.

## Permission
Authorization for operations such as filesystem write, network, or process execution.

## Emergency Stop
A stop mechanism that propagates cancellation and can terminate unresponsive process trees.

## Resource Governor
Core-level scheduling and resource-budget control for RAM, VRAM, CPU, cache, and task priority.

## Memory: Safe Auto
Default resource policy that preserves system headroom and may request offload/streaming instead of targeting 100% memory use.

## Reference
An opaque reference used instead of embedding large artifacts directly in Flow JSON.

## Safetensors
A tensor-weight format preferred for portable canonical weights when practical.

## PyTorch .pt / .pth
PyTorch checkpoint formats supported through a safe weights-only import path by default.

## ONNX
An executable graph format containing model graph structure and weights, with explicit runtime dependencies for Custom Ops.

## GGUF
A runtime-oriented model format commonly used by llama.cpp-family runtimes.
