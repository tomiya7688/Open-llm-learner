# Dataset Contract

Status: **draft**

## 1. 目的

Dataset Contractは「このTaskのTraining Exampleが何を意味するか」を定義します。

CSV/JSONL/DB等の物理形式そのものではありません。

同じCSVでも、どのColumnがInputでどのColumnがTargetかをMappingしてTaskへ接続します。

## 2. 例: Classification

Contract:

```json
{
  "id": "official.dataset.text-classification.v1",
  "inputs": {
    "text": {"type": "string"}
  },
  "targets": {
    "label": {"type": "string"}
  }
}
```

Dataset Mapping:

```json
{
  "source": "dataset.csv",
  "map": {
    "sentence": "inputs.text",
    "category": "targets.label"
  }
}
```

## 3. 例: Code Transform

```json
{
  "inputs": {
    "source": {"type": "string"},
    "instruction": {"type": "string"}
  },
  "targets": {
    "result": {"type": "string"}
  }
}
```

Structured Edit RecipeではTarget Contractを変更できます。

## 4. Preflight

Dataset MODはContractに基づいて:

- missing field;
- type mismatch;
- empty target;
- label distribution;
- duplicate;
- sequence/token length;
- train/validation/test split;
- leakage risk;
- invalid structured target;

等を検査できます。

どの検査を提供するかはMOD次第です。

## 5. Dataset Split

ProjectはSplitを明示的に記録します。

```text
train
validation
test
```

Skill evolutionやModel selectionではvalidationとtestを混同しません。

## 6. Generated / Teacher-labeled Data

Dataset Contractは人手Labelだけを前提にしません。

Teacher Model、Rule、既存Program、Simulation等で生成したDataも同じContractへ入れられます。

Origin metadataを保持できるようにします。

## 7. Versioning

Contract IDとVersionを持ちます。

Dataset自体のDigest/Versionとは別です。

Task/Recipe側は対応Contract Versionを宣言できます。
