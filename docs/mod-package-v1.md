# MOD Package v1

Status: **baseline specification**

A MOD package is a directory containing a `mod.json` manifest and everything needed to start the MOD.

## Directory layout

Recommended layout:

```text
example-mod/
├─ mod.json
├─ bin/
│  └─ ...
├─ schemas/
│  └─ ...
├─ scripts/
│  └─ ...
└─ assets/
   └─ ...
```

A distributable package should be self-contained for its target platform.

Official packages must not require a user-managed system Python installation.

## Manifest

Example:

```json
{
  "format":"open-llm-learner.mod",
  "format_version":1,
  "id":"official.file",
  "name":"Official File MOD",
  "version":"0.1.0",
  "protocol":"1.0",
  "entry":{
    "command":"bin/file-mod",
    "args":[]
  },
  "concurrency":"serial",
  "permissions":[
    "host.filesystem.read"
  ],
  "provides":{
    "read_text":{
      "capability":"filesystem.read_text",
      "input_schema":"schemas/read-text.input.schema.json",
      "output_schema":"schemas/read-text.output.schema.json"
    }
  },
  "dependencies":[]
}
```

## Required fields

- `format`: must be `open-llm-learner.mod`.
- `format_version`: package manifest format version.
- `id`: stable MOD identifier.
- `version`: semantic version string.
- `protocol`: required MOD protocol version.
- `entry.command`: executable path or a MOD URI.
- `provides`: actions exposed by the MOD.

## Entry command

A normal executable uses a package-relative path:

```json
{"command":"bin/my-mod","args":[]}
```

A Runner-backed script MOD may start through another installed MOD:

```json
{
  "command":"mod://official.runner.python",
  "args":["--script","scripts/main.py"]
}
```

The referenced Runner remains an ordinary MOD package. The script package declares it as a dependency.

The Core only resolves the `mod://` dependency to an installed package entrypoint; it does not need to understand Python, Luau, Ruby, or any other language.

## Actions

Each action may declare:

- a capability name;
- input JSON Schema;
- output JSON Schema;
- optional human-readable title/description.

Capability names are stable semantic identifiers. Multiple MODs may provide the same capability.

A project chooses which provider is bound to a capability.

## Concurrency

`concurrency` is one of:

- `serial`: one active invocation per MOD process;
- `parallel`: multiple active invocations may be sent concurrently.

Default: `serial`.

## Permissions

Permissions are strings declared by the MOD.

Initial reserved host permission namespace:

- `host.filesystem.read`
- `host.filesystem.write`
- `host.network`
- `host.process`
- `host.environment`

Additional permission namespaces may be introduced later.

A MOD may also require use of project capabilities through `dependencies` / project bindings.

The manifest declares requirements; the Core and project decide what is granted.

## Dependencies

Example:

```json
{
  "id":"official.runner.python",
  "version":">=0.1.0 <1.0.0"
}
```

v1 dependency resolution only needs to support installed package dependencies. A registry is not required for Core v0.1.

## Packaging rule

Official MODs and user MODs use the same manifest format and protocol.

There is no private official-only MOD interface.


## Discovery and resolution order

For a project that requests a MOD id/version, the Core resolves packages in this order:

1. project-local packages under `packages/mods/`;
2. the user's application-managed MOD installation directory;
3. official MODs bundled with the Open LLM Learner distribution.

The first exact compatible match is used.

Project-local packages intentionally have highest priority so a project can carry a pinned/private MOD without changing the global installation.

A project may record a package digest. If a digest is present, the resolved package must match it or the Core must report a reproducibility/integrity error rather than silently substituting another package.


## Deterministic v0.1 package resolution

The initial resolver stores installed packages as:

```text
<mods-root>/
└─ <mod-id>/
   └─ <version>/
      └─ mod.json
```

MOD ids used by this resolver are restricted to lowercase ASCII letters, digits, `.`, `_`, and `-`, with the first character required to be alphanumeric.

Package directory versions and manifest versions use SemVer 2.0.0 without a leading `v`.

A project/dependency version selector may be:

- an exact SemVer;
- `*`;
- a whitespace-separated AND expression using `=`, `>`, `>=`, `<`, or `<=`.

Within one search root, the highest compatible version is selected. Search-root priority remains project-local, then user-managed, then official-bundled. If a higher-priority root contains the selected compatible package but that package is malformed or fails integrity validation, resolution fails rather than silently substituting a lower-priority package.

### Package digest

When a project records a package digest, v0.1 calculates it over every regular file in the package directory.

Files are processed by UTF-8 relative path in bytewise sorted order. For each file the SHA-256 stream receives:

1. relative-path byte length as unsigned 64-bit big-endian;
2. UTF-8 relative-path bytes using `/` separators;
3. file-content byte length as unsigned 64-bit big-endian;
4. exact file bytes.

The stored form is `sha256:<lowercase hex>`.

Symbolic links and non-regular files are rejected for digestable MOD packages. This avoids integrity results that depend on external filesystem targets.

### Direct dependencies

The selected package's direct `dependencies` are resolved using the same search-root and version rules.

v0.1 does not recursively resolve dependencies-of-dependencies yet. Missing or malformed direct dependencies fail package resolution.
