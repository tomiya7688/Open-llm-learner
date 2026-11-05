# Runtime Capability Contract

Status: **draft**

## 1. RuntimeはModelをCapabilityとして見せる

Runtime MODはModel Artifactを読み込み、semantic capabilityを提供します。

例:

- model.text.generate;
- model.text.transform;
- model.code.transform;
- model.text.classify;
- model.text.extract;
- model.text.fill;
- model.rank;
- model.decision.evaluate;
- model.embedding.create.

## 2. Capability Schema

各Capabilityは:

- input schema;
- output schema;
- streaming support;
- batching support;
- cancellation;
- resource estimate;
- determinism controls;

を宣言できます。

## 3. Batch

Classifier/Embedding/Ranker等ではBatchが重要です。

Runtimeは:

```json
{
  "batch": [
    {"text": "a"},
    {"text": "b"}
  ]
}
```

のようなBatch Actionを提供可能です。

CoreにBatch semanticsを固定しません。

## 4. Streaming Output

Text Generatorはtoken/event streamingを提供してもよい。

ただしFlow v1でeventを通常Data Edgeに流すこととは別です。

Runtime eventはPreview/Telemetryとして扱い、Final resultはterminal resultで返します。

## 5. Structured Output

Classifier/Decision/Extractor等はtyped resultを返します。

例:

```json
{
  "label": "help_question",
  "probabilities": {
    "help_question": 0.94,
    "other": 0.06
  }
}
```

## 6. Runtime Metadata

Runtimeは:

- supported artifact formats;
- supported capabilities;
- max sequence/context;
- supported dtype;
- batch;
- memory modes;
- accelerator support;

をinspect/preflightで返せます。

## 7. Hot Load / Unload

Local AIでは複数Modelを交代で使うため:

- load;
- unload;
- trim;
- replan;

が重要です。

Resource Governorと接続します。

## 8. Determinism

Task Modelではtemperatureのような生成設定が不要な場合があります。

Runtime settingsはCapabilityごとに異なります。

## 9. Error

共通Error:

- unsupported artifact;
- incompatible recipe/runtime;
- missing tokenizer;
- missing custom op;
- resource limit;
- invalid input;
- cancelled.

Model-specific detailsはdetailsへ入れます。
