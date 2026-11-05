# Model Size / Resource Budget

Status: **draft**

## 1. Parameter countだけでModelを選ばない

Model Builderは「何Bか」だけでModelを評価しません。

重要な軸:

- Task quality;
- artifact size;
- RAM;
- VRAM;
- startup time;
- latency;
- throughput;
- disk;
- power/thermal constraints where measurable;
- target device.

## 2. User-facing target

利用者は必要なら:

- target artifact size;
- target RAM;
- target VRAM;
- target latency;
- target device class;

を指定できます。

例:

```text
Target
  RAM: <= 2 GB
  VRAM: none
  First response: <= 200 ms
```

Recipe Recommendationはこの条件を使います。

## 3. Architecture Resolution

Task + BudgetからRecipeが具体構成を解決します。

例:

```text
Task: help intent classification
Budget: CPU / 512 MB
        ↓
Tiny Encoder
4 layers
hidden 256
classification head
```

利用者はAdvancedでoverrideできます。

## 4. LLM

LLMでは:

- target parameter scale;
- context length;
- dense/MoE;
- total/active parameters;
- quantization target;
- memory mode;

を扱います。

## 5. MoE

MoEでは必ず:

- target_total_parameters;
- target_active_parameters;

を分けます。

Active Parametersはresident memoryの意味ではありません。

## 6. Budget Failure

指定Budgetに収まらない場合:

- smaller recipe;
- quantized runtime;
- distillation;
- lower context;
- streaming/offload;
- different architecture;

を候補として出します。

勝手にQualityを落とす設定へ変更せず、Resolved Planを表示します。

## 7. Measured vs Estimated

Resource表示は:

- Estimated;
- Measured;
- Guaranteed limit;

を区別します。

推定値を保証値として表示しません。
