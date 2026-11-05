# Testing / Compatibility設計

Status: **draft**

## 1. 何を保証するかを分ける

Open LLM Learnerは:

- Architecture capability;
- Officially tested path;
- Community-supported path;

を区別します。

## 2. Test Matrix

最低限の軸:

- Task;
- Recipe;
- Training Method;
- Artifact Format;
- Runtime;
- Export Target;
- OS;
- accelerator;
- memory mode.

全組み合わせを公式QAする必要はありません。

## 3. Reference Paths

v1.0では少なくとも:

### Specialized Task Model
Task -> Dataset -> Recipe -> Train -> Evaluate -> Runtime -> Export

### LLM
Dataset -> LLM Recipe -> FT/LoRA -> Evaluate -> Runtime -> Export

### Agent
Model(s) -> Capability -> Tool loop -> Verify -> Final

をReference E2Eとして固定します。

## 4. Golden Tests

小型Task ModelはDeterministic Testを作りやすい。

例:

- classification labels;
- structured edits;
- extraction fields;
- rank order.

これらをRelease Regressionへ使います。

## 5. Runtime Export Tests

Exporterごとに:

- generated files;
- target application load;
- basic inference;
- metadata;
- version compatibility;

を確認します。

## 6. Model Artifact Round-trip

可能な形式では:

```text
Import
 -> Canonical
 -> Export
 -> Re-import
 -> inference/evaluation
```

を確認します。

## 7. Resource Tests

- Stop;
- Force Stop;
- memory headroom;
- streaming fallback;
- checkpoint/resume;
- low priority preemption;

もProduct QA対象です。

## 8. Unsupported

未検証構成は:

- Unsupported;
- Experimental;
- Community;

をUI/Docsへ表示します。

「設計上可能」と「公式保証済み」を混同しません。
