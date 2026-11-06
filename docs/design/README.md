# 設計書

Status: **設計中**

このディレクトリはOpen LLM Learnerの製品・モデル作成・学習・実行に関する設計の中心です。

Open LLM Learnerは「LLMだけを作るアプリ」ではありません。

中心となる考え方は:

> **利用者が作りたい仕事から逆算して、必要なモデルを作る。**

LLMが必要ならLLMを学習します。
小さなTransformer、分類器、編集モデル、判断モデル、ランキングモデルで十分なら、巨大な生成LLMを使う必要はありません。

## 基本フロー

```text
やらせたい仕事
   ↓
Task
   ↓
Dataset Contract
   ↓
Model Recipe
   ↓
Training Method / Training Stages
   ↓
Evaluation
   ↓
Model Artifact
   ↓
Weight Export / Runtime Export / AI Package
```

## 1. 製品・概念設計

- [製品の方向性](product-direction.md)
- [TaskとModel Capability](model-tasks.md)
- [Standard Task Catalog](standard-task-catalog.md)
- [Official Model Recipe Catalog](official-recipe-catalog.md)
- [Contextと構造化Input](context-and-structured-input.md)
- [複数Modelの組み合わせ](model-composition.md)
- [小型用途特化モデル戦略](small-model-strategy.md)
- [反復Prune / Retrain](iterative-prune-retrain.md)
- [Model Size / Resource Budget](model-size-and-budget.md)

## 2. Model Builder

- [Model Builder](model-builder.md)
- [Model Recipe](model-recipes.md)
- [Recipe Recommendation](recipe-recommendation.md)
- [Model Size / Resource Budget](model-size-and-budget.md)
- [Model Preset / Template](presets-and-templates.md)
- [Config Resolution](config-resolution.md)
- [Model Builder Flow](model-builder-flow.md)
- [Task / Recipe Versioning](versioning-and-migration.md)

## 3. Dataset / Text Representation

- [Dataset Contract](dataset-contracts.md)
- [Dataset Pipeline](dataset-pipeline.md)
- [TokenizerとText Representation](tokenizer-and-text-representation.md)

## 4. Training

- [学習方式](training-methods.md)
- [Training Stage / Pipeline](training-stages.md)
- [LLM学習](llm-training.md)
- [Teacher / Distillation](distillation-and-teacher.md)
- [Training Stage / Pipeline](training-stages.md)
- [Teacher / Distillation](distillation-and-teacher.md)

## 5. Evaluation / Artifact / Runtime

- [Evaluation](evaluation.md)
- [Model Artifact Lifecycle](model-artifact-lifecycle.md)
- [Runtime Capability Contract](runtime-capability-contracts.md)
- [Model Artifact Lifecycle](model-artifact-lifecycle.md)
- [Testing / Compatibility](testing-and-compatibility.md)
- [Testing / Compatibility](testing-and-compatibility.md)

## 6. 代表ユースケース

- [用途特化モデルの例](specialized-model-examples.md)
- [Code Model](code-models.md)
- [ゲーム・アプリ組み込みModel](game-embedded-models.md)
- [Code Model](code-models.md)
- [ゲーム・アプリ組み込みModel](game-embedded-models.md)

## 7. 既存仕様書との関係

`docs/design/` は「何を作るか・どういう製品にするか」を扱います。

Protocolや保存形式などの機械的な仕様は `docs/` 直下に置きます。

主な関連仕様:

- `docs/architecture.md`
- `docs/project-format-v1.md`
- `docs/flow-v1-draft.md`
- `docs/mod-protocol-v1.md`
- `docs/mod-package-v1.md`
- `docs/model-capabilities-draft.md`
- `docs/model-artifact-formats-v1-draft.md`
- `docs/runtime-export-v1-draft.md`
- `docs/low-memory-streaming-draft.md`
- `docs/resource-safety-draft.md`

## 8. 設計原則

1. **Task-first**
   - Architectureより先にTaskを決める。

2. **LLMはFirst-classだが唯一ではない**
   - LLM学習を本格対応しつつ、小型専用Modelも同じくらい重要に扱う。

3. **Coreは小さく**
   - Architecture、Tokenizer、Loss、Optimizer等はMOD/Recipe側。

4. **構造化Input/Outputを使える**
   - 何でもString生成にしない。

5. **Deterministic ToolとModelを組み合わせる**
   - Modelに全部を生成させない。

6. **Portable Artifact**
   - Weightだけでも外へ持ち出せる。

7. **複数Modelを組み合わせられる**
   - 一つの巨大Modelへ全Taskを押し込まない。

8. **Resource-aware**
   - Model qualityだけでなくRAM/VRAM/Latency/Artifact sizeも設計変数。

9. **ProjectにResolved Stateを保存**
   - GUI Wizard内だけに重要設定を閉じ込めない。

10. **設計上可能と公式QA済みを区別**
    - 未検証構成を保証済みのように扱わない。
