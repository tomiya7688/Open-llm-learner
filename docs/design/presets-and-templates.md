# Model Preset / Template設計

Status: **draft**

## 1. Presetの目的

利用者がArchitecture詳細を知らなくてもModelを作れるようにします。

PresetはRecipeのDefault Configurationです。

## 2. Preset例

### Tiny Text Classifier
- small encoder;
- CPU-friendly;
- short context.

### Local Code Edit
- structured edit;
- compact context;
- deterministic postprocessor.

### Game Help
- intent + ranking + optional small generator.

### Character LLM
- decoder LLM;
- SFT;
- optional LoRA.

### Local Coding LLM
- code-oriented decoder;
- LoRA;
- Safe Auto.

## 3. Preset != Recipe

Recipe:
> どういうModel/Objectiveか。

Preset:
> そのRecipeの使いやすい初期値。

同じRecipeへ複数Presetを作れます。

## 4. Official / Community

公式Presetも普通のPackage/JSON definitionとして配布します。

Community Presetも同じ形式。

## 5. User Preset

利用者は調整後設定を自分のPresetとして保存できます。

## 6. Hardware-aware Defaults

PresetはResource Preflight後にresolved configを調整可能です。

ただし変更内容は表示・保存します。

## 7. Reproducibility

Preset名だけで再現しません。

Projectにはresolved configも保存します。

Presetが後で更新されても既存Projectの挙動が勝手に変わらないようにします。
