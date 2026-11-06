# 言語処理手法カタログ

Status: **draft**

Open LLM LearnerはLLM/Transformerだけを言語処理手法として扱いません。

Taskに最も適した方法を選べるよう、複数の方式を同じModel Builderから利用できる設計にします。

## 1. Rule / Template
- keyword rules
- regular expressions
- dictionary lookup
- finite-state rules
- templates
- deterministic parser
- AST rewrite

用途: fixed-format extraction、simple routing、code/config edits、fallback responses。

## 2. Lexical Search / Retrieval
- exact match
- inverted index
- TF-IDF
- BM25-like ranking
- n-gram search

用途: Help search、document lookup、candidate generation、FAQ。

Generatorが不要なら検索結果をそのまま返せます。

## 3. Statistical / Classical ML
- logistic regression
- linear SVM
- Naive Bayes
- decision tree / boosting where appropriate

Feature例: bag-of-words、n-gram、TF-IDF、hand-crafted language features。

用途: classification、intent detection、routing、filter。

## 4. Sequence Tagging
- HMM
- CRF
- neural encoder + CRF

用途: token labeling、entity extraction、field/span extraction。

## 5. N-gram Language Model

用途: autocomplete、constrained local text prediction、baseline、tiny-domain generation。

## 6. CNN for Text

Architecture: token/character representation上の1D convolution + pooling + task head。

用途: classification、short-text intent、character-level pattern tasks。

## 7. RNN / LSTM / GRU

用途: sequence classification、sequence generation、tagging、compact streaming inference。

Transformerとは異なるMemory/Latency特性を持ちます。

## 8. Encoder-Decoder RNN

用途: short transformation、translation-like tasks、constrained generation。

## 9. Transformer Encoder

用途: classification、extraction、ranking、embedding、judgment。

生成LLM以外の重要な主力Recipeです。

## 10. Transformer Encoder-Decoder

用途: rewrite、translation、code/text transform、structured generation。

## 11. Decoder-only Transformer

用途: open-ended generation、chat、code generation、LLM。

## 12. State Space Model

用途: language modeling、long/streaming sequence processing、stateful inference。

特定実装名はCoreへ固定しません。

## 13. Hybrid Architecture
- attention + recurrent
- attention + state space
- convolution + attention
- neural model + CRF

## 14. Embedding + Retrieval

```text
text
 -> embedding model
 -> vector search
 -> result
```

Taskによっては生成Model不要です。

## 15. Ranker / Reranker

```text
cheap retrieval
 -> candidate set
 -> small ranker
 -> best result
```

ゲームHelp、検索、Agent tool selection等に使えます。

## 16. Judgment / Scoring

自由文を生成せずchoice、boolean、score、probabilityを返します。

## 17. Distilled Student

Strong Teacherからclassical model、small encoder、RNN、SSM、compact decoder等へTask能力を移せます。

## 18. Composite Processing

一つの手法で全部を担当する必要はありません。

```text
Regex
 -> Classifier
 -> BM25 Retrieval
 -> Ranker
 -> Template
```

でもAI Packageとして成立します。

## 19. Recommendation principle

Taskに対して小さく単純な方式から検討します。

```text
Rule / Search
 -> Classical ML
 -> Small neural model
 -> Transformer
 -> LLM
```

ただしQualityや柔軟性要求によって途中を飛ばして構いません。

## 20. Official support

設計・Schema・UIは最初から特定Familyに閉じません。

少なくともRule/Template、Retrieval、Classical classifier、Transformer task model、non-Transformer neural model、LLMの複数経路をE2Eで証明することを目標にします。