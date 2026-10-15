# Model Artifact Formats v1 draft

Status: **draft**

Open LLM Learner supports multiple model artifact formats without treating all formats as equivalent.

## Format classes

- `weights`: portable tensor weights.
- `checkpoint`: framework-specific checkpoint/state.
- `executable_graph`: graph + weights + execution semantics.
- `runtime_prepared`: runtime-specific derived artifact.

Initial formats:

| Class | Format | Role |
| --- | --- | --- |
| weights | safetensors | preferred canonical/portable weights |
| checkpoint | pytorch .pt/.pth | compatibility import/export |
| executable_graph | ONNX | portable executable graph |
| runtime_prepared | GGUF | runtime-oriented derived artifact |
| runtime_prepared | AirLLM prepared | streaming runtime artifact |

## Safetensors

Safetensors is the preferred canonical weight container when practical.

- no arbitrary executable Python objects;
- good fit for sharded portable weights;
- suitable for base weights and adapters;
- may be lazily/memory-mapped by supporting runtimes.

Official portable weight output should prefer Safetensors where the selected model family supports it.

## PyTorch .pt / .pth

PyTorch checkpoints are supported for compatibility.

They may contain tensor/state_dict data, optimizer/checkpoint state, or pickled Python objects.

Default import policy:

1. inspect the checkpoint where possible;
2. use `torch.load(..., weights_only=True)`;
3. load to CPU by default;
4. accept tensor/state_dict-style content;
5. convert to a safer canonical representation such as Safetensors when practical;
6. reject executable/object content by default.

`weights_only=false` must never be selected automatically.

A trusted legacy import path, if provided, must be explicit and run in an isolated converter process.

Official `.pt/.pth` export should prefer state_dict/tensor-only checkpoints rather than pickled module objects.

## ONNX

ONNX is an executable graph artifact, not merely a weight container.

It can represent:

- graph structure;
- tensor initializers;
- standard operators;
- reusable Functions;
- control flow such as If, Loop and Scan;
- Custom Operator nodes;
- external tensor data.

ONNX is therefore useful for text models, decision models, preprocessing/postprocessing, and non-Transformer model graphs.

### Large ONNX models

ONNX external data is first-class.

An Artifact may contain:

```text
model.onnx
model.data
```

or multiple external-data shards. The graph and all referenced tensor files form one model Artifact.

## ONNX and custom/script behavior

ONNX must not be treated as an arbitrary self-contained script format.

Standard graph behavior should use standard operators, Functions and control flow.

When extra behavior is required, ONNX Runtime Custom Ops may be used. Their implementation is registered by the runtime and is therefore an explicit runtime dependency.

An Open LLM Learner Package may bundle:

```text
model.onnx
runtime/
  custom-ops.dll / .so / ...
```

or bind to an installed Custom Op provider.

Preferred portability order:

1. standard ONNX operators/functions;
2. widely available ONNX Runtime Extensions;
3. bundled native Custom Op library;
4. Python-backed Custom Op only when explicitly requested.

General AI workflow scripts remain Flow/Script MODs rather than being hidden inside an ONNX graph.

## ONNX Runtime Package

A Package may bind:

- ONNX model Artifact;
- ONNX Runtime MOD;
- Execution Provider;
- Custom Op libraries;
- runtime settings.

## Import and conversion

Conversions always create a new Artifact.

```text
.pth
  -> safe PyTorch importer
  -> tensor/state_dict
  -> Safetensors canonical weights

Safetensors + architecture
  -> ONNX exporter
  -> ONNX executable graph

ONNX
  -> ONNX Runtime Package
```

## Security

- do not automatically execute Python while importing checkpoints;
- treat pickle/object checkpoints as untrusted;
- validate ONNX external-data paths inside the Artifact boundary;
- inspect Custom Op domains and dependencies before execution;
- require explicit trust/permission for bundled native Custom Op libraries.

## Scope

Text-to-text remains the primary project target. ONNX support exists for language-model and text-connected workflows; it does not make Open LLM Learner a general ML studio.