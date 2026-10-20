# Shell / CLI Guide

Status: **user-facing draft**

This document describes how to operate Open LLM Learner from a Shell. `ollm` is a provisional command name and may change during implementation.

The GUI is a visual control panel for the same operations.

## Project
`ollm project create my-project` — create a Project.
`ollm project validate ./my-project` — validate Project structure and configuration.

## Flow
`ollm flow list ./project` — list Flows.
`ollm flow validate ./project flows/main.flow.json` — validate a Flow.
`ollm flow run ./project --entry default` — run an entry Flow.

## MOD
`ollm mod list` — list available MODs.
`ollm mod inspect official.file` — inspect a MOD.
`ollm mod validate ./path/to/mod` — validate a MOD Package.

## Training
`ollm train ./project --entry train` — run the training Flow.
`ollm train ./project --config configs/training.config.json` — use a specific training config.

Memory examples:
`ollm train ./project --memory safe-auto`
`ollm train ./project --memory layer-stream`

Safe Auto does not target 100% RAM or VRAM utilization.

## Create an Agent
`ollm agent create ./project --preset coding` — create from the Coding Agent preset.
`ollm agent create ./project --preset local-lite` — create from the Local Lite preset.

## Run an Agent
`ollm agent run ./project --goal "Fix the failing CI tests"` — run autonomously from a goal.

Planned interactive commands:
- `/pause`
- `/resume`
- `/stop`
- `/force-stop`
- `/steer <instruction>`
- `/state`

Stop requests are accepted even while the model is generating.

## Resource Settings
`ollm run ./project --memory safe-auto` — run with Safe Auto.
`ollm run ./project --max-ram 24G --max-vram 10G` — set RAM/VRAM limits.

The Resource Governor preserves configured headroom and can request a lower-memory offload/streaming plan from compatible runtimes.

## Model Import
`ollm model import model.safetensors` — import Safetensors.
`ollm model import model.pth --safe` — safely import a PyTorch checkpoint.
`ollm model import model.onnx` — import ONNX.
`ollm model preflight model.onnx` — inspect opset, external data, and Custom Op requirements.

## Model Runtime
`ollm model run ./model-ref --runtime auto` — resolve a runtime automatically.
`ollm model run ./model-ref --runtime official.runtime.onnx` — choose a specific runtime.

## Weight Export
`ollm export weight ./project --format safetensors` — export portable weights.
`ollm export weight ./project --artifact lora` — export a LoRA artifact.

## Runtime Export
`ollm export runtime ./project --target llama.cpp --quantization Q4_K_M --out ./exports/llama` — export for llama.cpp.
`ollm export runtime ./project --target ollama` — export for Ollama.
`ollm export runtime ./project --target airllm` — create an AirLLM-prepared artifact.

Runtime Export creates target-ready files. Installing/importing them into an external application is a separate action.

## Package Export
`ollm package build ./project --profile portable` — build a Portable Package.
`ollm package build ./project --profile optimized` — build a target-optimized Package.

## Runs
`ollm run list ./project` — list run history.
`ollm run show ./project 42` — show one Run.

## Permissions
Interactive runs may request permissions as needed. Non-interactive runs can use a preconfigured permission policy. Unapproved actions are not executed automatically.

## Feature status
Shell help should identify Implemented / Experimental / Planned / Community features so planned commands are not presented as already available.