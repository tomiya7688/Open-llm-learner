# GUI / Shell 機能ガイド

Status: 利用者向けUI/CLIドラフト

この文書は、各機能がGUIまたはShellでどこに現れ、どう動くかを説明します。未実装の操作名は変更される可能性があります。

## ステータス表示
利用者向け画面と文書では、機能を次の4状態で表示します。
- Implemented: 実装済み
- Experimental: 動作するが互換性・品質保証が限定的
- Planned: 仕様やIssueはあるが未実装
- Community: Community MODで提供

## Project
GUI: Home > Projects から New Project / Open Project。開いた後は Weight Builder、Agent Builder、Flow、Models、MODs、Runs、Export、Settingsへ進みます。
CLI予定: ollm project create <dir> / ollm project validate <dir>

## Weight Builder
GUI: Base Model、Dataset、Training method、Recipe、Memoryを選び、Preflight後にStart Training。学習中はStep、Loss、RAM、VRAM、Memory Modeを表示し、Pause / Stopを常時出します。
CLI予定: ollm train <project> --entry train。Memoryは safe-auto / resident / cpu-offload / layer-stream などを指定可能にします。

## Agent Builder
GUI: Preset、Model Roles、Capabilities、Permissionsを選んでAgentを作成します。予定Presetは Coding Agent、Local Lite、WikiSkill Standard、WikiSkill Teacher-to-Local などです。
CLI予定: ollm agent create <project> --preset coding

## Agent実行
GUI: Goalを入力してRun。Current Step、Recent Operations、RAM/VRAM、Runtime Modeを表示します。Pause / Stop / Steerを常時操作できます。
Stop時は、新規Action停止 → active MODへcancel → 子プロセス停止 → 必要なら強制停止 → Safe Task State保存、の順で動きます。
CLI予定: ollm agent run <project> --goal "..."。対話Shellでは /pause /resume /stop /force-stop /steer /state を用意します。

## Flow
GUI: Flow EditorでNodeとEdgeを編集。Node詳細にはInput、Output、Capability、Retry、Timeout、Permission requirementを表示します。
CLI予定: ollm flow list <project> / ollm flow validate <project> <flow> / ollm flow run <project> --entry <name>

## MOD
GUI: Project > MODs でInstalled / Project-local MODを表示し、Version、Provides、Permissions、Dependenciesを確認します。
CLI予定: ollm mod list / ollm mod inspect <id> / ollm mod validate <path>

## Permission
GUI: 必要になった時に、Allow once / Allow for this Project / Deny を表示します。書込みやprocess executionなどの権限はCapability単位で示します。
CLI予定: 非対話実行ではpermission configを事前指定し、未許可Actionは勝手に実行しません。

## Resource Governor / Safe Auto
GUI: Resources > Memory = Safe Auto を標準表示。AdvancedでRAM headroom、VRAM headroom、Disk cache、Heavy task concurrency、Priority、Memory modeを変更できます。
Safe Autoは100%使用を目標にしません。予算に収まらない時は対応Runtimeへ低メモリPlanを要求します。
CLI予定: ollm run <project> --memory safe-auto。詳細指定では --max-ram / --max-vram を使えるようにします。

## Model Runtime
GUI: Models > Runtime でAutoまたは特定Runtimeを選択。Resolved RuntimeとMemory Modeを表示します。
CLI予定: ollm model run <model-ref> --runtime auto

## Export
GUI: Export画面で Weight / Runtime / Package を明確に分けます。

### Weight Export
GUI: Native / Safetensors、LoRA Adapter、Merged Model。
CLI予定: ollm export weight <project> --format safetensors

### Runtime Export
GUI: targetを llama.cpp / LM Studio / Ollama / AirLLM / ONNX Runtime / Community Exporter から選択。必要に応じてQuantizationやTarget versionを選びます。
CLI予定: ollm export runtime <project> --target llama.cpp --quantization Q4_K_M --out <dir>
Exportはファイル生成です。外部アプリへ登録するInstallは別操作にします。

### AI Package Export
GUI: Model、Runtime、Flow、Capabilities、Scripts、Filters、Permission defaultsを確認してBuild Package。Portable / Optimized / Hybrid profileを選択できます。
CLI予定: ollm package build <project> --profile portable

## Model Import
GUI: Model ImportでSafetensors、PyTorch .pt/.pth、ONNXなどを検出し、Artifact ClassとFormatを表示します。
PyTorch checkpointは通常Safe Importを使います。Legacy object checkpointは明示的なTrusted Importに分けます。
ONNXはOpset、External Data、Custom OpsをPreflight表示します。
CLI予定: ollm model import <file> / ollm model preflight <file>

## ONNX Custom Ops
GUI: Custom Op名、Provider、Trust状態を表示し、未解決依存があれば実行前に停止します。
Custom OpはONNX内に任意Pythonが完全内包されるという意味ではなく、Runtime側の明示依存です。

## AirLLM
GUI: Runtime ExportのTargetにAirLLMを用意し、Version、Layer shard preparation、compression/prefetchなどを設定します。AirLLM Runtime AdapterをPackageへ組み込むこともできます。
Open LLM Learner内製Streaming Runtimeとは別実装です。

## WikiSkill Agent
GUI: Agent Builder > PresetからWikiSkill Standard / Local Lite / Teacher-to-Localを選択。Evolution画面ではRaw traces、Wiki、Skills、Validation score、Accepted/Rejected更新を確認します。
CLI予定: ollm agent evolve <project> --preset wikiskill-standard

## Runs / 実行履歴
GUI: Runs画面でStatus、Flow、Models、MODs、Timelineを表示します。Agentではhidden chain-of-thoughtではなく、観測可能なAction、Tool result、State update、Verification resultを記録します。
CLI予定: ollm run list <project> / ollm run show <project> <run-id>

## GUIとCLIの関係
GUIとCLIは同じProject JSON、Flow JSON、Artifactを扱います。片方で作ったProjectをもう片方でも使えることを前提にします。