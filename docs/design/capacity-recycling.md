# Capacity Recycling / Structural Reallocation

Status: **draft**

## 1. 目的

Open LLM LearnerではPruningをModelを小さくするためだけに使いません。

もう一つの重要なTraining方式として、役に立っていないParameter Capacityを捨て、そのParameter予算を新しい学習余地として再利用します。

```text
3.0B Model
   ↓ Usage / Importance Profile
2.5B retained structure
+ 0.5B low-value structure removed
   ↓
0.5B fresh capacityを再配置
   ↓
3.0B Model
(new topology / refreshed capacity)
   ↓
重点学習
   ↓
統合学習
```

最終Modelを3Bのまま維持しても構いません。

目的はParameter数削減ではなく、限られたParameter Budgetの中身を入れ替えることです。

## 2. Compressionとの違い

### Compression Mode

```text
3.0B -> prune -> 2.5B -> retrain
```

目的はModel size、memory、latency、training costを下げることです。

### Recycling Mode

```text
3.0B -> prune 0.5B -> retained 2.5B
     -> allocate fresh 0.5B -> train -> 3.0B
```

目的はobsolete/redundant capacityを捨て、新しいTaskやDomainへcapacityを移し、総Parameter Budgetを増やさず新しい学習余地を作ることです。

## 3. 基本Cycle

```text
Current Model
    ↓
Profile old + current workloads
    ↓
Importance / Retention analysis
    ↓
Retain high-value structure
    ↓
Remove low-value capacity
    ↓
Capacity Budget = removed parameters
    ↓
Find where new capacity is needed
    ↓
Allocate fresh structure
    ↓
Focused Training
    ↓
Consolidation Training
    ↓
Regression / New-task Evaluation
   / \
accept rollback
  |
  v
Next Cycle
```

## 4. Capacity Budget

Recycling CycleはParameter Budgetを明示します。

```json
{
  "total_parameter_budget": 3000000000,
  "recycle_target_parameters": 500000000,
  "max_growth_over_budget": 0
}
```

原則は retained + fresh <= configured total budget です。

## 5. 何を削るか

削除側はIterative Prune / Retrainと同じProfileを利用できます。

Dense Transformerではattention heads/head groups、MLP channels/groups、hidden subspaces、blocks/layers等を候補にします。

MoEではunderused/redundant experts、expert groups、router paths等を候補にします。

ただしFrequencyだけでは決めず、Retention PolicyとSensitivityを併用します。

## 6. どこへ新しいCapacityを足すか

削除量と追加先は同じ場所である必要はありません。

```text
Layer 4 MLPから100M削除
Layer 11 Attentionから50M削除
Layer 18 MLPから150M削除
Unused expertsから200M削除

Total freed: 500M
```

これを:

```text
new code-specialized block       +180M
new MLP capacity in layers 20-24 +120M
new experts for new domain       +200M
```

へ再配置できます。

## 7. Capacity Demand Profile

Usage Profileは何を残すべきかを測ります。

Capacity Demand Profileはどこへ新しい容量を使うべきかを測ります。

Signal候補:
- gradient magnitude;
- residual/reconstruction error;
- persistent training loss by layer/module;
- activation saturation;
- adapter update magnitude;
- teacher/student disagreement;
- task-specific attribution;
- router pressure / expert overload.

この二つは別Artifactにできます。

## 8. Fresh Capacity

新しく追加するParameterは古い低価値Weightをそのまま移動するのではなく、原則としてfresh trainable capacityとして扱います。

初期化候補:
- random initialization;
- parent-derived initialization;
- teacher-derived initialization;
- duplicated + perturbed structure;
- function-preserving initialization;
- zero/near-zero residual initialization.

## 9. Function-preserving Growth

可能なArchitectureでは、新しい構造を追加した直後の出力を元Modelに近く保ちます。

例:
- residual branch outputをzero-init;
- new MLP channelのoutput contributionをzero-init;
- new adapter/gated blockをclosed gateで開始;
- duplicated structureを分割/perturb.

完全なFunction Preservationが難しい場合は追加直後のRegression Testを必須にします。

## 10. Focused Training

Fresh Capacity追加直後は全3Bを同じLearning Rateで学習する必要はありません。

```text
Retained 2.5B
  -> frozen / low LR

Fresh 0.5B
  -> high LR / primary train target
```

設定例:

```json
{
  "retained_parameters": {
    "mode": "low_lr",
    "lr_scale": 0.05
  },
  "fresh_parameters": {
    "mode": "train",
    "lr_scale": 1.0
  }
}
```

## 11. Consolidation Training

Focused Training後に全体を軽くJoint Trainingします。

```text
Focused phase
  fresh capacity learns new task
       ↓
Consolidation phase
  retained + fresh jointly adjust
```

目的は新Capacityと既存Representationを馴染ませ、routing/attention/normalizationを調整することです。

## 12. Replay / Teacher Retention

古い能力維持のため、Consolidationではprotected old-task samples、rare-case buffer、parent-model distillation、previous task examples等を混ぜられます。

完全なRetention保証はできません。

## 13. Parent ModelをTeacherにする

```text
Parent 3B
  ├─────────────┐
  ↓             │
Recycle 0.5B    │ teacher outputs
  ↓             │
New 3B Student <-┘
```

新Taskは新Datasetで学習し、旧TaskはParentからDistillationできます。

## 14. Capacity Allocation Strategy

Allocation StrategyはMODとして交換可能にします。

- Same-place Regrowth: 削った場所へfresh structureを再生成。
- Demand-guided Reallocation: 新Task gradient等を見て別Layerへ移す。
- New Specialized Module: 新Task専用Block/Adapter/Expertとして追加。
- MoE Expert Recycling: 低価値Expertを削除しfresh Expertとして再初期化。

## 15. MoEは特に相性が良い

```text
Expert 12: almost unused
Expert 27: redundant
        ↓
remove/reinitialize
        ↓
fresh Expert A/B
        ↓
new domain/taskへrouting
```

総Expert数を維持したまま中身を入れ替えられます。Routerも再学習します。

## 16. Dense Transformerでも行う

Dense TransformerではParameterを完全に独立した経路として扱いにくいため、grouped MLP channels、attention head groups、replaceable residual blocks、recyclable side modules等のStructure単位で扱います。

## 17. Recycling-friendly Architecture

継続学習前提RecipeではArchitecture自体をRecyclingしやすくできます。

例:
- modular MLP groups;
- gated residual blocks;
- replaceable experts;
- separated task-capacity blocks;
- structured head groups.

standard-compatible Recipeとrecycling-optimized Recipeは分けて提供できます。

## 18. Total Budgetを固定したContinual Learning

```text
Fixed 3B parameter budget
Cycle 1: general language
Cycle 2: code ability追加
Cycle 3: domain knowledge更新
Cycle 4: new tool/action behavior
```

各Cycleで古い低価値Capacityを入れ替え、Model sizeは3B前後を維持します。

## 19. Budget Split / Plasticity Zones

全Parameterを同じ扱いにしません。

```text
3B total

2.0B stable core
0.7B semi-plastic
0.3B recycle pool
```

Parameter groupへstable / adaptive / recyclable / protected等のPlasticity classを付けられます。

## 20. 成功条件

Recycling Cycleは以下を比較します。

Old capability:
- protected task quality;
- rare cases;
- regression.

New capability:
- new task score;
- adaptation speed.

Efficiency:
- total parameters;
- training FLOPs;
- peak memory;
- latency;
- artifact size.

Plasticity:
- fresh capacityが実際にgradient/updateを受けているか;
- retained capacityへ過度にupdateが集中していないか.

## 21. Failure / Rollback

新Modelが旧能力を失いすぎた場合:
- ParentへRollback;
- recycle量を減らす;
-別Allocationを使う;
- Focused/Consolidation比率を変える;
- replay/distillationを増やす;
- protected structuresを増やす.

## 22. Compression / Streamingとの組み合わせ

CompressionはModelを小さくする。

Recyclingは固定Budget内でParameterの役割を入れ替える。

StreamingはModelを小さくせずMemoryへ分割して載せる。

三つは組み合わせ可能です。

## 23. Model Builder UI

```text
Continual Learning

Capacity Policy
  ○ Grow freely
  ● Fixed budget recycling
  ○ Compress over time

Total parameter budget
  3.0B

Recycle per cycle
  0.5B

Old capability retention
  [ Protected Tasks ... ]

Fresh capacity training
  Focused phase: 20%
  Consolidation: 80%

Allocation
  [ Safe Auto ]
```

## 24. Artifact Lineage

Recyclingは同じ3BでもTopologyが変わる可能性があるため、ParentとTransformationを記録します。

```json
{
  "parent": "artifact://model-cycle-4",
  "transformation": {
    "type": "capacity_recycling",
    "removed_parameters": 500000000,
    "allocated_parameters": 500000000,
    "cycle": 5
  }
}
```

## 25. 設計上の位置づけ

Continual Training policyを三つに分けます。

```text
A. Grow
   新Capacityを足してModelを大きくする

B. Compress
   不要Capacityを削ってModelを小さくする

C. Recycle
   不要Capacityを削り、同じBudget内で新Capacityへ置換する
```

今回の方式はC. Recycleです。