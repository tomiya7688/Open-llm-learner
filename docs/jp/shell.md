# Shell / CLIガイド

Status: **利用者向けドラフト**

この文書はShellからOpen LLM Learnerを操作する方法を説明します。コマンド名は現時点では`ollm`を仮名としており、実装時に変更される可能性があります。

GUIはここで説明する操作を視覚化したパネルです。

## Project
`ollm project create my-project` — Projectを作成します。
`ollm project validate ./my-project` — Project構造と設定を検証します。

## Flow
`ollm flow list ./project` — Flow一覧を表示します。
`ollm flow validate ./project flows/main.flow.json` — Flowを検証します。
`ollm flow run ./project --entry default` — Entry Flowを実行します。

## MOD
`ollm mod list` — 利用可能なMODを表示します。
`ollm mod inspect official.file` — MOD詳細を表示します。
`ollm mod validate ./path/to/mod` — MOD Packageを検証します。

## 学習
`ollm train ./project --entry train` — Projectの学習Flowを実行します。
`ollm train ./project --config configs/training.config.json` — Training設定を指定します。

Memory例:
`ollm train ./project --memory safe-auto`
`ollm train ./project --memory layer-stream`

Safe AutoではRAM/VRAMを100%使い切ることを目標にしません。

## Agent作成
`ollm agent create ./project --preset coding` — Coding Agent PresetからAgentを作ります。
`ollm agent create ./project --preset local-lite` — Local Lite Presetから作ります。

## Agent実行
`ollm agent run ./project --goal "Fix the failing CI tests"` — Goalを渡して自律実行します。

対話Shellで予定する操作:
- `/pause` — 一時停止
- `/resume` — 再開
- `/stop` — 停止
- `/force-stop` — 強制停止
- `/steer <instruction>` — 実行中Agentへ追加指示
- `/state` — Task State表示

Stopはモデル生成中でも受け付けます。

## Resource設定
`ollm run ./project --memory safe-auto` — Safe Autoで実行します。
`ollm run ./project --max-ram 24G --max-vram 10G` — RAM/VRAM上限を指定します。

Resource Governorは設定したheadroomを残し、必要ならoffload/streaming対応Runtimeへ低メモリPlanを要求します。

## Model Import
`ollm model import model.safetensors` — SafetensorsをImportします。
`ollm model import model.pth --safe` — PyTorch checkpointを安全側設定でImportします。
`ollm model import model.onnx` — ONNXをImportします。
`ollm model preflight model.onnx` — Opset、External Data、Custom Opsなどを事前確認します。

## Model Runtime
`ollm model run ./model-ref --runtime auto` — Runtimeを自動選択して実行します。
`ollm model run ./model-ref --runtime official.runtime.onnx` — 特定Runtimeを指定します。

## Weight Export
`ollm export weight ./project --format safetensors` — Portable WeightをExportします。
`ollm export weight ./project --artifact lora` — LoRA ArtifactをExportします。

## Runtime Export
`ollm export runtime ./project --target llama.cpp --quantization Q4_K_M --out ./exports/llama` — llama.cpp向けに出力します。
`ollm export runtime ./project --target ollama` — Ollama向けに出力します。
`ollm export runtime ./project --target airllm` — AirLLM向けprepared artifactを作ります。

Runtime Exportは対象アプリ向けファイルを生成します。対象アプリへ登録するInstallは別操作にします。

## Package Export
`ollm package build ./project --profile portable` — Portable Packageを作ります。
`ollm package build ./project --profile optimized` — 対象Runtime/環境向け最適化Packageを作ります。

## Runs
`ollm run list ./project` — 実行履歴一覧を表示します。
`ollm run show ./project 42` — Run詳細を表示します。

## Permissions
対話実行では必要なPermissionを確認できます。非対話実行ではPermission configを事前指定できる設計にします。未許可Actionは勝手に実行しません。

## 機能状態
ShellのhelpでもImplemented / Experimental / Planned / Communityを表示し、未実装コマンドを実装済みのように扱わない方針です。