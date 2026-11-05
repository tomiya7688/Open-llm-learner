# Code Model設計

Status: **draft**

## 1. Code = free-form generationとは限らない

Code Taskを以下へ分けます。

- code generation;
- code completion;
- code transformation;
- structured edit;
- classification/routing;
- symbol extraction;
- error classification;
- patch ranking.

## 2. Context Builder

Repository全体をModelへ渡す代わりに:

- target file;
- symbol;
- AST;
- type;
- references;
- diagnostics;
- tests;
- current diff;

を抽出します。

Parser/IndexerはMODです。

## 3. Structured Edit

推奨例:

```json
{
  "operation": "replace_expression",
  "symbol": "checkLimit",
  "path": ["if", "condition"],
  "expression": "value >= 80"
}
```

Modelは変更意図だけ返します。

Edit ApplicatorがAST/Parserを使って適用します。

## 4. Patch Generation

より自由な変更では:

```text
context + instruction
 -> code transform model
 -> patch
 -> apply
 -> parse/build/test
```

## 5. Function Preset

利用者はFunction/Templateへ名前を付けられます。

例:

```text
preset id: limit-warning
description: 上限超過時に警告する関数
```

Instruction:

```text
limit-warningで、80以上なら警告
```

Context ResolverがPresetを探し、必要部分だけModelへ渡します。

## 6. Dataset Generation

Presetから自動的にMutationを生成できます。

- operator change;
- literal change;
- variable change;
- branch insertion/removal;
- function call replacement.

Ruleで正解Editを生成し、InstructionだけTeacher Modelで作る方式も可能です。

## 7. Evaluation

Code ModelはExecution-based evaluationを優先します。

- parse;
- format;
- compile;
- unit test;
- static analysis;
- expected behavior;
- unrelated change count.

## 8. Capability

- model.code.transform;
- model.text.classify;
- model.text.extract;
- model.rank;

など複数Capabilityを利用できます。

## 9. Agentとの違い

Code ModelはModel Capability。
Coding AgentはModel + Tools + Flow。

この二つを混同しません。
