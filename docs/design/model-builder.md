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
何を作りますか？

○ 会話・文章生成
○ コード生成
○ コード修正
○ テキスト変換
○ 分類
○ 判断・候補選択
○ 情報抽出
○ Ranking
○ Template補完
○ 自分でTaskを定義
```

Taskによって不要な設定は隠します。

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
