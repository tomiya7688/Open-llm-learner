# Iterative Usage-Guided Prune / Retrain

Status: **draft**

## 1. 目的

Open LLM Learnerは、Modelを一度学習して終わりにせず、

```text
Train
 -> Profile
 -> Prune
 -> Retrain
 -> Evaluate
 -> Profile
 -> Prune
 -> Retrain
 -> ...
```

と繰り返し、小さなModelへ縮めていくTraining PipelineをFirst-classで扱います。

狙いは:

- 学習済みModelから不要な構造を見つける;
- 実際にほとんど寄与しない経路を削る;
- Modelを物理的に小さくする;
- 小さくなったModelを再学習して性能を回復する;
- 次のCycleでは学習・推論Costそのものを下げる.

これは単なるPost-training compressionではなく、**Model creation loopの一部**です。

## 2. 基本Pipeline

```text
Initial Model
    ↓
Train / Pretrain
    ↓
Usage Profile
    ↓
Importance Score
    ↓
Candidate Prune Plan
    ↓
Structural Compaction
    ↓
Recovery Retraining
    ↓
Evaluation Gate
   / \
accept rollback
  |
  v
Next Cycle
```

終了条件:

- target parameter count;
- target RAM/VRAM;
- target latency;
- target artifact size;
- quality floor;
- no further safe reduction;
- max cycles.

## 3. 「使われない経路」の意味

Architectureによって意味が違います。

### Dense Transformer

Dense Transformerでは通常、ほぼ全WeightがForward計算へ参加します。

したがって「使われない」は単純なcall countではなく:

- activation magnitude;
- activation frequency;
- attention head contribution;
- MLP neuron/channel contribution;
- gradient sensitivity;
- ablation impact;
- learned gate/mask score;
- output reconstruction error;

等から評価します。

削除候補の単位:

- Attention Head;
- Head Group;
- MLP neuron/group;
- hidden dimension;
- layer/block;
- embedding dimension;
- structured channel group.

実際のSpeed/Memory削減を狙う場合、まず**構造Pruning**を優先します。

個々のWeightを0にするUnstructured sparsityは、対応Sparse Kernel/RuntimeがなければParameter数だけ減って実計算が減らないことがあります。

### MoE

MoEでは実際のRoutingがあるため:

- expert routing frequency;
- gate probability/weight;
- expert activation strength;
- expert output contribution;
- expert redundancy;

を直接利用できます。

削除候補:

- Expert;
- Expert group;
- Shared/Routed expert allocation;
- Router path.

MoEでは「ほとんど選ばれないExpert」という利用統計が特に分かりやすいですが、単純なFrequencyだけで削るとは限りません。

## 4. Usage Profile

Pruning前にCalibration / Usage Profile Datasetを流します。

Profile DatasetはTraining Datasetと同一である必要はありません。

重要なのは本番用途を代表することです。

例:

### General LLM
- general corpus;
- instruction tasks;
- code;
- reasoning;
- target deployment domains.

### Game Help Model
- actual help questions;
- common game states;
- rare but important questions.

### Code Model
- supported languages;
- common edit types;
- error cases;
- edge-case functions.

## 5. Profile Biasを避ける

特定DatasetだけでProfileすると、そのDatasetで使わない能力を誤って削る危険があります。

そのためProfileは:

- domain-balanced sampling;
- rare-case quota;
- protected task set;
- validation set;
- user-provided production traces;

を組み合わせられます。

「今ほとんど使わない」と「不要」は同義ではありません。

## 6. Importance Score

Pruning MODは複数指標を利用可能です。

例:

```text
importance =
  activation_score
  × sensitivity_score
  × task_coverage_weight
```

あるいは複数MetricをPareto比較しても構いません。

CoreはScore式を理解しません。

Pruning Strategy MODが実装します。

## 7. Pruning単位と実際の小型化

Pruning後はMaskを残すだけではなく、可能ならModel構造を再構成します。

例:

```text
32 attention heads
 -> 24 heads

MLP width 11008
 -> 8192

32 layers
 -> 28 layers
```

その結果:

- Parameter count低下;
- Weight Artifact縮小;
- Training step cost低下;
- Runtime memory低下;
- 対応RuntimeではLatency低下;

を期待できます。

## 8. Recovery Retraining

削除直後は性能が落ちる可能性が高いため、再学習します。

方法:

- same training objective;
- task-specific supervised training;
- continued pretraining;
- SFT;
- LoRA/Adapter recovery;
- full-weight recovery;
- Knowledge Distillation.

## 9. Pre-prune ModelをTeacherにする

Prune直前のModelをTeacher snapshotとして保存できます。

```text
Model N
  ├─ Teacher snapshot
  ↓
Prune
  ↓
Smaller Model N+1
  ↓
Distill from Model N
```

これにより元ModelのOutput distributionやTask behaviorを小型Modelへ戻しやすくします。

Teacher snapshotは次Cycle完了後に保持/削除を選択できます。

## 10. Iterative Compression

一度に50%削るより、小さな割合を複数Cycleで削るStrategyを提供できます。

例:

```text
1.0B
 -> 900M
 -> 780M
 -> 650M
 -> 520M
```

各Cycleで:

1. profile;
2. prune;
3. recover;
4. evaluate.

を行います。

後半CycleほどModelが小さいため、Recovery Training Costも下がります。

## 11. Quality Gate

Cycleを採用する条件をProjectで設定します。

例:

```json
{
  "max_quality_drop": 0.02,
  "min_parameter_reduction": 0.08,
  "max_latency_ms": 100,
  "max_ram_bytes": 1073741824
}
```

EvaluationはTask-specificです。

LLMなら:

- held-out loss/perplexity;
- downstream benchmark;
- generation tests.

Code Modelなら:

- compile;
- unit tests;
- exact edit correctness.

Classifierなら:

- accuracy/F1/calibration.

## 12. Rollback

Evaluation Gateを通らないCycleは:

```text
Candidate compressed model
 -> reject
 -> previous Canonical restored
```

Profile情報自体は分析用Artifactとして保持可能です。

## 13. Adaptive Prune Rate

Cycleごとに削除割合を固定する必要はありません。

例:

- quality margin大 -> prune率を上げる;
- quality drop大 -> prune率を下げる;
- sensitive layer ->保守的;
- redundant layer ->積極的.

Strategy MODが決めます。

## 14. Protected Structures

利用者/Recipeは削除禁止領域を指定できます。

例:

- embedding;
- output head;
- safety classifier head;
- specific experts;
- first/last layers;
- domain-critical module.

## 15. Task-specific Compression

Open LLM LearnerではGeneral-purpose能力を必ず維持する必要はありません。

Task-specific deploymentなら、そのTaskで使わない能力を積極的に削れます。

例:

```text
General Code Model
  ↓ profile on config-edit tasks
Task-specialized compact model
```

これはGeneral-purpose LLM compressionとは別Profileです。

## 16. Model Family Generation

一つの大Modelから複数Sizeを派生できます。

```text
Base 3B
 ├─ 2B
 ├─ 1B
 ├─ 500M
 └─ 200M
```

各Artifactは:

- source parent;
- prune strategy;
- profile dataset;
- recovery training;
- evaluation;

を記録します。

## 17. Model Builder UI

Training Method / Pipelineで:

```text
Compression

[ ] Iterative Prune + Retrain

Target
  ○ Parameter count
  ○ Artifact size
  ○ RAM
  ○ VRAM
  ○ Latency

Profile Dataset
  [ Deployment-like validation set ]

Strategy
  [ Safe Auto ]

Quality floor
  [ 98% of baseline ]
```

Advanced:

- prune axes;
- cycle prune rate;
- max cycles;
- protected structures;
- importance method;
- teacher distillation;
- recovery method.

## 18. Shell

概念的には:

```text
ollm model compress <project>
ollm model profile <artifact>
ollm model prune <artifact>
ollm model recover <artifact>
```

実際のCommand名はCLI実装時に確定します。

## 19. Training Pipeline integration

`training-pipeline`のStageとして:

- profile;
- prune;
- compact;
- recover;
- evaluate;
- promote;

を利用できます。

CycleはFlow/Training Pipelineで表現します。

## 20. Artifact lineage

圧縮ModelはParentを記録します。

```json
{
  "parent": "artifact://model-1b",
  "transformation": {
    "type": "iterative_structured_prune",
    "cycle": 3
  }
}
```

これにより:

- lineage表示;
- rollback;
- model family export;
- reproduction;

が可能です。

## 21. Runtimeとの関係

CompressionとStreamingは別です。

### Compression
Modelそのものを小さくする。

### Streaming
Modelを小さくせず、Memoryへ分割して載せる。

両方組み合わせ可能です。

```text
Large Model
 -> iterative compression
 -> Smaller Model
 -> layer streaming if still needed
```

## 22. Success Criteria

この機能の成功は単なるsparsity ratioではありません。

評価軸:

- actual parameter reduction;
- artifact size;
- measured RAM/VRAM;
- measured latency/throughput;
- retraining cost;
- task quality;
- runtime compatibility.

「Weightを0にしただけで速くならない」結果を成功扱いしません。

## 23. 既存研究との関係

この設計は、既存のstructured pruning、activation/gradient-aware pruning、prune-then-retrain、knowledge distillation等と近い領域です。

Open LLM Learnerでの特徴は、それらを一つのAlgorithmへ固定せず:

- Usage Profile;
- Pruning Strategy MOD;
- Structural Compaction;
- Recovery Training;
- Evaluation Gate;
- Iterative Cycle;

として交換可能なModel Builder Pipelineにすることです。


## 24. 継続学習を重くしにくくする

この方式の重要な目的の一つは、継続的に学習を繰り返すModelのTraining Costを抑えることです。

通常の継続学習では、元Modelの大きさを維持したままDatasetやTaskを追加すると、毎Cycle:

- Forward cost;
- Backward cost;
- optimizer state;
- checkpoint size;
- Runtime memory;

が大きいまま残ります。

Iterative Prune / Retrainでは、各Cycle後にModel自体をCompactにするため:

```text
Train N
 -> Compact N
 -> Train N+1 on smaller model
 -> Compact N+1
 -> ...
```

となり、次回学習の基礎Costを下げられます。

これは単なるInference Compressionではなく、**次のTrainingを軽くするためのCompression**でもあります。

## 25. 必要な要素を残すためのRetention Contract

「よく使われる構造を残す」だけでは不十分です。

Rareだが重要な能力を守るため、ProjectはRetention Contractを持てます。

例:

```json
{
  "protected_tasks": [
    "safety-critical-format",
    "rare-code-edit",
    "fallback-help"
  ],
  "max_drop": {
    "overall": 0.02,
    "protected": 0.005
  }
}
```

Prune Candidateは通常TaskだけでなくProtected Taskでも評価します。

## 26. Rare-case Buffer

Profile Datasetとは別に、削除判定に使うRare-case / Canary Datasetを持てます。

用途:

- 低頻度だが重要な入力;
- 異常系;
- edge case;
- fallback behavior;
-安全上重要な判断;
- 将来残したい汎用能力.

Prune PlanがRare-case評価を大きく落とす場合、その構造は使用頻度が低くても保護できます。

## 27. Parent Artifactを残す

Pruningは完全に不可逆な操作として扱いません。

各Cycleで:

```text
Parent Model
   ↓
Candidate Prune
   ↓
Child Model
```

のlineageを残します。

Childで失われた能力が後から判明した場合:

- ParentへRollback;
- 別Prune Planからやり直す;
- ParentをTeacherにして能力を再蒸留する;

ことができます。

## 28. Capability Regression Test

Model BuilderはCompression前後でCapabilityごとのRegressionを比較できます。

例:

```text
Capability                  Before   After
code.transform               0.94     0.93
decision.evaluate            0.91     0.91
rare.config.edit             0.87     0.72  <- reject
```

平均Scoreだけで採用しません。

## 29. Protected Structureは最後の手段

特定Head/Layer/Expertを手動で保護することもできますが、最初からArchitecture内部を人が固定するより:

1. Protected Taskを定義;
2. Rare-case評価;
3. sensitivityを測る;
4. それでも必要ならStructureをpin;

の順を推奨します。

## 30. 将来のRegrowth

一度削った構造を同じWeightとして復活させることは簡単ではありません。

ただし将来的には:

- pruned parentから再分岐;
-新規小規模Block/Expertを追加;
- adapter capacityを追加;
- teacher distillationで不足能力を再獲得;

のようなRegrowth/Expansion StageをTraining Pipelineへ追加できます。

重要なのは、**現在の小さいModelを永久に唯一の系譜としないこと**です。

Model Family / Parent lineageを保持しておけば、Compact化と再拡張の両方を選べます。
