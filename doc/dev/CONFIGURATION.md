# Configuration

## Communication

Value changes use the existing [comm system](COMMUNICATION.md).

`config.set`

- Sets values.
- Frontend request: `{"ui.scale": 1.75, "ui.compactNavbar": false}`
- Backend response: `{}`

There is no `config.get` however, as configuration values will simply be injected into the frontend using Go templating, so the frontend can read from a `window._CONFIG` constant instead of requesting everything individually. This avoids cases like flashing the wrong theme, and reduces network load.

## Data Storage

This will use JSON. You may hate it (which is understandable, JSON in a daemon is heresy), but JSON lets the backend inject the config into pages very easily.

Simple keys used in `config.set` are converted into the JSON structure by splitting by dot characters. Basically, `ui.scale` turns into `config["ui"]["scale"]` internally.

Each user's config is stored at `~/.config/summit.json`.

## Example Configuration File

[Default configuration](../../frontend/assets/defaultconfig.json)