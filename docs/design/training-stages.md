# Training Stage / Pipeline設計

Status: **draft**

## 1. Trainingは一回のNodeに限定しない

Model作成は複数Stageになり得ます。

```text
Dataset Prepare
 -> Base Training
 -> Fine-tune
 -> Calibration
 -> Evaluation
 -> Promotion
```

## 2. Stage Definition

Stageは:

- input artifact;
- Task/Recipe;
- Training Method;
- config;
- output artifact;
- evaluation gate;

を持ちます。

## 3. Specialized Model例

```text
Teacher label generation
 -> dataset validation
 -> student training
 -> calibration
 -> validation
 -> canonical promotion
```

## 4. LLM例

```text
Pretraining
 -> Continued Pretraining
 -> SFT
 -> Preference Optimization
 -> Evaluation
 -> Export
```

全Stage必須ではありません。

## 5. Checkpoint

各Stageは独立Checkpoint/resume可能にできます。

Stage完了Artifactを次Stageへ渡します。

## 6. Failure

Stage失敗時:

- retry;
- config変更;
- previous checkpoint;
- previous stage artifact;

から再開可能です。

## 7. Evaluation Gate

Stage間にGateを置けます。

例:

```text
Student Training
 -> validation accuracy >= threshold?
    yes -> Calibration
    no  -> stop/revise
```

## 8. Project Save

Resolved Stage PipelineをProjectへ保存します。

UI上のWizardだけに閉じ込めません。

## 9. Flow Integration

Training PipelineはFlowで表現可能です。

ただしModel Builderは利用者がNodeを一つずつ組まなくても標準Flowを生成します。


## 10. Compression Cycle

Training Pipelineは反復Compression Stageを持てます。

```text
profile
 -> prune
 -> compact
 -> recover
 -> evaluate
 -> promote / rollback
```

各CycleのParent Artifact、Profile Artifact、Prune Plan、Evaluationを保存します。

Cycle全体をLoopとして表現しても、Model Builderが展開した明示Stage群として保存しても構いません。
