# TaskとModel Capability

Status: **draft**

## 1. TaskとCapabilityを分ける

**Task** は「何を学習させるか」です。

**Capability** は「学習後のModelがRuntimeで何を提供するか」です。

例:

```text
Task: instruction_code_edit
        ↓ training
Capability: model.code.transform
```

複数のTraining Recipeが同じCapabilityを作っても構いません。

## 2. Initial Task families

### Text Generation

入力:
- prompt / context

出力:
- text

Capability:
- `model.text.generate`

Recipe候補:
- Decoder-only;
- Encoder-Decoder.

### Text Transform / Edit

入力:
- source text;
- instruction;
- optional structured context.

出力:
- transformed text.

Capability:
- `model.text.transform`

用途:
- rewrite;
- translation;
- code edit;
- style transform;
- template modification.

### Code Transform

入力:
- code;
- instruction;
- optional symbol/function context.

出力:
- updated code or structured edit.

Capability:
- `model.code.transform`

重要:
モデルが全文コードを生成するRecipeと、Edit Operationだけを返すRecipeの両方を許可します。

### Classification

入力:
- text/context.

出力:
- class / label.

Capability:
- `model.text.classify`

### Decision / Judgment

入力:
- text-bearing state;
- typed questions / candidates.

出力:
- boolean / choice / score / probability.

Capability:
- `model.decision.evaluate`

### Extraction

入力:
- text.

出力:
- spans / fields / structured values.

Capability:
- `model.text.extract`

### Fill / Template Completion

入力:
- template;
- variables/context;
- instruction.

出力:
- filled fields or completed text.

Capability:
- `model.text.fill`

### Ranking

入力:
- query/context;
- candidates.

出力:
- scores / ranking.

Capability:
- `model.rank`

### Embedding

入力:
- text.

出力:
- vector.

Capability:
- `model.embedding.create`

## 3. Task Definition

Task Recipeは少なくとも以下を宣言します。

```json
{
  "id": "official.task.code-transform",
  "inputs": {
    "source": "text",
    "instruction": "text"
  },
  "outputs": {
    "result": "text"
  },
  "capability": "model.code.transform",
  "dataset_contract": "official.dataset.code-transform.v1",
  "supported_recipe_families": [
    "encoder_decoder",
    "decoder_only",
    "structured_edit"
  ]
}
```

Task定義はModel Architectureを固定しません。

## 4. Dataset Contract

TaskごとにTraining Dataの意味を定義します。

Code Transform例:

```json
{
  "source": "if (value > limit) { warn(); }",
  "instruction": "条件を80以上にしてください",
  "target": "if (value >= 80) { warn(); }"
}
```

Structured Edit RecipeならTargetを:

```json
{
  "operation": "replace_condition",
  "operator": ">=",
  "right": 80
}
```

としても構いません。

## 5. Evaluation Contract

Taskは推奨Metricも宣言できます。

例:

- Classification: accuracy / F1等;
- Ranking: ranking metric;
- Extraction: field/span correctness;
- Code Transform: compile/test/result equivalence;
- Text Generation: task-specific evaluator;
- Decision: held-out correctness / calibration;
- Template Fill: exact field correctness.

Metric実装はEvaluator MODです。

## 6. User-defined Task

利用者は独自Taskを作れます。

必要なのは:

- input schema;
- output schema;
- dataset mapping;
- target Capability;
- compatible Training Recipe;
- evaluation.

Core変更は不要です。
