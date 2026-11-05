# Evaluation設計

Status: **draft**

## 1. Training完了とModel完成は別

Lossが下がっただけではModel完成とはしません。

TaskごとにEvaluation Contractを持ちます。

```text
Train
 -> Validation
 -> Task Evaluation
 -> Runtime Test
 -> Export
```

## 2. EvaluatorはMOD

CoreはMetric意味を理解しません。

Evaluator MODが:

- input;
- expected;
- actual;
- trace/artifact;

を受け取りScore/Detailsを返します。

共通Capability:
- `evaluation.score`

## 3. Deterministic Evaluationを優先できるTask

### Code Transform
- parse;
- compile;
- unit test;
- AST equality;
- unrelated code preservation.

### Extraction
- exact field match;
- span match.

### Classification
- label correctness;
- confusion matrix;
- calibration.

### Ranking
- top-k correctness;
- ranking metric.

### Template Fill
- required field correctness;
- schema validation.

## 4. Generative Model

Free-form textでは単一Metricへ固定しません。

- held-out loss;
- exact/structured test;
- domain evaluator;
- model-based evaluator;
- downstream task success;

を組み合わせます。

## 5. Model Comparison

Recipe RecommendationやModel selectionでは複数Candidateを同じEvaluation Setで比較できます。

```text
Recipe A
Recipe B
Recipe C
   ↓
same validation tasks
   ↓
score + latency + memory
```

Accuracyだけでなく:

- latency;
- peak RAM/VRAM;
- artifact size;
- tokens/sec;
- startup time;

も比較できます。

## 6. Deployment-aware Evaluation

小型専用Modelの価値はQualityだけでなくDeployment costにもあります。

Model Builderは:

> 1%良い巨大Model

と

> 十分な精度で10倍軽い小型Model

を同じ表で比較できる設計にします。

## 7. Acceptance Gate

Projectは最低ScoreやTest passをExport Gateにできます。

ただし利用者が明示してOverride可能にします。
