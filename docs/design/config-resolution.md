# Config Resolution

Status: **draft**

## 1. 設定は複数層から解決する

Model Builderは設定を次の順に解決します。

```text
Task defaults
 -> Recipe defaults
 -> Preset
 -> Hardware/Resource plan
 -> User overrides
 -> Validation
 -> Resolved Config
```

## 2. Resolved Configが再現性の基準

ProjectはPreset名だけを保存しません。

必ず最終的なResolved Configを保存します。

例:

```json
{
  "recipe": "official.recipe.tiny-classifier",
  "preset": "cpu-small",
  "resolved": {
    "layers": 4,
    "hidden_size": 256,
    "max_length": 256
  }
}
```

## 3. Resource Plan

Safe Autoが設定を変更する場合:

-何を変更したか;
-なぜ変更したか;
-推定Resource;
-Performance impact;

を表示します。

## 4. User Override

User overrideは明示的に最後に適用します。

ただし安全制限と矛盾する場合:

- warning;
- reject;
- require explicit unsafe override;

のどれかをMOD/Core policyが返します。

## 5. Unknown Fields

CoreはRecipe固有Config fieldを理解しません。

Schema validationはRecipe/Training MODが行います。

## 6. Version

Resolved Configには:

- Task version;
- Recipe version;
- Training MOD version;
- Runtime/Exporter version;

を記録できます。

## 7. UI

簡単画面:
- Recommended;
- Resource summary;
- important knobs.

Advanced:
- resolved raw config;
- per-Recipe fields.

GUIとShellは同じResolved Configを使います。
