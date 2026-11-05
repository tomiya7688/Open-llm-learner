# 用語集

Status: 利用者向けドラフト

Open LLM Learnerで使う言葉を、内部実装を知らなくても分かるように説明します。

## Project
作業全体を保存する単位です。モデル、学習設定、Flow、MOD、権限、Dataset参照、Scriptなどをまとめます。GUIではProjectとして開き、ShellではProjectディレクトリを指定します。

## Flow
何を、どの順番で動かすかを表す処理の流れです。学習なら Dataset → Train → Test → Export、Agentなら Goal → Model → Action → Tool → Result → Model のようになります。保存形式はJSONです。

## Node
Flowの中の一処理です。モデル実行、ファイル読込、コマンド実行、分岐、LoopなどがNodeになります。

## MOD
実際の処理を担当する追加機能です。CoreはFlowを進め、MODが学習、推論、ファイル操作、コマンド実行などを行います。公式MODもユーザーMODも同じ仕組みです。

## Capability
MODが提供する『できること』の名前です。例: filesystem.read_text、model.text.generate、model.decision.evaluate、evaluation.score。FlowはCapabilityを通じて実装を差し替えられます。

## Runner
ScriptをMODとして動かすためのMODです。Python、Luau、Rubyを予定しています。公式Python機能では利用者にPython環境構築を要求しません。

## Weight / Canonical Weight
モデルの重みです。Canonical WeightはRuntime向け変換前の正本です。Runtime用にGGUFやAirLLM形式へ変換しても、元のWeightは別に保持します。

## LoRA / Adapter
元モデル全体を書き換えず、小さな追加重みだけを学習する方式です。LoRA単体Export、実行時Attach、BaseへのMergeを別々に扱います。

## Runtime
モデルを実際に読み込み推論する仕組みです。公式Runtime、ONNX Runtime、AirLLM Adapterなどが該当します。

## Weight Export
重みそのものを持ち出すExportです。Safetensors、LoRA、Merged Modelなどを対象にし、特定アプリに依存しない形を基本とします。

## Runtime Export
モデルを別アプリが期待する形式やディレクトリ構成へ変換する機能です。例: llama.cpp向けGGUF、LM Studio向け配置、Ollama向けModelfile、AirLLM向けprepared model、ONNX Runtime向けONNX。

## AI Package
モデルだけでなく、Runtime、Flow、Tools、Scripts、Filters、Prompt、権限などをまとめた配布物です。Weightだけ配る場合とは別です。

## Prepared Runtime Artifact
Runtime用に分割・変換した派生成果物です。GGUF、AirLLM用分割データ、Layer Streaming用prepared dataなど。Canonical Weightとは区別します。

## Layer Streaming
モデル全体をVRAMに載せず、必要なLayerを順番に読み込みながら実行する方式です。VRAMを減らせますが、I/Oのため遅くなることがあります。

## Expert Streaming
MoEで、選択されたExpertだけを読み込む方式です。

## Partitioned Streaming Training
学習でもモデル全体を常駐させず、分割単位でForward / Backward / Updateする方式です。LoRA向けとFull Fine-tuning向けは別実装です。

## MoE / Active Parameters
MoEは多数のExpertから一部を使うモデル構造です。Open LLM Learnerでは総パラメータ規模とActiveパラメータ規模を分けて扱います。Active Parametersは保存サイズやRAM/VRAM使用量そのものではありません。

## Model Capability
モデルが何をできるかを示します。主対象はtext-to-textです。text-to-decision、image-to-text、text-to-image、text-to-embeddingなども拡張として扱えます。TransformerかどうかはCoreには関係ありません。

## Decision / Judgment Model
文章生成ではなく、選択肢、真偽、スコアなどを返すモデルです。AgentのAction選択やRoutingに使えます。

## Agent
Goalを受け取り、自分で次の操作を選びながら作業を進めるAIです。Plan / Decide → Tool → Observation → Verifyを繰り返します。

## Agent Preset
Agent作成のひな形です。Coding Agent、Local Lite、WikiSkill系などを予定しています。Presetから作った後は普通のProjectとして編集できます。

## WikiSkill
Agentの経験をRaw execution history、persistent Wiki、executable Skillsに分け、Validationで改善したSkillだけを採用する考え方を取り入れたPresetです。

## Task State
AgentのGoal、Plan、Current Step、Observations、Attempts、Final Resultなどの明示的状態です。Pause / Resumeにも利用します。

## Permission
MODやAgentが必要とする権限です。filesystem read/write、network、process executionなどがあります。MODが宣言し、Coreが許可状態を確認します。

## Emergency Stop
実行中のAgentやFlowを止める機能です。通常Cancelで止まらないMODや子プロセスは、猶予後にProcess Treeごと停止できます。

## Resource Governor
RAM、VRAM、CPUなどを使い切らないように実行を調整するCore側の仕組みです。重いNodeの開始待ち、同時実行制限、低優先度Task退避、Runtimeへの低メモリPlan要求などを行います。

## Memory: Safe Auto
OSや他アプリ用の余裕を残して実行する標準モードです。RAM/VRAM 100%使用を目標にしません。必要ならResident → CPU Offload → Layer Streaming → Expert Streamingのように縮退します。

## Reference
大きなモデルやDatasetをJSONに直接埋めず、場所やArtifact IDだけを渡す仕組みです。

## Safetensors
Tensor Weightを保存する主要形式です。Open LLM LearnerではCanonical Weightの第一候補として扱います。

## PyTorch .pt / .pth
PyTorchのCheckpoint形式です。任意Python Objectを含められる場合があるため、通常Importはweights-onlyの安全側経路を使います。

## ONNX
重みだけでなく計算Graphも持てるExecutable Model形式です。Custom Opを使う場合は、その実装をRuntime依存として明示します。

## GGUF
llama.cpp系Runtimeでよく使われるRuntime向け形式です。Canonical WeightからRuntime Exportで生成する派生Artifactとして扱います。