# Task / Recipe Versioning and Migration

Status: **draft**

## 1. Task/Recipeは更新される

公式TaskやRecipeは将来改善されます。

既存Projectの再現性を壊さないためVersionを持ちます。

## 2. Identity

例:

```text
official.task.code-transform@1
official.recipe.structured-edit@1.2.0
```

実際のJSON field形式はSchemaに従います。

## 3. Compatible Update

Patch/minor updateで互換と判断できる変更:

-説明;
-optional metadata;
-新しいoptional setting;
-改善されたRecommendation metadata.

Breaking:

- input/output meaning変更;
- target schema変更;
- capability contract変更;
- artifact semantics変更.

## 4. Project Pin

Projectは使用したTask/Recipe versionを固定できます。

最新版へ勝手に移行しません。

## 5. Migration

Migration MOD/Toolは:

- old definition;
- new definition;
- project config;

から変更案を作れます。

適用前にDiffを表示します。

## 6. Artifact Compatibility

Recipe更新後でも既存ArtifactがRuntimeで使える場合があります。

Recipe versionとArtifact format/runtime compatibilityを分離します。

## 7. Community

Community Task/RecipeもVersioning ruleを使えます。

ID namespace collisionを避ける必要があります。
