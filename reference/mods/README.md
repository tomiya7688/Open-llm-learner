# Official reference MOD packages

These packages exist to prove the public MOD Package and MOD Protocol contracts. They do not use official-only private APIs.

The checked-in package manifests expect a platform-native executable next to `mod.json`:

- `official.echo/0.1.0/oll-mod-echo`
- `official.json/0.1.0/oll-mod-json`
- `official.text/0.1.0/oll-mod-text`

The corresponding Go entry points are:

- `cmd/oll-mod-echo`
- `cmd/oll-mod-json`
- `cmd/oll-mod-text`

For development, build them directly into the package directories. Release packaging may do the same for each target platform before calculating the final package digest.

## Capabilities

### official.echo

- `reference.echo` — returns input unchanged;
- `reference.sleep` — emits a status event, sleeps, supports cancellation.

### official.json

- `reference.json.identity`;
- `reference.json.pick`;
- `reference.json.merge`;
- `reference.json.delegate` — demonstrates MOD -> Core -> capability -> MOD calling.

### official.text

- `reference.text.read` — reads text from the host;
- declares `host.filesystem.read`.

The text MOD is intentionally permission-bearing so v0.1 can exercise Project permission allow/deny behavior. Permission approval is a Core policy gate, not a hostile-code sandbox.
