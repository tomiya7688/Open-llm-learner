# Dataset Pipeline

Status: **draft**

## 1. DatasetはTraining前にPipelineを持つ

DatasetをそのままTraining MODへ渡すだけではありません。

```text
Source
 -> Import
 -> Normalize
 -> Map to Dataset Contract
 -> Validate
 -> Split
 -> Optional Transform/Augment
 -> Cache
 -> Train
```

各段階はMODで差し替え可能です。

## 2. Source

Source例:

- JSONL;
- CSV;
- text files;
- source code tree;
- database export;
- chat logs;
- synthetic data;
- model-generated labels;
- game help database.

Coreは物理形式を理解しません。

## 3. Mapping

Physical dataset columns/fieldsをDataset ContractへMappingします。

例:

```json
{
  "map": {
    "before_code": "inputs.source",
    "instruction_text": "inputs.instruction",
    "after_code": "targets.result"
  }
}
```

## 4. Normalize

Normalize例:

- encoding;
- newline;
- whitespace;
- text cleanup;
- code normalization;
- label normalization;
- duplicate removal;
- invalid sample filtering.

Normalize処理はTask-specificでも構いません。

## 5. Split

train / validation / testは明示的に扱います。

Split方法:

- random;
- stratified;
- group-aware;
- time-aware;
- project/repository-aware;
- user-provided.

Code Taskでは同一Repository由来Sampleがtrain/testへ跨がないようにするなど、Leakage対策をRecipeが推奨できます。

## 6. Synthetic Data

Teacher ModelやRuleからDatasetを生成できます。

例:

```text
Function Presets
 -> Mutation Generator
 -> Instruction Generator
 -> Target Edit
 -> Validation
 -> Dataset
```

生成元をmetadataへ残します。

## 7. Human Correction

Model生成Labelを人が修正した場合:

- original label;
- corrected label;
- source;
- reviewer metadata;

を保持できます。

## 8. Dataset Artifact

Processed DatasetはArtifactとしてCacheできます。

Source Datasetを破壊しません。

```text
raw source
 -> processed artifact
 -> tokenized/cache artifact
```

## 9. Reproducibility

Projectは:

- source reference;
- mapping;
- transforms;
- split seed/policy;
- generated artifact digest;

を保存します。

## 10. Streaming Dataset

巨大Datasetでは全件をMemoryへ読み込まない。

Dataset MODは:

- streaming read;
- chunking;
- shuffle buffer;
- sample index;

を提供可能です。
