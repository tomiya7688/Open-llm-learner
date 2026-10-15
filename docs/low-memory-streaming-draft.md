# Low-memory model streaming draft

Status: **draft**

Open LLM Learner should support AirLLM-style low-memory execution as a **Runtime / Training MOD strategy**, not as Core behavior.

The design is inspired by AirLLM's layer-wise model streaming and its newer expert-wise streaming for sparse MoE models.

Reference implementation/project:

- https://github.com/lyogavin/airllm

## Goals

- allow models larger than available VRAM to run locally;
- preserve the normal model capability interface;
- make low-memory execution an implementation strategy rather than a special Flow primitive;
- support both dense models and sparse MoE models;
- avoid forcing streaming when the model already fits efficiently in device memory.

## Runtime modes

A Runtime MOD may expose:

```json
{
  "memory_mode": "auto"
}
```

Initial values:

- `auto`
- `resident`
- `layer_stream`
- `expert_stream`
- `cpu_offload`

The Core does not interpret these modes. They belong to the Runtime MOD configuration.

### auto

The Runtime performs a preflight using:

- model/checkpoint size;
- free VRAM;
- free system RAM;
- disk availability;
- model architecture support;
- quantization/storage format;
- requested sequence length / runtime settings.

It then selects the least expensive mode that fits.

The preferred order is generally:

1. resident/full load when practical;
2. partial/offload mode;
3. layer streaming;
4. expert streaming for supported sparse MoE models.

Exact policy belongs to the Runtime MOD.

## Dense layer streaming

For a dense/sequential model:

1. split or index checkpoint weights by executable layer;
2. keep the current layer's weights on the execution device;
3. run the layer;
4. evict weights that are no longer needed;
5. load the next layer;
6. optionally prefetch the next layer while the current one computes.

Large checkpoints should remain on disk or mapped host storage rather than being materialized in full RAM/VRAM.

The implementation may cache prepared layer shards for later runs.

## Sparse MoE expert streaming

Where the model architecture permits independent expert loading:

1. keep non-expert/resident components available as required;
2. determine which experts are selected by routing;
3. load only selected experts;
4. execute them;
5. evict/reuse expert slots;
6. cache/prefetch likely or upcoming experts when useful.

This is preferable to loading an entire MoE layer when only a small subset of experts is active.

The Runtime must fall back to layer streaming or another safe strategy if the architecture/checkpoint format cannot support expert-level access.

## Prefetch and double buffering

Streaming is usually I/O-bound.

A Runtime may overlap:

- disk -> host loading;
- host -> device transfer;
- current-layer/expert compute.

The implementation should make prefetch optional because the ideal depth depends on storage, RAM, PCIe/device bandwidth, and model structure.

## Resident components

Some modules may need or benefit from remaining resident:

- embeddings;
- output heads;
- routers;
- small normalization/config state;
- adapters;
- vision/text towers where applicable;
- KV cache and other runtime state.

The exact resident set is Runtime-specific.

## Streamed adapter training

Low-memory training is a separate capability from inference.

For LoRA/adapter training:

- base weights remain frozen;
- frozen base layers are streamed one layer at a time;
- trainable adapters remain resident on the GPU/device;
- forward/backward scheduling must preserve what is required for gradients;
- only adapter parameters are optimized/saved.

This is suitable for an official low-VRAM LoRA path.

It is **not equivalent to native/full fine-tuning**.

## Native/full fine-tuning

Full fine-tuning changes the base weights and therefore needs a different write-back/checkpoint strategy.

Open LLM Learner should keep this as a separate implementation/research path.

Potential approaches include:

- layer-wise weight update and write-back;
- optimizer-state sharding/offload;
- activation checkpointing/offload;
- staged/layerwise fine-tuning;
- distributed or CPU/NVMe offload.

These must not be marketed as AirLLM-style streamed LoRA unless they actually implement equivalent semantics.

## Project configuration

Runtime/training config may record:

```json
{
  "memory": {
    "mode": "auto",
    "target_vram_bytes": null,
    "max_ram_bytes": null,
    "cache_dir": null,
    "prefetch": true,
    "delete_prepared_source": false
  }
}
```

Advanced implementation-specific fields remain under the selected MOD's config.

## Prepared artifacts

A Runtime may generate a prepared model representation:

```text
prepared-model/
├─ manifest.json
├─ layers/
│  └─ ...
├─ experts/
│  └─ ...
└─ metadata/
```

The exact layout is owned by the Runtime MOD and is referenced as an opaque Artifact by the Core.

Prepared artifacts should include enough source identity metadata to detect stale caches after the source model changes.

## Performance principle

Low VRAM is not automatically better.

Streaming can greatly reduce device memory while severely reducing tokens/second because of storage and transfer latency.

Therefore the UI should show separate estimates/measurements for:

- peak VRAM;
- peak RAM;
- disk requirement;
- preparation time;
- first-token latency;
- tokens/second.

The default `auto` strategy should prefer resident execution when the model comfortably fits.

## Quality guarantee

The reference project may test only hardware available to maintainers.

The architecture must permit very large models, but advertised working model/hardware combinations require measured validation.


## Distribution and packaging

Low-memory streaming is not only a training/runtime implementation detail. It may also be part of how an Open LLM Learner Package is distributed and executed.

Open LLM Learner keeps two independent distribution paths:

### Weight-only distribution

A user may export only the portable model artifact:

- native/full model weights;
- LoRA/adapter weights;
- merged weights;
- other standard/exported formats.

A weight-only artifact must not require Open LLM Learner or an AirLLM-style runtime unless the chosen external format itself has such a requirement.

The original portable weight artifact remains the canonical model artifact.

### Package distribution

A Package may additionally declare:

- bundled model weights or an external model reference;
- the Runtime MOD that should execute the model;
- Runtime capability bindings;
- memory strategy such as resident/layer_stream/expert_stream/auto;
- prepared layer/expert artifacts;
- cache/preparation metadata;
- Flow, tools, scripts, filters, and other AI-package features.

Example conceptual package configuration:

```json
{
  "model": {
    "ref": "models/main.ref.json"
  },
  "runtime": {
    "mod": "official.runtime.streaming",
    "memory": {
      "mode": "auto"
    },
    "prepared_artifact": "refs/main-streamed-runtime.ref.json"
  }
}
```

A Package may therefore be optimized for a target machine without changing the underlying model weights.

## Prepared runtime artifacts are derived artifacts

Layer shards, expert indexes, mmap layouts, caches, and other streaming-specific representations are **derived runtime artifacts**.

They are not the canonical trained weights.

This distinction is important:

```text
Canonical Weight Artifact
        |
        +----> standard external export
        |
        +----> Runtime preparation
                  |
                  +----> layer/expert prepared artifact
                            |
                            +----> Package
```

A prepared artifact should retain a source digest/reference so it can be invalidated or regenerated if the canonical weights change.

## Portable vs optimized packages

Package creation may offer:

- **Portable**: include/reference canonical weights and prepare the runtime representation on the destination machine;
- **Optimized**: bundle a prepared runtime artifact for a known target/runtime;
- **Hybrid**: include canonical weights plus optional prepared cache.

The Package Builder should expose this as a deployment choice rather than making low-memory streaming mandatory.


## Three separate AirLLM-related paths

Open LLM Learner intentionally separates three concerns.

### 1. Native Open LLM Learner Streaming Runtime

This is an official MOD implemented and maintained as part of Open LLM Learner.

Its behavior is inspired by layer/expert streaming, but its internal file layout and implementation are owned by this project.

Goals:

- stable Open LLM Learner Runtime capability contract;
- dense layer streaming;
- sparse MoE expert streaming;
- automatic memory planning;
- support for Project/Package prepared artifacts;
- freedom to change implementation without depending on AirLLM internals.

This path must not require the AirLLM Python package.

### 2. AirLLM Compatibility / Preparation

A separate compatibility MOD prepares a canonical model for use by **AirLLM itself**.

The compatibility layer should prefer invoking the pinned AirLLM implementation's own preparation/splitting path instead of reimplementing an undocumented shard layout.

AirLLM currently accepts a Hugging Face repository id or local model path and decomposes the original model layer-wise on first use. It also exposes `layer_shards_saving_path`, optional compression, prefetching, and `delete_original`.

The compatibility MOD therefore owns actions conceptually similar to:

- `airllm.prepare`
- `airllm.run`
- `airllm.inspect`

Example preparation config:

```json
{
  "source_model": "refs/model.ref.json",
  "airllm": {
    "version": "pinned-version",
    "layer_shards_path": "artifacts/airllm/main/",
    "compression": null,
    "prefetching": true,
    "delete_original": false
  }
}
```

Output:

```json
{
  "prepared_model": {
    "$ref": "artifact://airllm-prepared-model"
  },
  "source_digest": "sha256:...",
  "airllm_version": "..."
}
```

The prepared representation is a derived runtime artifact. The canonical source weights remain separately exportable.

A Package may choose either:

- the native Open LLM Learner streaming Runtime; or
- the AirLLM compatibility Runtime.

The two are not required to share an internal shard format.

### 3. Partitioned Streaming Training

Training must be able to use partitioned/streamed base-model execution as a first-class Training MOD strategy.

This is broader than inference streaming.

The training engine divides the base model into executable partitions and schedules:

```text
partition 0 load
  -> forward
  -> retain/offload required state
  -> next partition
  ...
  -> loss
  -> backward schedule
  -> partition/gradient/update handling
```

The exact backward strategy depends on the training type.

#### Adapter/LoRA mode

For LoRA or other small trainable adapters:

- base partitions stay frozen;
- base partitions stream through device memory;
- adapter parameters remain resident when practical;
- optimizer state exists only for trainable adapter parameters;
- output is the same portable adapter artifact as normal LoRA.

This is the simplest and first production target.

#### Native/full-weight mode

For native/full fine-tuning:

- base parameters themselves are trainable;
- gradients and optimizer state cannot be treated like frozen streaming weights;
- a partition must be loaded, participate in forward/backward, be updated, and persist its new state;
- optimizer/gradient state may need CPU/NVMe offload or partitioned persistence;
- correctness across partition boundaries must be validated against a normal full-resident reference run.

A possible implementation schedule is:

```text
forward partitions
  -> offload activations/checkpoints as required
  -> backward partitions in reverse
  -> load trainable partition + optimizer state
  -> apply update
  -> write updated partition back
  -> release memory
```

Alternative schedules are allowed if mathematically equivalent for the selected optimizer/training recipe.

The Training MOD must explicitly report whether a mode is:

- exact/equivalent full fine-tuning;
- staged/layer-wise approximation;
- adapter-only training.

The UI must not label an approximate staged method as ordinary Full Fine-tuning.

## AirLLM third-party boundary

AirLLM compatibility support is a third-party integration.

The Open LLM Learner native streaming implementation is independent.

An AirLLM-compatible prepared artifact should record:

- AirLLM package/version;
- source model digest;
- preparation options;
- target platform/runtime metadata when relevant.

This protects against format/behavior drift between AirLLM versions.
