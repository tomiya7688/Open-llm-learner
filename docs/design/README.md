# 設計書

Status: **設計中**

このディレクトリはOpen LLM Learnerの製品・モデル作成・学習・実行に関する設計の中心です。

Open LLM Learnerは「LLMだけを作るアプリ」ではありません。

中心となる考え方は:

> **利用者が作りたい仕事から逆算して、必要なモデルを作る。**

LLMが必要ならLLMを学習します。
小さなTransformer、分類器、編集モデル、判断モデル、ランキングモデルで十分なら、巨大な生成LLMを使う必要はありません。

## 設計の基本フロー

```text
やらせたい仕事
   ↓
Task Definition
   ↓
Dataset Contract
   ↓
Model Recipe
   ↓
Training Method
   ↓
Evaluation
   ↓
Model Artifact
   ↓
Weight Export / Runtime Export / AI Package
```

## 設計書一覧

- [製品の方向性](product-direction.md)
- [TaskとModel Capability](model-tasks.md)
- [Model Builder](model-builder.md)
- [Model Recipe](model-recipes.md)
- [学習方式](training-methods.md)
- [LLM学習](llm-training.md)
- [用途特化モデルの例](specialized-model-examples.md)

既存のprotocol、Flow、MOD、Package等の仕様書は `docs/` 直下に残します。
この `docs/design/` は「何を作るか、どういう製品にするか」をまとめる設計書です。
