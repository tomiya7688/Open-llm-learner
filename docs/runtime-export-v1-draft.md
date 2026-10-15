# Runtime Export v1 draft

Status: **draft**

Runtime Export converts a canonical model artifact into the files, metadata, directory layout, and configuration expected by a specific external runtime/application.

Runtime Export is distinct from both Weight Export and AI Package Export.

## Three output paths

```text
Canonical Weight Artifact
├─ Weight Export
│  └─ portable/general-purpose model artifact
├─ Runtime Export
│  └─ target-application-specific derived artifact
└─ Package Export
   └─ model + runtime + Flow + tools + scripts + filters
```

The canonical model artifact remains the source of truth.

A Runtime Export is a derived artifact and may be regenerated.

## Runtime Exporter MOD

Runtime-specific behavior is provided by ordinary MODs.

Suggested capability:

- `runtime.export`

Optional capability:

- `runtime.install`

The Core does not know Ollama, LM Studio, llama.cpp, AirLLM, MLX, or any other runtime.

## Export vs install

### runtime.export

Produces a target-ready artifact without modifying the target application's installation.

Example:

```json
{
  "model": {
    "$ref": "artifact://canonical-model"
  },
  "target": {
    "runtime": "ollama",
    "profile": "default"
  },
  "options": {
    "quantization": "Q4_K_M"
  },
  "output_dir": "exports/ollama/my-model/"
}
```

Possible output:

```json
{
  "artifact": {
    "$ref": "artifact://ollama-export"
  },
  "files": [
    "model.gguf",
    "Modelfile"
  ],
  "target_runtime": "ollama",
  "target_version": null,
  "install_hint": {}
}
```

### runtime.install

Optionally installs/imports a previously exported artifact into an installed target application.

This action may require:

- `host.filesystem.write`;
- `host.process`;
- target-specific permissions.

Export should work without install.

## Export profile

A Runtime Exporter should expose profiles representing a target application's expectations.

A profile can define:

- accepted source model families/formats;
- required conversion path;
- target weight format;
- quantization choices;
- file naming rules;
- directory layout;
- metadata;
- templates/configuration files;
- adapter handling;
- split/shard handling;
- target runtime/version compatibility;
- optional install command/operation.

The profile belongs to the Exporter MOD, not the Core.

## Reusable conversion stages

Runtime exporters should be able to reuse lower-level conversion capabilities instead of duplicating model conversion logic.

Example:

```text
Canonical Safetensors
        |
        v
   merge adapter (optional)
        |
        v
   convert to GGUF
        |
        v
   quantize GGUF
        |
        +----> llama.cpp exporter
        |
        +----> LM Studio exporter
        |
        +----> Ollama exporter
```

A target exporter is responsible for the final application-specific layout/configuration.

## Initial runtime targets

### llama.cpp

Current llama.cpp model execution expects GGUF for local model files.

The exporter should be able to:

- convert supported canonical weights to GGUF;
- quantize when requested;
- preserve useful GGUF metadata;
- produce predictable filenames.

### LM Studio

Current LM Studio supports importing local GGUF models and expects a model directory convention similar to:

```text
publisher/
└─ model/
   └─ model-file.gguf
```

The exporter should create an LM Studio-ready directory/artifact.

Optional install support may use LM Studio's supported import mechanism rather than writing undocumented internal state.

### Ollama

Current Ollama can import supported Safetensors model directories or GGUF files through a `Modelfile` with `FROM`.

The exporter should be able to generate:

- supported model weights or GGUF;
- a `Modelfile`;
- optional runtime defaults supported by the selected profile.

Optional install support may invoke the supported Ollama create/import workflow.

### AirLLM

AirLLM export/preparation is handled by the AirLLM Compatibility Preparer MOD.

This is a Runtime Export target even though its derived representation differs from GGUF-oriented runtimes.

## Adapters

Runtime exporters declare how they handle adapters:

- keep adapter separate;
- merge into base model;
- convert adapter to target-specific representation;
- unsupported.

The UI must show the selected behavior before export.

## Quantization

Quantization is part of Runtime Export only when the target exporter supports it.

The user should be able to choose:

- no quantization;
- target-specific quantization preset;
- Auto/recommended where the exporter provides one.

Quantized exports are derived artifacts. Canonical weights remain unchanged.

## Version compatibility

A Runtime Export artifact should record:

- exporter MOD id/version;
- target runtime;
- target runtime version/range when relevant;
- source model digest;
- conversion steps;
- quantization;
- generated file digests when practical.

This provides enough information to detect stale or incompatible exports.

## Output location convention

Project-local runtime exports may use:

```text
exports/
└─ runtimes/
   ├─ ollama/
   ├─ lmstudio/
   ├─ llama.cpp/
   └─ airllm/
```

The user may also select an external output directory.

## UI

Weight Builder / Export UI should present:

```text
Export

Weight
  Native weights
  LoRA / Adapter
  Merged model

Runtime
  llama.cpp
  LM Studio
  Ollama
  AirLLM
  ...

Package
  Open LLM Learner Package
```

Runtime exporters are discoverable MODs, so community exporters can add additional applications without changing the Core.

## ONNX Runtime target

ONNX is a first-class Runtime Export target.

A Runtime Exporter may produce:

- `.onnx` graph;
- external tensor data files;
- ONNX Runtime configuration;
- optional declared Custom Op dependencies;
- Execution Provider recommendations.

Standard ONNX operators/functions should be preferred where possible.

When Custom Ops are required, export metadata must declare them explicitly. The Package may bundle the required native library or bind to a compatible installed provider.

An ONNX Artifact with unavailable Custom Ops must fail preflight rather than silently changing behavior.