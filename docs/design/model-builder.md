# Model Builder

Status: **draft**

## 1. 目的

Model Builderは「データから重みを作る画面」だけではありません。

利用者のTaskから:

- Dataset形式;
- Model Recipe;
- Training Method;
- Evaluation;
- Runtime;
- Export;

を組み立てる標準UI/Flowです。

旧称Weight Builderの設計はModel Builderへ統合します。

## 2. 標準フロー

```text
1. やらせたい仕事
2. Dataset
3. Recipe
4. Model Size / Architecture
5. Training Method
6. Resource Preflight
7. Train
8. Evaluate
9. Test
10. Export
```

## 3. Step 1: Task

GUI例:

```text
何をしたいですか？

[文章を作る]
[文章を書き換える]
[種類を判定する]
[必要な情報を取り出す]
[候補から選ぶ]
[候補を並べ替える]
[似ているものを探す]
[コードを書く]
[コードを直す]
[質問に答える]
[自分で組み合わせる]
```

複数Stepは:

```text
「質問か判定し、Helpから答えを探し、おじいさん風に返す」
```

のような自然な操作文として表示し、内部ではTask Composition / Flowへ展開します。

Taskによって不要な設定は隠します。初心者向け画面ではTransformerやCRF等の方式名を最初に要求しません。

分類器を作る利用者にtemperatureやchat templateを最初から見せません。

## 4. Step 2: Dataset

DatasetをTask ContractへMappingします。

例:

```text
Column: source       -> Task input.source
Column: instruction  -> Task input.instruction
Column: result       -> Task target.result
```

不足Field、型、Label分布、Length分布等をPreflightします。

## 5. Step 3: Recipe

利用者にはまずRecommended Recipeを提示します。

コード修正例:

```text
Recommended:
  Structured Edit Transformer

Alternatives:
  Encoder-Decoder Edit Model
  Small Decoder Generator
  Existing LLM + LoRA
```

それぞれ:

- 何が得意か;
- Model size;
- Training cost;
- Runtime cost;
- 制約;
- Export compatibility;

を表示します。

## 6. Step 4: Model configuration

Recipeに応じて設定項目が変わります。

Encoder分類器なら:

- encoder size;
- pooling;
- head;
- label count.

Decoder LLMなら:

- layers;
- hidden size;
- context;
- tokenizer;
- dense/MoE.

ユーザーに共通で不要なArchitecture項目を見せないことが重要です。

## 7. Step 5: Training Method

表示候補はRecipeに応じて絞ります。

例:

- Train from scratch;
- Full Fine-tuning;
- Continued Pretraining;
- LoRA;
- Adapter/Head-only;
- Distillation;
- Partitioned Streaming variant.

すべてのTaskにLoRAを出す必要はありません。

## 8. Step 6: Resource Preflight

Resource GovernorとTraining MODが:

- RAM;
- VRAM;
- disk;
- estimated checkpoint size;
- selected memory strategy;

を解決します。

Safe Autoでは100% RAM/VRAMを目標にしません。

## 9. Evaluation

Training完了だけを成功にしません。

TaskごとのEvaluatorを実行します。

コード修正なら:

```text
Generated edit
 -> apply
 -> parse/compile
 -> tests
 -> score
```

ゲームHelpなら:

- Help retrieval correctness;
- intent correctness;
- forbidden hallucination cases;
- response format.

## 10. Export

結果は三系統:

### Weight Export
portable model weights.

### Runtime Export
Ollama / LM Studio / llama.cpp / ONNX Runtime / AirLLM等.

### Package Export
Runtime、Flow、Tools等も含めた完成AI.

## 11. Shell

GUIはModel Builder Commandを視覚化します。

将来のShell概念例:

```text
ollm model create
ollm model task set code-transform
ollm model recipe recommend
ollm model train
ollm model evaluate
ollm export ...
```

実際のCommand名はCLI設計時に固定します。


## 12. Continual Learning Capacity Policy

継続学習ではCapacity Policyを選べます。

- **Grow** — 新しいCapacityを追加しModelを大きくする。
- **Compress** — 低価値Capacityを削除しModelを小さくする。
- **Recycle** — 低価値Capacityを削除し、同じParameter Budget内でFresh Capacityへ置換する。

Recycleでは、例えば3B Modelから0.5Bを削り、その0.5Bを新しいTaskへ重点学習するFresh Capacityとして再配置できます。

Model Builderは総Parameter Budget、Recycle量、Protected Task、Focused/Consolidation Training比率を設定できる設計とします。


## 13. Modelを使わないRecommendation

Taskに対して、学習モデルが不要な場合はRule/Search/Template等を候補にします。

例:

```text
「Helpから答えを探す」
 -> Exact/Keyword Search
 -> BM25/Retrieval
 -> Embedding + Ranker
 -> LLM
```

のように、必要以上に重い方式を最初から選ばせません。

詳細は `docs/design/beginner-action-ui.md` と `docs/design/language-processing-method-catalog.md`。
