# Model Artifact Lifecycle

Status: **draft**

## 1. Model Artifactは段階を持つ

```text
Source/Base
 -> Training Checkpoint
 -> Candidate
 -> Evaluated
 -> Canonical
 -> Derived Runtime Artifacts
 -> Export / Package
```

## 2. Source/Base Artifact

学習元Model。

- external reference;
- imported Safetensors;
- imported PyTorch checkpoint;
- previous Open LLM Learner artifact.

## 3. Training Checkpoint

Training再開用。

含む可能性:

- model state;
- optimizer state;
- scheduler;
- RNG;
- step;
- adapter state.

Portable Model Artifactとは区別します。

## 4. Candidate Artifact

Training完了後、Evaluation前の成果物。

複数Candidateが存在しても構いません。

例:

- epoch 3;
- epoch 5;
- best validation;
- distilled student A/B.

## 5. Evaluated Artifact

Evaluation結果が紐づいたModel Artifact。

Metadata例:

- dataset split;
- evaluator version;
- score;
- latency;
- memory;
- test result.

## 6. Canonical Artifact

Projectが正本として選んだWeight/Graph Artifact。

Runtime-derived形式とは分離します。

## 7. Derived Runtime Artifact

例:

- GGUF;
- AirLLM prepared;
- ONNX optimized;
- native layer shards.

必ずsource digestを持ちます。

Canonicalが変わればstale判定できます。

## 8. Promotion

CandidateからCanonicalへPromotionする操作を明示します。

自動Promotion条件も設定可能:

- validation score;
- all tests pass;
- memory target;
- manual approval.

## 9. Rollback

以前のCanonicalへ戻せます。

Training/Exportの履歴を破壊しません。

## 10. Garbage Collection

Cacheや古いDerived Artifactは削除可能ですが:

- Canonical;
- explicitly pinned artifact;
- required checkpoint;

を誤削除しない。

## 11. Package

PackageはArtifactを:

- bundle;
- reference;
- runtime-prepared bundle;

として取り込めます。

Packageを削除してもCanonical Weightは別物です。
