# v0.1 Core Proof sample

This Project exercises the v0.1 Core using only public Project, Flow, MOD Package, and MOD Protocol contracts.

It demonstrates:

- Project-local MOD discovery;
- MOD A -> Core -> MOD B capability calls;
- MOD events and results;
- a reusable `flow.call` subflow;
- branch routing;
- a handled timeout/error route;
- a permission-controlled host text read;
- Project permission grant/revoke persistence.

## Prepare the reference MOD executables

From the repository root:

```sh
go build -o examples/v0.1-core-proof/packages/mods/official.echo/0.1.0/oll-mod-echo ./cmd/oll-mod-echo
go build -o examples/v0.1-core-proof/packages/mods/official.json/0.1.0/oll-mod-json ./cmd/oll-mod-json
go build -o examples/v0.1-core-proof/packages/mods/official.text/0.1.0/oll-mod-text ./cmd/oll-mod-text
```

The executables are intentionally not committed. A packaged release may place the same binaries next to the checked-in `mod.json` manifests.

## Main E2E flow

```sh
go run ./cmd/oll-core flow run \
  --project examples/v0.1-core-proof \
  --entry default \
  --input '{"message":"hello"}'
```

The expected successful path is:

```text
main.start
  -> official.json/delegate
     -> Core capability resolution
        -> official.echo/echo
  -> flow.call flows/wrap.flow.json
     -> official.echo/echo
  -> branch
  -> success end
```

The run result retains the child Run under the `wrap` node and records parent/child Run IDs.

## Timeout/error route

```sh
go run ./cmd/oll-core flow run \
  --project examples/v0.1-core-proof \
  --entry timeout
```

The `slow` node asks the Echo MOD to sleep for 250 ms but has a 25 ms node timeout. Its structured `TIMEOUT` error is routed through `slow.error` to the `recovered` end node, so the Flow itself completes successfully while the node remains `timed_out`.

## Permission allow/deny

The checked-in Project grants `official.text` the permission declared by its manifest:

```text
host.filesystem.read
```

Run the read flow:

```sh
go run ./cmd/oll-core flow run \
  --project examples/v0.1-core-proof \
  --entry read
```

Then revoke the grant:

```sh
go run ./cmd/oll-core project permission revoke \
  examples/v0.1-core-proof \
  official.text \
  host.filesystem.read
```

A later Flow run fails before the permission-requiring MOD is launched. Restore it with:

```sh
go run ./cmd/oll-core project permission grant \
  examples/v0.1-core-proof \
  official.text \
  host.filesystem.read
```

The grant state is persisted in `project.json`.

> Note: v0.1 permission checking is a Core policy gate, not a hostile-code sandbox. A granted process MOD still executes with the host OS permissions of the Open LLM Learner process.
