# TokenizerとText Representation

Status: **draft**

## 1. TokenizerはTask Recipeの一部

TokenizerをCoreの共通前提にしません。

Task/Recipeによって必要なText Representationは異なります。

例:

- Decoder LLM: subword tokenizer;
- Encoder classifier: subword tokenizer;
- Code Transform: code-aware tokenizer / symbol-aware representation;
- Structured Edit: tokenized text + structured fields;
- Character/NPC model: compact domain vocabulary;
- Judgment Model: text encoder + typed candidate representation.

## 2. Tokenizerの選択

Model Builderは以下を選択肢として扱います。

- 既存Base ModelのTokenizerを再利用;
- 新規Tokenizerを学習;
- 固定Vocabulary;
- Character-level / byte-level;
- external tokenizer MOD;
- structured serializer.

Tokenizer selectionはRecipe側が対応範囲を宣言します。

## 3. Dataset Preflight

Tokenizerを決めた後、Datasetへ対して:

- token length distribution;
- truncation rate;
- unknown token rate;
- special token collision;
- padding cost;
- packed sequence efficiency;

などを確認できます。

## 4. Structured Input

Capability inputがObjectの場合、Runtime MODがModel入力へserializeします。

例:

```json
{
  "function_name": "checkLimit",
  "condition": "value > limit",
  "instruction": "80以上なら警告"
}
```

を単純にJSON文字列化する必要はありません。

Recipeは:

- field separator;
- field embedding;
- segment ids;
- type embedding;
- multiple encoder inputs;

などを選べます。

## 5. Output Representation

Outputも自然言語に限定しません。

- text;
- class id;
- span;
- score;
- structured edit;
- bounded choice.

Training target representationとRuntime output mappingをRecipeが定義します。

## 6. Tokenizer Artifact

TokenizerはModel Artifactと別Referenceとして保持可能にします。

これにより:

- 同じTokenizerを複数Modelで再利用;
- Tokenizerだけ差し替え;
- Runtime Export時に必要ファイルを同梱;

が可能です。

## 7. Compatibility

Weight Export時はTokenizerを必ずWeightファイルへ埋め込むとは限りません。

Runtime Exporterが対象Runtimeの期待形式へ:

- tokenizer.json;
- vocab;
- merges;
- sentencepiece model;
- metadata;

等を変換します。
