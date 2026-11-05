# 学習方式

Status: **draft**

この文書はLLMだけでなく、用途特化モデルも含むTraining Methodの設計をまとめます。

## 1. Train from scratch

新しいModel Weightを初期化してDatasetから学習します。

用途:
- 小型分類器;
- 小型Transformer;
-独自Tokenizer/語彙を持つモデル;
-十分なDataがある特殊モデル.

Recipeが:

- architecture;
- initialization;
- objective;
- optimizer;
- schedule;

を解決します。

## 2. Full Fine-tuning

既存Modelの対象Weightを更新します。

用途:
- 既存Encoderを分類Taskへ適応;
- Encoder-Decoderを変換Taskへ適応;
- LLM全体の適応.

出力はPortable Full Weight Artifactです。

## 3. Head-only / Partial Fine-tuning

Base Modelを凍結し、分類HeadやTask Headだけ学習します。

用途:
- Classification;
- Extraction;
- Ranking;
- Decision.

大きなBaseを使いつつTraining costを抑えられます。

## 4. Adapter / LoRA

Base Modelを主に凍結し、追加Parameterを学習します。

用途:
- LLM;
- Encoder;
- Encoder-Decoder;
-その他対応Architecture.

Adapter ArtifactとBase Referenceを保存します。

## 5. Continued Pretraining

既存Language ModelをドメインCorpusでLanguage Modeling Objectiveのまま追加学習します。

その後SFT等へ進められます。

## 6. Supervised Task Training

Input → Targetが明確なDatasetで学習します。

例:

- classification;
- code edit;
- translation;
- extraction;
- structured decision;
- instruction response.

Task ContractがData Mappingを定義します。

## 7. Distillation

Teacher Modelの出力やscoreを使い、小さいStudent ModelへTask能力を移します。

Open LLM Learnerでは特に重要です。

例:

```text
Large Coding Model
   ↓ generates labels/edits
Task Dataset
   ↓
Small Code Edit Model
```

または:

```text
Strong Agent Model
  ↓ action choices
Small Judgment Model
```

TeacherはTraining後に不要でも構いません。

## 8. Preference / Ranking Training

候補A/B、順位、Preference等をDatasetとして学習する方式です。

用途:
- Ranker;
- Reward/Judgment Model;
- LLM preference optimization.

具体的AlgorithmはTraining MOD Recipeとして差し替えます。

## 9. Partitioned Streaming Training

Model全体をDeviceへ常駐させずPartition単位で学習します。

対象:
- Adapter/LoRA;
- exact Full FTが成立する構成;
- future expert/partition training.

これはMemory execution strategyでありTaskそのものではありません。

## 10. Multi-stage Training

一つのModel作成Projectで複数Stageを持てます。

例:

```text
Pretraining
 -> Continued Pretraining
 -> SFT
 -> Preference stage
 -> Evaluation
```

小型Task Modelなら:

```text
Teacher labeling
 -> Student task training
 -> Calibration
 -> Evaluation
```

## 11. Training MethodはCapabilityではない

同じRuntime Capabilityを異なるTraining Methodで作れます。

例:

```text
model.text.classify

Train from scratch
Full FT
Head-only
Adapter
Distillation
```

Flow/Coreはどの方法で作ったかを前提にしません。
