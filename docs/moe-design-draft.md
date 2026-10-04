# MoE design draft

Status: **draft**

Mixture-of-Experts support belongs to the weight-generation side of Open LLM Learner and should be implemented through training / model-construction MODs rather than as special logic in the Core.

## Goal

A user should be able to describe the intended scale of an MoE model using two independent targets:

- **target total model size**: the approximate total parameter count of the complete model;
- **target active model size**: the approximate parameter count participating in one token's forward pass.

The second value represents active compute, not necessarily the amount of model memory that must remain resident at runtime.

## Why the distinction matters

An MoE model may contain a large total number of expert parameters while routing each token through only a subset of experts.

For example, a model could target a much larger total parameter count while keeping the active parameter count closer to that of a smaller dense model.

Storage size, resident RAM/VRAM, active compute, KV cache, and actual latency are different quantities. The UI and MOD API should not collapse all of them into one ambiguous "inference size" value.

## Proposed user-facing inputs

A model-construction MOD may expose inputs such as:

```json
{
  "target_total_parameters": 30000000000,
  "target_active_parameters": 7000000000,
  "architecture": "moe",
  "advanced": {}
}
```

The default recipe may derive implementation details from those targets.

Advanced users may override details such as:

- number of experts;
- experts selected per token;
- shared expert configuration;
- expert hidden dimensions;
- router behavior;
- load-balancing strategy;
- dense vs expert layer placement;
- capacity / routing limits.

The Core does not interpret these settings. They are owned by the selected model-construction or training MOD.

## Validation

A MOD should return the resolved architecture before training begins, including at minimum:

- estimated total parameter count;
- estimated active parameter count;
- expert count;
- experts active per token;
- major architecture dimensions.

This lets the user see what the recipe actually produced from the requested targets.

## Future metrics

Later versions may also allow explicit targets for:

- expected weight storage size;
- target resident RAM / VRAM;
- target quantization;
- target tokens per second;
- target training hardware budget.

These should remain separate from active parameter count because they are not equivalent.
