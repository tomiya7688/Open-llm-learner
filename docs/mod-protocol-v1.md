# MOD Protocol v1

Status: **baseline specification**

MOD Protocol v1 is the local process protocol between the Go Core and MODs.

## Transport

- Transport: standard input / standard output.
- Encoding: UTF-8.
- Framing: NDJSON; exactly one JSON object per line.
- `stdout` is reserved for protocol messages.
- Human-readable logs must go to `stderr` or be emitted as protocol `event` messages.
- Large payloads should not be embedded in protocol messages; pass references instead.

The Core may terminate a MOD that writes invalid protocol data to stdout.

## Process lifetime

A MOD process is normally long-lived for the duration of a project/run session and may serve multiple invocations.

The Core starts the process, performs the handshake, invokes actions as needed, and finally requests shutdown.

## Handshake

Core -> MOD:

```json
{"type":"hello","protocols":["1.0"],"core":{"version":"0.1.0"}}
```

MOD -> Core:

```json
{"type":"hello","protocol":"1.0","mod":{"id":"official.file","version":"0.1.0"}}
```

The returned MOD id/version must match the loaded package manifest.

If no compatible protocol exists, the MOD exits with a non-zero status after returning an error when possible.

## Invocation

Core -> MOD:

```json
{
  "type":"invoke",
  "id":"inv-001",
  "action":"read_text",
  "input":{"path":"README.md"}
}
```

`id` is unique for an active invocation in the process.

A MOD may execute multiple invocations concurrently if its manifest declares that capability. Otherwise the Core serializes invocations for that MOD.

## Events

MOD -> Core:

```json
{
  "type":"event",
  "id":"inv-001",
  "event":"progress",
  "data":{"current":18,"total":30}
}
```

Common event names include:

- `progress`
- `log`
- `metric`
- `preview`
- `status`

Event payloads are opaque to the Core except for generic display/recording behavior.

Events are telemetry in v1 and are not Flow edge values.

## Result

MOD -> Core:

```json
{
  "type":"result",
  "id":"inv-001",
  "output":{"text":"..."}
}
```

Exactly one terminal message is allowed for an invocation.

## Error

MOD -> Core:

```json
{
  "type":"error",
  "id":"inv-001",
  "error":{
    "code":"FILE_NOT_FOUND",
    "message":"README.md was not found",
    "details":{}
  }
}
```

Error codes are owned by the MOD unless reserved by the protocol.

## Cancellation

Core -> MOD:

```json
{"type":"cancel","id":"inv-001"}
```

A cooperative MOD should stop the invocation and answer:

```json
{"type":"cancelled","id":"inv-001"}
```

If the MOD does not stop within the Core's cancellation grace period, the Core may terminate the process.

Timeout is enforced by the Core and may use the same cancellation path before forced termination.

## MOD-to-Core calls

A MOD may request another capability through the Core. This is the mechanism used by Runner-hosted scripts and composite MODs.

MOD -> Core:

```json
{
  "type":"call",
  "id":"call-001",
  "parent":"inv-001",
  "target":{"capability":"filesystem.read_text"},
  "input":{"path":"src/main.go"}
}
```

The target may alternatively name a specific provider:

```json
{
  "type":"call",
  "id":"call-002",
  "parent":"inv-001",
  "target":{"mod":"official.file","action":"read_text"},
  "input":{"path":"src/main.go"}
}
```

The Core resolves capability bindings, checks project permissions, invokes the target MOD, and returns one of:

```json
{"type":"call_result","id":"call-001","output":{"text":"..."}}
```

or

```json
{"type":"call_error","id":"call-001","error":{"code":"PERMISSION_DENIED","message":"...","details":{}}}
```

A MOD must not bypass the Core when it expects project permission enforcement.

## Shutdown

Core -> MOD:

```json
{"type":"shutdown"}
```

MOD -> Core:

```json
{"type":"shutdown_ack"}
```

The Core may terminate the process if it does not exit after acknowledgement within a grace period.

## Reserved protocol errors

The Core may use these generic error codes:

- `INVALID_REQUEST`
- `PROTOCOL_ERROR`
- `PERMISSION_DENIED`
- `CAPABILITY_NOT_FOUND`
- `ACTION_NOT_FOUND`
- `CANCELLED`
- `TIMEOUT`
- `MOD_TERMINATED`

## Compatibility

Protocol major version changes may be breaking.

Protocol minor version changes must remain backward compatible within the same major version.

The handshake selects a mutually supported version.
