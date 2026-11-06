# 非Transformer Model Recipe

Status: **draft**

## 1. 目的

Open LLM LearnerがTransformerを選ぶためだけのUIにならないよう、非Transformer RecipeをFirst-classに追加できる設計を明示します。

## 2. RNN / LSTM / GRU Recipe

Task候補: classification、sequence labeling、short generation、streaming text processing。

Config例: vocabulary/tokenizer、embedding size、recurrent type、hidden size、layers、bidirectional、task head。

## 3. Text CNN Recipe

Task: classification、intent detection、character/short-sequence tasks。

Config: kernel sizes、channels、pooling、embedding、head。

## 4. CRF / Structured Sequence Recipe

Task: extraction、sequence labeling。

Standalone statistical CRFとNeural Encoder + CRFを別Recipeにできます。

## 5. N-gram Recipe

Task: tiny autocomplete、domain prediction、baseline generation。

Weight Artifactはneural tensorである必要はありません。

## 6. Linear Text Classifier Recipe

Task: text_classification。

Input: word/character n-gram、TF-IDF、optional external embedding。

Model: logistic regression、linear SVM等。

## 7. State Space Language Model Recipe

Task: text_generation、sequence modeling。

Architecture familyはstate_space / recurrent_state等として表現し、特定実装名をCoreへ固定しません。

## 8. Hybrid Recipe

例:
```text
CNN encoder
 -> recurrent layer
 -> classification head
```

または:
```text
state-space blocks
 -> attention blocks
 -> decoder head
```

## 9. Training API

Training MOD共通APIはTensor neural model、statistical model、index/retrieval modelを扱える必要があります。

Trainingという言葉が不自然な方式はBuild/Fitでも、Model Creation Stageとして扱います。

## 10. Artifact

非Transformer ArtifactもCanonical Artifact、Evaluation metadata、Runtime binding、Exportの共通Lifecycleを使います。

## 11. First reference targets

1. Linear/TF-IDF classifier
2. GRU/LSTM classifier
3. State-space/recurrent language model
4. Transformer classifier
5. LLM

この順でCore/Model BuilderがTransformer固有でないことを早期に検証できます。