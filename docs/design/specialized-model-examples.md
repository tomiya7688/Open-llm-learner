# 用途特化モデルの設計例

Status: **draft**

この文書の例は「将来できたらよい機能」ではなく、Model Builder設計が満たすべき代表ケースです。

## 1. 関数プリセットを条件に合わせて変更するモデル

### 利用者体験

事前に:

```text
「上限を超えたら警告する関数」
```

というFunction Presetを登録しておきます。

利用者:

```text
「上限を超えたら警告する関数」で
条件を80以上にしてください
```

### Recipe A: Text-to-code edit

```text
Preset Code + Instruction
        ↓
Small Encoder-Decoder
        ↓
Updated Code
```

### Recipe B: Structured edit

より軽い構成:

```text
Function metadata
+ relevant variables
+ current condition
+ instruction
        ↓
Small Transformer
        ↓
{
  "operation": "replace_condition",
  "operator": ">=",
  "right": 80
}
        ↓
AST Rewrite MOD
        ↓
Updated Code
```

この場合、ModelはProgram全体を生成する必要がありません。

### Dataset

```json
{
  "preset": "limit_warning",
  "context": {
    "variables": ["value", "limit"],
    "condition": "value > limit"
  },
  "instruction": "80以上なら警告してください",
  "target": {
    "operation": "replace_condition",
    "operator": ">=",
    "right": 80
  }
}
```

### Evaluation

- target edit correctness;
- AST parse;
- compile;
- unit test;
- unrelated code unchanged.

## 2. ゲーム内Helpのおじいさん

### 要件

- ゲーム内常駐;
- offline/local;
-低Memory;
-ゲームHelpの範囲だけ答える;
-必要ならゲーム状態を参照;
-自由な百科事典LLMである必要はない.

### Composite Recipe

```text
Player message
 + game state
      ↓
Intent Classifier
      ↓
Help Ranker / Retriever
      ↓
Small Response Generator
      ↓
「その鍵なら東の塔じゃよ」
```

質問でなければ:

```text
「何か質問はあるかノ？」
```

のような固定/Template応答でも構いません。

### Modelを分ける理由

Intent判断:
- very small classifier.

Help選択:
- embedding/ranker.

表現:
- small generator or template.

一つの巨大LLMへ全部背負わせません。

### Dataset

Intent:

```json
{
  "text": "東の扉が開かない",
  "label": "help_question"
}
```

Ranking:

```json
{
  "query": "東の扉が開かない",
  "state": "old_castle",
  "candidate": "東の扉には塔の鍵が必要",
  "relevant": true
}
```

Response:

```json
{
  "fact": "東の扉には塔の鍵が必要",
  "style": "old_wizard",
  "output": "東の扉なら、塔の鍵が要るぞい。"
}
```

## 3. Agent Action Selector

Autonomous Agentの次Actionだけ選ぶModel。

```text
Task State
+ available actions
       ↓
Judgment Model
       ↓
run_tests
```

文章生成Modelと分離できます。

## 4. Template / Config Editor

入力:

- current config;
- allowed fields;
- instruction.

Model出力:

- field/value changes.

その後deterministic ToolがJSON/TOML/コード等を安全に書き換えます。

## 5. 設計上の結論

Open LLM Learnerの強みは:

> **何でも生成するModelを作ることではなく、用途に合う最小の学習可能ModelとFlowを組み合わせられること**

です。

LLMはその選択肢の一つです。
