# Model Builder Flow

Status: **draft**

## 1. Model Builderは標準Flowを生成する

利用者はGUI Wizardを使ってもShellを使っても、最終的にはProject内の普通のFlow/Configへ落ちます。

## 2. Creation Flow

```text
Select Task
 -> Resolve Dataset Contract
 -> Map Dataset
 -> Dataset Preflight
 -> Recommend Recipe
 -> Resolve Config
 -> Resource Preflight
 -> Train
 -> Evaluate
 -> Promote Candidate
 -> Test Runtime
 -> Export
```

## 3. Task Selection

Task definitionを選択します。

Custom Taskも可能。

## 4. Dataset Mapping

Physical DatasetをContractへMapping。

Invalid sample数等を表示します。

## 5. Recipe Recommendation

対応Recipeを:

- quality expectation;
- memory;
- latency;
- training cost;
- target runtime;

と共に比較します。

## 6. Training

Training MODはprogress eventを返します。

Model Builderは:

- loss;
- metric;
- step;
- RAM/VRAM;
- checkpoint;

を表示できます。

## 7. Evaluation

Task-specific Evaluatorを実行。

Threshold未達なら:

- keep candidate;
- revise training;
- try another Recipe;
- manual promote;

を選択可能。

## 8. Promotion

CandidateをCanonical ArtifactへPromotion。

自動/手動。

## 9. Runtime Test

選択Runtimeで実際にload/inference。

Training successだけで終わらない。

## 10. Export

Weight / Runtime / Package。

## 11. Resume

途中でProjectを閉じても:

- mapped dataset;
- resolved config;
- checkpoints;
- candidate artifact;
- evaluation;

から再開できるようにします。

## 12. Shell

Shellも同じFlowを段階実行できます。

例概念:

```text
ollm model task ...
ollm model dataset ...
ollm model recipe ...
ollm model preflight
ollm model train
ollm model evaluate
ollm model promote
ollm export ...
```

Command名はCLI実装時に確定します。
