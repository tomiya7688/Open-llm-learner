# Model Capabilities draft

Status: **draft**

Open LLM Learner must not assume that every model is:

- a Transformer;
- autoregressive;
- a text generator;
- a chat model.

The Core and project format should work with models through declared capabilities.

## Principle

A model implementation exposes what it can do.

Examples:

- text generation;
- typed decision / judgment;
- embedding;
- ranking / scoring;
- image understanding;
- multimodal generation;
- other future model interfaces.

Architecture names such as Transformer, state-space model, recurrent model, hybrid model, or another future design are implementation metadata, not Core behavior.

## Suggested capability classes

Initial capability identifiers may include:

- `model.text.generate`
- `model.decision.evaluate`
- `model.embedding.create`
- `model.score`

These identifiers are semantic contracts, not implementation classes.

A model/runtime MOD may expose one or many capabilities.

## Text generation

A text generation capability may accept:

```json
{
  "input": "...",
  "generation": {
    "max_tokens": 256,
    "temperature": 0.7
  }
}
```

and return generated text plus optional metadata.

The exact tokenizer, decoding algorithm, architecture, and model backend remain inside the Runtime MOD.

## Decision / judgment models

A judgment model does not need to produce natural-language text.

A generic decision capability should support:

- shared state/context;
- one or more typed questions;
- bounded answer spaces;
- structured answers;
- probabilities/scores/confidence where the model provides them.

Example request:

```json
{
  "state": {
    "ci_log": "...",
    "diff": "..."
  },
  "questions": {
    "next_action": {
      "type": "choice",
      "options": [
        "inspect_file",
        "run_tests",
        "apply_patch",
        "stop"
      ]
    },
    "safe_to_apply": {
      "type": "boolean"
    }
  }
}
```

Example result:

```json
{
  "answers": {
    "next_action": {
      "value": "inspect_file",
      "probabilities": {
        "inspect_file": 0.78,
        "run_tests": 0.12,
        "apply_patch": 0.08,
        "stop": 0.02
      },
      "confidence": 0.78
    },
    "safe_to_apply": {
      "value": false,
      "probability": 0.83
    }
  }
}
```

The exact question primitives are owned by the selected model/runtime MOD. Open LLM Learner should not hard-code one vendor's decision API into the Core.

## Training

Training MODs declare what model/task types they support.

Examples:

- causal text generation;
- decision / classification;
- ranking;
- embedding;
- multimodal tasks.

The Weight Builder should choose available training recipes based on declared support rather than assuming causal language modeling.

## Non-Transformer language models

A language model is valid if a Runtime MOD can expose the required capability and a Training MOD can train or adapt it.

The project format must not require Transformer-specific fields such as:

- attention heads;
- Transformer layer count;
- Hugging Face AutoModelForCausalLM compatibility.

Those may appear inside model-specific configuration documents owned by a MOD.

This allows support for:

- state-space models;
- recurrent models;
- hybrid architectures;
- custom research architectures;
- future non-Transformer language models.

## Model references

A model reference may include descriptive metadata, but capability support should come from the Runtime/Training MOD rather than from Core-side architecture inference.

Example:

```json
{
  "format": "open-llm-learner.ref",
  "format_version": 1,
  "uri": "file:///models/example",
  "metadata": {
    "kind": "model",
    "architecture_family": "custom"
  }
}
```

## AI Package behavior

An AI Package may bind different model capabilities for different jobs.

Example:

```text
Decision model
  -> choose operation

Text model
  -> write patch/explanation

Vision model
  -> interpret screenshot
```

All three remain normal MOD capabilities connected by Flow.

This is preferred over forcing every task through a single generative LLM.
