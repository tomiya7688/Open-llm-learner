# Contextと構造化Input

Status: **draft**

## 1. 小型Modelへ全部読ませない

用途特化Modelでは、巨大Contextをそのまま渡す必要がない場合があります。

例: Code Edit ModelにRepository全体を渡す代わりに:

- target function;
- variables;
- types;
- current condition;
- relevant callers;
- requested change;

だけを入力できます。

## 2. Context Builder

Context BuilderはModelそのものではなくMOD/Flowとして実装できます。

```text
Repository
 -> Parser / Indexer
 -> relevant context
 -> Model
```

ゲームなら:

```text
Game State
 + Help DB
 -> Context Builder
 -> Help Model
```

## 3. Structured Input

Model Capabilityは必ずしも一つのStringだけを受け取る必要はありません。

例:

```json
{
  "function": {
    "name": "checkLimit",
    "parameters": ["value", "limit"],
    "condition": "value > limit"
  },
  "instruction": "80以上なら警告してください"
}
```

Runtime MODがTokenizer/serializationを決めます。

## 4. Outputも構造化可能

Modelは自然言語全文ではなく:

- enum;
- boolean;
- span;
- edit operation;
- JSON-like typed structure;

を返せます。

これにより小型Modelとdeterministic Toolを組み合わせやすくなります。

## 5. Parserは交換可能

コードのAST、ゲーム状態、独自DSL等をCoreへ入れません。

Parser/Context Builder MODとして追加します。
