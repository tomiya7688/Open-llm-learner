# Model Recipe

Status: **draft**

## 1. Recipeとは

Model Recipeは:

> あるTaskを、どんなModel構造・Training Objective・Runtime Capabilityで実現するか

を定義する交換可能な設計単位です。

Core機能ではなくTraining/Runtime MOD側の契約です。

## 2. Recipeが宣言するもの

```json
{
  "id": "official.recipe.text-classifier.encoder",
  "tasks": ["text_classification"],
  "architecture_family": "transformer_encoder",
  "training_objective": "classification",
  "output_capability": "model.text.classify",
  "supports": {
    "train_from_scratch": true,
    "full_finetune": true,
    "adapter": true
  }
}
```

実際にはDataset Contract、Config Schema、Metric、Export compatibility等も持ちます。

## 3. Initial Recipe families

### Transformer Encoder Classifier

用途:
- intent;
- category;
- routing;
- moderation/filter;
- binary/multiclass decision.

出力:
- class;
- score/probability.

### Transformer Encoder Extractor

用途:
- variable extraction;
- field extraction;
- span selection;
- symbol selection.

### Cross Encoder Ranker

用途:
- query/candidate ranking;
- help entry selection;
- action candidate scoring.

### Encoder-Decoder Transformer

用途:
- rewrite;
- translation;
- code edit;
- structured text transform.

### Small Decoder Generator

用途:
- constrained generation;
- short response generation;
- local NPC dialogue;
- code snippets.

### Structured Edit Model

用途:
- code/config/template changes.

Model outputは全文ではなくEdit Operation。

例:

```json
{
  "target": "condition",
  "operator": ">=",
  "value": 80
}
```

その後の変更はdeterministic Tool/MODが行います。

### Judgment Model

用途:
- choice;
- yes/no;
- scoring;
- Agent action selection.

### Embedding Model

用途:
- retrieval;
- similarity;
- clustering support.

### Decoder-only LLM

用途:
- general generation;
- chat;
- code generation;
- broad instruction following.

### MoE LLM

用途:
-大きな総容量と一部Active計算を両立した生成モデル.

## 4. Composite Recipe

一つのAIを一つのModelに押し込む必要はありません。

ゲームHelp NPC例:

```text
Intent Classifier
      ↓
Help Ranker
      ↓
Small Response Generator
```

コード編集例:

```text
Symbol Extractor
      ↓
Edit Decision Model
      ↓
Deterministic AST Rewrite
```

これら全体はAI Package / Flowで構成します。

## 5. Recipe Recommendation

Recommendationは:

- Dataset size/shape;
- output type;
- latency requirement;
- available RAM/VRAM;
- deployment target;
- need for open-ended generation;
- deterministic constraints;

から候補を出します。

推薦結果は強制ではありません。

## 6. Architecture-neutral

Transformer Recipeを多く公式提供しても、Recipe system自体はTransformer専用にしません。

SSM、RNN、Hybrid等のRecipeも同じ契約に追加できます。
