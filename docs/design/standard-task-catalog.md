# Standard Task Catalog

Status: **draft**

この文書はOpen LLM Learnerが最初から理解できる代表Taskの意味を定義します。

TaskはCore予約機能ではありません。
公式Task definitionとして提供し、Community/User Taskも同じ仕組みで追加できます。

## text_generation

入力:
- prompt;
- optional context.

出力:
- text.

Capability:
- model.text.generate.

用途:
- chat;
- completion;
- story;
- code generation;
- general instruction following.

## text_transform

入力:
- source;
- instruction;
- optional context.

出力:
- transformed text.

Capability:
- model.text.transform.

用途:
- rewrite;
- translation;
- format conversion;
- constrained rewriting.

## code_transform

入力:
- source code or structured code context;
- instruction.

出力:
- updated code or structured edit.

Capability:
- model.code.transform.

Evaluation:
- parse/build/test;
- expected edit;
- unrelated-change check.

## text_classification

入力:
- text/context.

出力:
- label;
- optional probabilities.

Capability:
- model.text.classify.

用途:
- intent;
- routing;
- category;
- filter.

## decision

入力:
- text-bearing state;
- questions/candidates.

出力:
- boolean;
- choice;
- scores/probabilities.

Capability:
- model.decision.evaluate.

## extraction

入力:
- text/context.

出力:
- spans;
- fields;
- typed values.

Capability:
- model.text.extract.

用途:
- variable;
- entity;
- config field;
- code symbol;
- structured information.

## template_fill

入力:
- template;
- context;
- instruction.

出力:
- field values or completed template.

Capability:
- model.text.fill.

## ranking

入力:
- query/context;
- candidates.

出力:
- score per candidate;
- ranking.

Capability:
- model.rank.

用途:
- Help entry selection;
- retrieval rerank;
- action ranking;
- patch ranking.

## embedding

入力:
- text.

出力:
- vector.

Capability:
- model.embedding.create.

## custom

利用者が:

- input schema;
- output schema;
- Dataset Contract;
- target Capability;
- Recipe;

を定義します。

## Task Variant

同じTask familyでもVariantを持てます。

例:

```text
code_transform
├─ full_code
├─ unified_diff
└─ structured_edit
```

VariantをCore enumにはしません。
Task/Recipe definitionが宣言します。
