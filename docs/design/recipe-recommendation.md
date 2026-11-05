# Recipe Recommendation

Status: **draft**

## 1. 推薦はModel Builderの中心機能

Taskを選んだ利用者に対して、Model Builderは複数のModel Recipeを比較表示します。

推薦は強制ではありません。

## 2. Recommendation Input

- Task;
- Dataset size;
- label/output type;
- average/max input length;
- available RAM/VRAM;
- latency target;
- model artifact size target;
- offline requirement;
- target Runtime;
- need for open-ended generation;
- need for deterministic output;
- existing Base Model availability.

## 3. Recommendation Output

例:

```text
1. Structured Edit Transformer
   Recommended
   Accuracy target: high
   Runtime memory: low
   Training cost: low-medium
   Output: structured edit

2. Encoder-Decoder Edit Model
   More flexible
   Runtime memory: medium
   Training cost: medium

3. Small Decoder LLM + LoRA
   Most flexible
   Runtime memory: high
   Training cost: medium-high
```

## 4. Recommendation Heuristics

初期版ではRule-basedで構いません。

例:

- bounded class output -> classifier/judgmentを優先;
- exact structured field output -> extractor/fillを優先;
- candidate listがある -> ranker/judgmentを優先;
- source textを部分修正 -> transform/structured editを優先;
- open-ended response -> generatorを候補;
- long free-form multi-domain -> LLMを優先.

将来的にRecommendation自体をModel化しても、Coreには入れません。

## 5. Resource-aware

同じQualityなら:

- smaller artifact;
- lower peak memory;
- lower latency;

を推奨できます。

## 6. Explainability

Recommendationには短い理由を表示します。

例:

> 出力が4種類の固定候補なので、生成LLMよりClassifierの方が軽量で安定します。

これはhidden reasoningではなくProduct explanationです。

## 7. User Override

利用者は:

- Recommendationを無視;
- Recipeを手動選択;
- Advanced configを上書き;

できます。

## 8. Saved Decision

選択したRecipeとRecommendation結果をProjectへ保存できます。

再Open時に同じ選択理由を確認できます。
