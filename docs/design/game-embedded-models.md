# ゲーム・アプリ組み込みModel設計

Status: **draft**

## 1. 目的

ゲームやDesktop Appへ常駐する小型ModelをFirst-class use caseとして扱います。

重要要件:

- offline;
- low memory;
- low latency;
- bounded domain;
- predictable behavior;
- easy stop/unload;
- portable runtime export.

## 2. Game Help NPC

例:

```text
Player message
 + game state
       ↓
Intent Classifier
       ↓
Help Ranker
       ↓
Response Formatter/Small Generator
```

## 3. State Input

Gameから必要状態だけ渡します。

```json
{
  "area": "old_castle",
  "quest": "find_east_key",
  "inventory": ["wooden_sword"]
}
```

ModelがGame process全体を直接読む必要はありません。

## 4. Knowledge

Help Knowledgeは:

- static JSON;
- text entries;
- embedding index;
- structured database;

として外に持てます。

Model Weightへ全部暗記させる必要はありません。

## 5. Dialogue Style

StyleだけSmall Generatorへ担当させてもよい。

```text
Fact: 東の扉には塔の鍵が必要
Style: old_wizard
      ↓
「東の扉なら、塔の鍵が要るぞい」
```

## 6. No-question Behavior

質問でない場合はModelを呼ばず:

```text
「何か質問はあるかノ？」
```

というRule/Templateでも構いません。

Flowが判断します。

## 7. Runtime Export

Target例:

- ONNX Runtime;
- native app runtime;
- GGUF-compatible runtime;
- self-contained Package;
- game-engine-specific Community Exporter.

## 8. Resource

Game本体のRAM/VRAMを圧迫しないよう、Model Budgetを明示します。

Resource GovernorとRuntime Exportのprofileへ:

- max RAM;
- max VRAM;
- CPU only;
- startup target;

等を渡せます。

## 9. Versioning

Game UpdateとModel/Knowledge versionを別に持てます。

Help DBだけ更新してModelを再学習しない構成も可能です。
