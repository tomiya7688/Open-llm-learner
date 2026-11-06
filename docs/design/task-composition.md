# Task Composition

Status: **draft**

## 1. 「〇〇し、〇〇する」を正式概念にする

利用者向けUIで複数Taskを自然に組み合わせられるようにします。

例:
```text
種類を判定し、候補を並べ替える
```

内部:
```text
text_classification
 -> ranking
```

## 2. CompositionはFlowへ展開

Task Composition自体を巨大な新Model Typeにはしません。

標準ではFlowへ展開します。

各StepはRule、Search、Model、Toolのどれでも構いません。

## 3. Composition Pattern

Sequential:
```text
Aし、Bする
A -> B
```

Conditional:
```text
AならBし、それ以外ならCする
```

Retrieve + Generate:
```text
関連情報を探し、それを使って返事を作る
```

Extract + Format:
```text
必要な項目を取り出し、JSONにする
```

## 4. Beginner UI

選択式で組み立てられます。

```text
[質問か判定する]
        ↓
[Helpから探す]
        ↓
[おじいさん風に返す]
```

画面上では:

> 質問か判定し、Helpから探し、おじいさん風に返す

とまとめて表示できます。

## 5. Recommendation per step

各Stepごとに実装方式を推薦します。

例:
- 判定 -> Linear classifier
- 探す -> BM25
- 返事 -> Template

結果として学習Modelが1つも不要なFlowになることもあります。

## 6. Optimize Composition

後からStep統合、Model蒸留、Multi-task model化しても構いません。

最初は分解したFlowで正しさを優先できます。

## 7. Package

CompositionはAI Packageとして保存できます。

Task-only Model ProjectとAI Packageの橋渡しになります。