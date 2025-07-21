# Configuration (Concept)

## Implementation Details

### Communication

This uses the existing [comm system](COMMUNICATION.md).

`config.set`:

- Sets values.
- Frontend request: `{"ui.scale": 1.75, "ui.compactNavbar": false}`
- Backend response: `{}`

`config.get`:

- Returns a value.
- Frontend request: `{"key": "ui.scale"}`
- Backend response: `{"value": 1.75}`

### Data Storage

This will use JSON. You may hate it (which is understandable, JSON in a daemon is heresy), but JSON is a simple format with typing support and has a nested structure, which is good enough.

Simple keys used in communication are converted into the JSON structure by splitting by dot characters. Basically, `ui.scale` turns into `config["ui"]["scale"]` internally.

Each user's config is stored at `~/.config/summit/config.json`.

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