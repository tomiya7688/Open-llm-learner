# 複数Modelの組み合わせ

Status: **draft**

## 1. 一つのAI = 一つのLLMではない

Open LLM Learnerでは複数の専用ModelをFlowで組み合わせられます。

例:

```text
Intent Classifier
 -> Ranker
 -> Small Generator
```

または:

```text
Judgment Model
 -> Tool
 -> Generative Model
```

## 2. Model Role

AI PackageはModelへRole名を付けられます。

例:

- planner;
- action_selector;
- generator;
- verifier;
- retriever;
- intent_classifier;
- response_styler.

RoleはCore予約語ではありません。

Package/Flow側の意味です。

## 3. Capability Binding

RoleはCapabilityへBindingされます。

```json
{
  "roles": {
    "action_selector": {
      "model": "refs/action-model.ref.json",
      "capability": "model.decision.evaluate"
    },
    "writer": {
      "model": "refs/code-model.ref.json",
      "capability": "model.code.transform"
    }
  }
}
```

## 4. Resource Scheduling

複数Modelを同時常駐させる必要はありません。

Resource Governorは:

- small model resident;
- large model stream;
- sequential load/unload;
- role reuse;

を許可します。

## 5. Composite Model RecipeとAI Flowの境界

一つのWeight Artifact内部で複数Headを持つ構成も可能です。

しかし複数Model ArtifactをFlowで組み合わせる方が交換可能性が高い場合があります。

RecipeはModel内部。
FlowはModel間/Tool間の構成。

この境界を混同しません。

## 6. Distillation

Composite AIをTeacherとして小さい単一Modelへ蒸留しても構いません。

逆に一つの巨大LLMの役割を複数の小型Modelへ分割することも可能です。
