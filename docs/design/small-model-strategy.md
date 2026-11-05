# 小型用途特化モデル戦略

Status: **agreed direction**

## 1. 目的

Open LLM Learnerは大規模LLMだけでなく、アプリへ常駐できる小型Modelを重要対象にします。

## 2. 小さくする方法

### Taskを狭める
「何でも答える」ではなく:
- Help質問判定;
- 4択Action;
- code condition edit;
- config field edit.

### Contextを狭める
Repository全体ではなく:
- target symbol;
- variables;
- types;
- relevant callers;
- test failure.

### Outputを狭める
全文生成ではなく:
- class;
- span;
- score;
- edit operation.

### Deterministic Toolへ渡す
Modelは判断だけして、変更はAST/Parser/Toolが行う。

### 複数Modelへ分ける
一つの大Modelではなく:
- classifier;
- ranker;
- generator.

## 3. Distillation

大きなTeacherから小さなStudentへ能力を移します。

Teacherは:

- labels;
- rankings;
- edits;
- explanations;
- synthetic data;

を作れます。

本番PackageにTeacherを入れる必要はありません。

## 4. Evaluation

小型ModelはQualityだけでなく:

- startup;
- RAM/VRAM;
- latency;
- throughput;
- artifact size;

を重要Metricにします。

## 5. Runtime

小型ModelならResident実行を優先します。

AirLLM-style streamingは大Modelで有効ですが、小型Modelで無駄に使わない。

## 6. Game/Application Embed

小型ModelのExport先候補:

- ONNX Runtime;
- native runtime;
- GGUF/runtime-specific;
- app-bundled Package.

## 7. Model Builder UX

Recommendationで:

> このTaskは固定候補の分類なので、LLMより小型Encoderを推奨します。

のように提示します。

## 8. Success Criteria

「大Modelより高精度」でなくてもよい。

特定Taskで:

- 必要十分なQuality;
- 小さなMemory;
- 低Latency;
- offline;
- predictable output;

を満たせば成功です。
