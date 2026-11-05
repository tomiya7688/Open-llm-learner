# LLM学習設計

Status: **draft**

LLMはOpen LLM Learnerの重要なFirst-class Model Recipeです。
ただし他のTask Modelと同じModel Builder / Training MOD基盤を利用します。

## 1. LLM Recipe

代表的な出力Capability:

- `model.text.generate`;
- optional structured generation capability.

代表Architecture:

- Decoder-only dense Transformer;
- Decoder-only MoE Transformer;
-将来の非Transformer/Hybrid Language Model.

## 2. Pretraining from scratch

必要要素:

- tokenizer / vocabulary;
- architecture config;
- corpus;
- sequence packing;
- causal language modeling objective;
- optimizer/schedule;
- checkpoint;
- evaluation.

Model Builderは簡単設定とAdvanced設定を分けます。

簡単設定では:

- target model scale;
- context target;
- dense/MoE;
- available resource.

からRecipeがresolved architectureを提案できます。

## 3. Continued Pretraining

既存Base ModelへDomain Corpusを追加学習します。

出力は新しいFull Weight Artifactまたは対応Recipeが許せばAdapter系Artifactです。

## 4. Supervised Fine-tuning / Instruction Tuning

Dataset例:

```json
{
  "instruction": "質問",
  "input": "追加Context",
  "output": "期待応答"
}
```

Chat形式DatasetもTask Mappingで同じ内部Contractへ変換できます。

## 5. Full Fine-tuning

Base Weightを更新します。

要件:

- checkpoint/resume;
- metrics;
- cancellation;
- portable final artifact;
- Resource Preflight;
- optional Partitioned Streaming mode.

## 6. LoRA / Adapter

設定:

- rank;
- alpha;
- target modules;
- dropout等のRecipe-specific options.

通常LoRAとStreamed LoRAは同じAdapter Artifact契約へ揃えます。

## 7. Quantized / Low-memory Adapter Training

Baseを低精度/低Memory表現で扱いAdapterを学習するRecipeを追加可能にします。

特定Library名をCore仕様には入れません。

## 8. Preference Optimization

Instruction-following LLM向けに:

- chosen/rejected;
- ranking;
- reward/judgment;

Datasetを利用するTraining Stageを追加可能にします。

AlgorithmはMOD側です。

## 9. MoE

Model Builderでは:

- target_total_parameters;
- target_active_parameters;

を別々に指定します。

Advancedで:

- expert count;
- top-k;
- shared experts;
- router;
- load balancing;
- expert dimensions;

等をoverrideできます。

## 10. Low-memory execution

推論:

```text
Resident
 -> CPU Offload
 -> Layer Streaming
 -> Expert Streaming
```

学習:

```text
Resident Training
 -> Partitioned Streaming Training
```

Resource Governorのbudget内でRuntime/Training MODがplanを解決します。

## 11. Evaluation

LLM Evaluationは単一指標に固定しません。

Projectは:

- held-out loss;
- task-specific evaluation;
- generation tests;
- structured tests;
- downstream Agent tests;

を組み合わせられます。

## 12. Export

LLM Training後も:

- Weight Export;
- Runtime Export;
- Package Export;

を分離します。

Safetensors等のCanonical Weightを保持し、GGUF/AirLLM等は派生Runtime Artifactとして扱います。
