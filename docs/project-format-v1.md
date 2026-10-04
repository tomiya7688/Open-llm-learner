# Project Format v1

Status: **baseline specification**

A project is a directory. `project.json` is the entry point and canonical saved state.

## Directory convention

```text
my-project/
├─ project.json
├─ flows/
│  └─ *.flow.json
├─ configs/
│  └─ *.config.json
├─ refs/
│  └─ *.ref.json
├─ schemas/
│  └─ *.schema.json
├─ scripts/
│  └─ ...
└─ packages/
   └─ ...
```

Document types are separated by folder and filename suffix.

YAML is not used.

## project.json

Example:

```json
{
  "format":"open-llm-learner.project",
  "format_version":1,
  "id":"example-project",
  "name":"Example Project",

  "entrypoints":{
    "default":"flows/main.flow.json",
    "train":"flows/train.flow.json"
  },

  "mods":[
    {
      "id":"official.file",
      "version":"0.1.0",
      "digest":"sha256:optional-package-digest",
      "permissions":[
        "host.filesystem.read"
      ]
    }
  ],

  "capabilities":{
    "filesystem.read_text":{
      "mod":"official.file",
      "action":"read_text"
    }
  },

  "configs":{
    "training":"configs/training.config.json"
  },

  "refs":{
    "base_model":"refs/base-model.ref.json",
    "dataset":"refs/dataset.ref.json"
  },

  "schemas":{},
  "metadata":{}
}
```

## Reproducibility rule

The project directory is the source of reproducible project state.

Anything required to reconstruct the intended project behavior should be represented directly or by stable reference in the project:

- Flow files;
- selected MOD ids and versions;
- granted permissions;
- capability bindings;
- configurations;
- model/dataset/artifact references;
- scripts;
- seeds and deterministic settings where relevant.

A separate reproducibility database is not required.

## Capability binding

Flows and MODs may ask for a semantic capability rather than a fixed implementation.

The project resolves the capability:

```json
{
  "capabilities":{
    "filesystem.read_text":{
      "mod":"official.file",
      "action":"read_text"
    }
  }
}
```

This allows an official provider to be replaced by a community/private provider without changing the Flow.

## References

Large resources are stored as references rather than embedded in project JSON.

A `.ref.json` document should minimally identify a URI/path and may include integrity metadata:

```json
{
  "format":"open-llm-learner.ref",
  "format_version":1,
  "uri":"file:///models/example",
  "digest":"sha256:optional",
  "metadata":{}
}
```

The Core treats the referenced resource as opaque.

## Portability

Relative project paths are preferred when practical.

External absolute paths and remote URIs are allowed, but naturally reduce portability.
