# 初心者向け「やりたいこと」選択UI

Status: **agreed direction**

## 1. 最初に専門用語を見せない

Model Builderの入口ではTransformer、Encoder、CRF、SSM、LoRA、Classification Headなどを理解していることを要求しません。

最初に見せるのは「何をしたいか」です。

例:
- 文章を作る
- 質問に答える
- 文章を書き換える
- コードを書く
- コードを直す
- 種類を判定する
- 必要な部分を取り出す
- 候補から選ぶ
- 候補を並べ替える
- 似ているものを探す
- テンプレートを埋める
- 短い返事を作る
- ゲーム内の質問に答える

## 2. 「〇〇する」形式

基本選択肢は動詞で終わる短い文にします。

```text
文章を分類する
コードを修正する
質問に答える
情報を取り出す
候補を並べ替える
```

利用者はArchitectureではなく目的を選びます。

## 3. 「〇〇し、〇〇する」形式

複数の処理を自然文で組み合わせます。

例:
- 質問かどうか判定し、Helpから答えを探す
- Helpから答えを探し、おじいさん風に返事を作る
- コードの対象箇所を見つけ、条件だけを書き換える
- エラーの種類を判定し、修正候補を選ぶ
- 文章から項目を取り出し、JSONに整える

内部ではTask Composition / Flowへ展開します。

## 4. 一文 = 一Modelとは限らない

例:
```text
質問か判定し、Helpから答えを探し、おじいさん風に返事をする
```

内部では:
```text
Intent Classifier
 -> Retriever/Ranker
 -> Small Generator / Template
```

にできます。

逆に一つのMulti-task Modelで実装しても構いません。

## 5. 選択UI

```text
何をしたいですか？

[文章を作る]
[文章を書き換える]
[種類を判定する]
[必要な情報を取り出す]
[候補から選ぶ]
[似た文章を探す]
[コードを書く]
[コードを直す]
[質問に答える]
[自分で組み合わせる]
```

次Stepを追加する場合:
```text
次に何をしますか？

○ ここで終わる
○ 結果を別の処理へ渡す

[候補を並べ替える]
[短い返事を作る]
[決まった形式へ整える]
```

## 6. Beginner / Advanced

Beginnerでは、やりたいこと、Dataset、品質/速さ/軽さの希望、出力先を中心にします。

AdvancedではTask definition、Model Recipe、Architecture、Training Method、Tokenizer、Loss、Resource、Runtimeを直接指定できます。

どちらも同じProject Definitionへ保存します。

## 7. 軽さと柔軟さ

初心者向けにはParameter数より次のように聞けます。

```text
どちらを優先しますか？

● 軽くて速い
○ バランス
○ 柔軟で高性能
```

内部でRule、Retrieval、Classical ML、Tiny neural model、Transformer、LLM等のRecommendationへ反映します。

## 8. 学習しなくていい場合

TaskによってはModel Trainingが不要です。

例えば決まったHelp文から完全一致・部分一致で探すだけならRule/Searchで十分な場合があります。

Model Builderは:

> この仕事は学習モデルを使わなくても実現できます。

と提案できます。

## 9. Localized Task Label

表示文はTask IDとは別にlocalization可能にします。

```json
{
  "task": "text_classification",
  "labels": {
    "ja": "種類を判定する",
    "en": "Classify text"
  }
}
```

日本語UIで英語Task名を理解することを要求しません。

## 10. Composite Action

複合選択は保存可能です。

```json
{
  "steps": [
    {"task": "text_classification"},
    {"task": "ranking"},
    {"task": "text_generation"}
  ]
}
```

## 11. 設計原則

> 初心者は「何をするか」を選ぶ。上級者は「どう実現するか」を選べる。