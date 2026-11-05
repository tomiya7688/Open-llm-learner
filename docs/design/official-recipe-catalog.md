# Official Model Recipe Catalog

Status: **draft**

この文書は最初に公式Referenceとして用意するRecipe候補を整理します。

## 1. Tiny Encoder Classifier

Task:
- text_classification;
- decisionの一部.

Architecture:
- Transformer Encoder + Classification Head.

Training:
- scratch;
- full FT;
- head-only;
- distillation.

Target:
- CPU/local;
- low latency;
- small resident memory.

## 2. Encoder Extractor

Task:
- extraction.

Architecture:
- Transformer Encoder + token/span/field head.

Training:
- scratch;
- full FT;
- head-only;
- distillation.

## 3. Cross Encoder Ranker

Task:
- ranking.

Architecture:
- query + candidate encoder;
- scalar score head.

Training:
- pairwise;
- listwise;
- supervised score;
- distillation.

## 4. Structured Code Edit Transformer

Task:
- code_transform.

Architecture:
- compact Transformer;
- structured input;
- structured edit output.

Output example:

```json
{
  "operation": "replace_condition",
  "operator": ">=",
  "right": 80
}
```

Postprocess:
- AST/Template edit MOD.

## 5. Encoder-Decoder Transform Model

Task:
- text_transform;
- code_transform;
- template_fill.

Architecture:
- Transformer Encoder-Decoder.

Output:
- transformed sequence.

## 6. Small Decoder Generator

Task:
- short text generation;
- constrained response;
- game NPC response;
- small code generation.

Architecture:
- Decoder-only Transformer.

Training:
- scratch;
- full FT;
- LoRA/Adapter;
- distillation.

## 7. Local Judgment Model

Task:
- decision.

Output:
- bounded answer;
- probabilities/scores.

Use:
- Agent action selection;
- yes/no;
- route selection.

## 8. Embedding Model

Task:
- embedding.

Use:
- retrieval;
- similarity;
- candidate generation.

## 9. Dense LLM

Task:
- text_generation.

Architecture:
- Decoder-only dense language model.

Training:
- pretraining;
- continued pretraining;
- SFT;
- Full FT;
- LoRA;
- preference stage.

## 10. MoE LLM

Task:
- text_generation.

Architecture:
- Decoder-only MoE.

Key targets:
- total parameters;
- active parameters.

## 11. Non-Transformer Language Model

Task:
- text_generation or another supported language task.

Purpose:
- prove Recipe/Runtime abstraction does not require Transformer.

## 12. Composite Reference

Composite AIは単一Recipeではありません。

例 Game Help:

```text
Tiny Classifier
 -> Ranker
 -> Small Decoder
```

AI Package/Flowが複数Recipe成果物を組み合わせます。

## Initial QA Priority

最初の公式QAは広げすぎません。

優先:

1. Tiny Encoder Classifier or Judgment;
2. Structured Code Edit;
3. Small/Dense Decoder LLM;
4. LoRA;
5. one Runtime Export path.

この少数経路をE2Eで固めた後、Recipeを増やします。
