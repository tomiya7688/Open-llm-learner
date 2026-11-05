# 製品の方向性

Status: **agreed direction**

## 1. Open LLM LearnerはLLM専用ではない

名前にLLMを含みますが、製品設計はLLMだけに限定しません。

主対象は **テキストを中心とした学習可能なモデル** です。

利用者が作りたいものに応じて:

- 小型Transformer;
- Encoder;
- Encoder-Decoder;
- Decoder-only生成モデル;
- Decision / Judgment Model;
- Ranking Model;
- Extraction Model;
- Edit / Transform Model;
- Embedding Model;
- 非Transformer / Hybrid Language Model;
- LLM;

を選択できる設計にします。

モデルの入力と出力の両方が非テキストである一般ML用途は、これまで通りFirst-class対象外です。

## 2. Task-first

Model Builderの最初の質問は:

> 何BのLLMを作りますか？

ではありません。

最初に尋ねるのは:

> 何をさせたいですか？

です。

例:

- 会話させたい;
- 文章を生成したい;
- コードを生成したい;
- 既存コードを指示に合わせて修正したい;
- テキストを分類したい;
- Yes/Noや候補から判断させたい;
- テキストから必要項目を抽出したい;
- 候補を順位付けしたい;
- テンプレートを条件に合わせて埋めたい;
- ゲーム内Helpにだけ答えさせたい;
- 自分でTaskを定義したい.

このTaskからDataset形式、Model Recipe、Loss、Evaluation、Runtime Capabilityを決めます。

## 3. 大きいモデルを使わないことも最適化

Open LLM Learnerは「大モデルを小さいGPUで無理に動かす」だけの製品ではありません。

重要な最適化は:

```text
その仕事に大きな生成LLMが本当に必要か？
```

を判断できることです。

例えば:

- 4択Action selection;
- 既知の関数テンプレートへの条件変更;
- ゲームHelpのIntent分類;
- 定型フォームへの値抽出;
- コマンド候補Ranking;

では、巨大な汎用LLMより小型専用モデルの方が適切な場合があります。

## 4. LLMは重要なFirst-class Recipe

LLMを軽視するわけではありません。

以下はFirst-classです:

- Decoder-only language model;
- causal language modeling;
- continued pretraining;
- instruction tuning / SFT;
- Full Fine-tuning;
- LoRA / Adapter;
- quantized / low-memory adapter training;
- MoE;
- low-memory inference;
- partitioned streaming training;
- portable/exportable weights.

ただしLLMは「すべてのTaskに使う唯一のModel」ではありません。

## 5. Transformerを使う様々なモデル

Transformerは生成LLM以外にも利用できます。

例:

### Classification

```text
Text
 -> Transformer Encoder
 -> Classification Head
 -> Class
```

### Extraction

```text
Text
 -> Encoder
 -> Token / Span Head
 -> Extracted fields
```

### Ranking

```text
Query + Candidate
 -> Encoder / Cross Encoder
 -> Score
```

### Edit / Transform

```text
Source + Instruction
 -> Encoder-Decoder / Edit Model
 -> Updated Text/Code
```

### Structured Edit

```text
Instruction + compact program context
 -> small Transformer
 -> Edit Operation
 -> deterministic AST/Template rewrite
```

この最後の構成では、モデルに最終コード全文を書かせる必要すらありません。

## 6. CoreはModel Architectureを理解しない

この方向性でもCoreは小さいままです。

Coreは:

- Flow;
- Capability;
- Artifact Reference;
- Permission;
- Resource;
- Run State;

を扱います。

Encoder、Decoder、Attention Head、Loss等はTraining/Runtime MOD側の責務です。

## 7. 作ったものは外へ持ち出せる

Model Builderの成果物はAI Packageに閉じ込めません。

```text
Model Artifact
 ├─ Weight Export
 ├─ Runtime Export
 └─ AI Package
```

小型Task ModelもLLMも同じ原則です。

## 8. 代表ユースケース

製品設計を検証する代表ケース:

1. 汎用/Character LLM;
2. 指示付きコード編集小型モデル;
3. Local Judgment Model;
4. ゲーム内Help NPC;
5. CI修正Autonomous Agent;
6. 小型Ranking / Routing Model.

これらが同じModel Builder / MOD / Flow基盤で作れることを設計目標にします。
