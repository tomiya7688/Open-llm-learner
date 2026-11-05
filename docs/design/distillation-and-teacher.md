# Teacher / Distillation設計

Status: **draft**

## 1. 強いModelを学習時だけ使う

Open LLM LearnerではTeacher ModelをProduction Runtimeに必須としません。

```text
Strong Teacher
    ↓
labels / edits / rankings / decisions
    ↓
Task Dataset
    ↓
Small Student
    ↓
Production
```

## 2. Teacherの用途

- synthetic label generation;
- instruction generation;
- candidate ranking;
- edit generation;
- negative example generation;
- explanation生成;
- skill generation;
- data filtering;
- evaluation補助.

## 3. Teacher Provider

Teacherは:

- local model;
- remote API;
- Open LLM Learner-trained model;
- external model;
- multiple teacher ensemble;

のどれでも構いません。

Flow/Capabilityで接続します。

## 4. Distillation Type

### Hard Label
Teacherの最終Class/Actionだけを利用。

### Soft Target
確率/score distributionを利用。

### Sequence
Teacher生成Text/CodeをTargetにする。

### Ranking
TeacherがCandidate順位を付ける。

### Structured Edit
TeacherがEdit Operationを生成する。

## 5. Validation

Teacher outputを無条件に正解扱いしません。

可能なら:

- tests;
- schema;
- compiler;
- domain rules;
- multiple teachers;
- human correction;

で検証します。

## 6. Provenance

Synthetic sampleには:

- teacher model reference;
- provider/version;
- generation config;
- source input;
- validation result;

をmetadataとして記録できます。

## 7. Teacher-to-Local

WikiSkill Teacher-to-Localだけでなく、Model Trainingでも同じ思想を使います。

強いModelでData/Skillを作り、本番は小型Local Modelだけにできます。

## 8. Cost Budget

Remote Teacherを使う場合:

- request count;
- token/cost budget;
- retry;
- cache;

をProject設定へ持てます。

## 9. Privacy

DatasetをRemote Teacherへ送る場合はNetwork Permissionだけでなく、GUIで外部送信であることを明示します。
