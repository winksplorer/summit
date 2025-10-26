# Configuration (Concept)

## Implementation Details

### Communication

Value changes use the existing [comm system](COMMUNICATION.md).

`config.set`

- Sets values.
- Frontend request: `{"ui.scale": 1.75, "ui.compactNavbar": false}`
- Backend response: `{}`

There is no `config.get` however, as configuration values will simply be injected into the frontend using Go templating, so the frontend can read from a variable like `window._CONFIG` instead of requesting everything individually. This avoids flashing as UI is changed in real-time (though not entirely), and reduces network load.

### Data Storage

This will use JSON. You may hate it (which is understandable, JSON in a daemon is heresy), but JSON is a simple format with typing support and has a nested structure, which is good enough.

Simple keys used in communication are converted into the JSON structure by splitting by dot characters. Basically, `ui.scale` turns into `config["ui"]["scale"]` internally.

Each user's config is stored at `~/.config/summit.json`.

## Example Configuration File

```json
{
    "ui": {
        "scale": 1.75,
        "compactNavbar": false
    },
    "stats": {
        "refreshTime": 5000
    }
}
```