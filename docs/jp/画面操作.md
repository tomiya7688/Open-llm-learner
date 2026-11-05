# GUIガイド

Status: **利用者向けドラフト**

Open LLM LearnerのGUIは、Shell/CLIで行える操作をボタン・フォーム・状態表示にした操作パネルです。GUI専用の実行方式はありません。

## Projects
New Project / Open ProjectからProjectを作成・開きます。Projectを開くとWeight Builder、Agent Builder、Flow、Models、MODs、Runs、Export、Settingsへ進みます。

## Weight Builder
Base Model、Dataset、Training method、Recipe、Memoryを選び、Preflight後に学習を開始します。

表示例:
- Full Fine-tuning / LoRA / Streamed LoRA
- Memory: Safe Auto
- Step / Loss / RAM / VRAM / Memory Mode
- Pause / Stop

## Agent Builder
Preset、Model Roles、Capabilities、Permissionsを選んでAgentを作成します。

予定Preset:
- Coding Agent
- Local Lite Agent
- Decision + Generator Agent
- WikiSkill Standard
- WikiSkill Local Lite
- WikiSkill Teacher-to-Local

## Agent実行
Goalを一度入力すると、Agentが複数Stepを自律的に進めます。Current Step、Recent Operations、RAM/VRAM、Runtime Modeを表示します。

実行中は常に:
- Pause
- Stop
- Steer
を操作できます。

### Emergency Stop
Stopを押すと、新しいActionを止め、実行中MODへCancelし、子プロセスへ停止要求を出します。必要ならProcess Treeを強制停止し、最後のSafe Task Stateを保存します。モデルの返答待ちにはなりません。

## Flow Editor
NodeとEdgeでFlowを編集します。Node詳細ではInput、Output、Capability、Retry、Timeout、Permission requirementなどを設定します。保存形式はJSONです。

## MODs
Installed MODとProject-local MODを表示します。Version、Provides、Permissions、Dependencies、Package locationを確認できます。

## Models
Safetensors、PyTorch .pt/.pth、ONNX、GGUF、AirLLM prepared artifactなどを管理します。Canonical WeightとRuntime向け派生Artifactは区別して表示します。

## Runtime
Autoまたは特定Runtimeを選びます。Resolved RuntimeとMemory Modeを表示します。候補には公式Runtime、ONNX Runtime、AirLLM、Community Runtimeなどがあります。

## Resources / Safe Auto
Safe AutoはRAM/VRAMを100%使い切ることを目標にしません。AdvancedではRAM headroom、VRAM headroom、Disk cache、Heavy task concurrency、Priority、Memory modeなどを変更できます。

対応Runtimeは必要に応じて Resident → CPU Offload → Layer Streaming → Expert Streaming のように低メモリ方式へ切り替えます。

## Permissions
filesystem read/write、network、process executionなどの権限要求を、Allow once / Allow for this Project / Denyで処理します。

## Model Import
Safetensorsは主要なWeight形式として扱います。PyTorch .pt/.pthは通常weights-onlyの安全側Importを使います。ONNXはOpset、External Data、Custom OpsをPreflight表示します。

## Export
Exportは3種類に分けます。

### Weight Export
Safetensors、LoRA Adapter、Merged Modelなど、重みそのものを出力します。

### Runtime Export
llama.cpp、LM Studio、Ollama、AirLLM、ONNX Runtime、Community Exporterなど、外部アプリが期待する形式・構成へ変換します。

### Package Export
Model、Runtime、Flow、Tools、Scripts、Filters、Permission defaultsなどをまとめたAI Packageを作ります。Portable / Optimized / Hybrid profileを選べる設計です。

## WikiSkill
WikiSkill系PresetではRaw traces、Wiki、Skills、Validation score、Accepted / Rejected skill updatesを確認できます。

## Runs
Status、Started/Ended、Flow、Models、MODs、Timeline、Tool calls、Results、Errors、Retry、Verificationを確認します。Agentではhidden chain-of-thoughtではなく観測可能なAction、Result、Task Stateを記録します。