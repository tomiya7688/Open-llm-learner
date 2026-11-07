# 手法選択マトリクス

Status: **draft**

## 1. 目的

初心者向けUIでは「Transformerを使う」「CRFを使う」ではなく、利用者が選んだ自然な目的文からTaskと実装手法候補へ展開します。

流れ:

```text
「何をしたいか」
   ↓
Task / Task Composition
   ↓
Method Candidate
   ↓
Model Recipe / Rule / Retrieval
   ↓
Training / Build / Evaluation
```

## 2. 単一Taskの例

| 利用者向け表現 | Task | 軽量候補 | Neural候補 | 生成候補 |
| --- | --- | --- | --- | --- |
| 種類を判定する | text_classification | Rule / Logistic Regression / SVM | CNN / RNN / Encoder | Small LLM |
| 必要な情報を取り出す | extraction | Regex / CRF | Encoder + Span Head | Structured LLM |
| 候補から選ぶ | decision / ranking | Rule / Linear Ranker | Cross Encoder / Judgment Model | LLM |
| 似ているものを探す | retrieval / embedding | TF-IDF / BM25 | Embedding Model | LLM補助 |
| 文章を書き換える | text_transform | Template / Rule | RNN Seq2Seq / Encoder-Decoder | Decoder LLM |
| コードを直す | code_transform | AST Rewrite | Structured Edit Model / Encoder-Decoder | Coding LLM |
| 質問に答える | retrieval + response | Search + Template | Ranker + Small Generator | LLM / RAG |
| 文章を作る | text_generation | Template / N-gram | RNN / SSM / Small Decoder | LLM |

## 3. 複合Taskの例

### 「質問か判定し、Helpから答えを探す」

```text
text_classification
 -> retrieval / ranking
```

候補:

- Logistic Regression + BM25;
- Tiny Encoder + Cross Encoder;
- Judgment Model + Embedding Search.

### 「Helpから答えを探し、おじいさん風に返す」

```text
retrieval
 -> text_generation / template_fill
```

候補:

- BM25 + Template;
- Embedding Search + Small Decoder;
- Retrieval + LLM.

### 「コードの対象を見つけ、条件だけを書き換える」

```text
extraction
 -> code_transform
```

候補:

- Parser + AST Rewrite;
- Symbol Extractor + Structured Edit Model;
- Coding LLM.

## 4. 推薦の基本順序

Taskが同じなら、原則として以下を順に検討します。

```text
Rule / Template
 -> Search / Retrieval
 -> Classical ML
 -> Small Neural Model
 -> Transformer / SSM / RNN
 -> LLM
```

ただし順序は固定ではありません。

利用者が以下を優先した場合は候補順を変えます。

- 軽さ;
- 速度;
- 精度;
- 柔軟性;
- Offline;
- 学習データ量;
- 対応Runtime;
- 出力の決定性.

## 5. 「学習しない」を正式候補にする

Taskに対して学習が不要なら、Model Builderは明示的に:

> 学習モデルを使わずに実現できます。

と提案します。

例:

- Keyword Search;
- Regex;
- Template;
- AST Rewrite;
- BM25.

これはFallbackではなく正式なRecipe候補です。

## 6. 手法候補の宣言

各Recipe / Method Providerは少なくとも:

- 対応Task;
- 対応Dataset Contract;
- Training/Fitが必要か;
- Runtime Capability;
- Resource characteristics;
- Export target;
- Beginner向け説明文;

を宣言します。

概念例:

```json
{
  "id": "official.method.linear-text-classifier",
  "tasks": ["text_classification"],
  "training_required": true,
  "capability": "model.text.classify",
  "beginner_labels": {
    "ja": "軽くて速い分類器",
    "en": "Lightweight text classifier"
  }
}
```

## 7. Beginner UI

初心者には方式名より説明を先に表示します。

```text
おすすめ

● 軽くて速い分類器
  CPUだけでも動かしやすい
  固定カテゴリ判定向け

○ 小型ニューラル分類器
  表現の幅が広い
  少し学習コストが高い

○ 小型言語モデル
  柔軟だが重い
```

Advancedで:

- Logistic Regression;
- GRU;
- Transformer Encoder;

等の実方式名を確認できます。

## 8. Community Method

Community MODは新しいTaskだけでなく、新しいMethod Candidateも追加できます。

Core変更は不要です。

## 9. Recommendation Engine

初期実装はRule-basedでよい。

入力:

- Task;
- Dataset size;
- output type;
- resource budget;
- latency target;
- offline requirement;
- deterministic requirement.

出力:

- ranked method candidates;
- short explanation;
- required dataset/training;
- estimated resource class.

Recommendation自体を将来Model化しても、Coreへ固定しません。

## 10. 設計原則

> 利用者は「何をするか」を選び、システムは「どう実現するか」の候補を出す。

この境界を維持します。
