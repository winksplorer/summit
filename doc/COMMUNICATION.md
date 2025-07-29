# Backend <-> Frontend Communication

> [!NOTE]
> All examples will be in JSON, however the actual communication would be done in MessagePack.

## Basic Info

### Protocols

- Network protocol: `WebSocket`
- Message protocol: `MessagePack`

### Cases

The backend can either:

- Wait for a valid request from the frontend
- Push unsolicited messages to the frontend

The frontend can either:

- Send a request to the backend
- Wait for data from the backend

### Data

As previously stated, MessagePack is the data protocol.

A `t` (type) value must be in every message. It's used to say what the purpose of the message is.

`id` is a random uint32 that must be repeated in all response messages. Not required for unsolicited messages from the backend.

## Example cases

### Stats

Every 5 seconds, the server can unexpectedly send stats to the frontend, and the frontend receive and display them.

1. Frontend: (no data)
2. Backend: `{"t": "stat.basic", "data": {"memTotal": "31g", "memUsage": 4.6, "memUsageUnit": "g", "cpuUsage": 13}}`

### Setting Configuration Values

The frontend can send a request by filling out `t` (type), `id`, and nothing else. The server will respond with the same `t` and `id`, plus the requested information. `data` can also be filled out if needed.

1. Frontend: `{"t": "config.set", "id": 123, "data": {"ui.scale": 1.75, "ui.compactNavbar": false}}`
2. Backend: `{"t": "config.set", "id": 123, "data": {}}`

### Errors

Here is what would happen if the frontend requests a message type that doesn't exist:

1. Frontend: `{"t": "foo.bar", "id": 456}`
2. Backend: `{"t": "foo.bar", "id": 456, "error": {"code": 404, "msg": "unknown type"}}`

The frontend (and backend, if it ever requests data from the frontend) will always check every response for `error` data and will do the appropriate handling.

## Message Types

### `stat.basic`

- Basic numerical stats.
- Backend pushes to the frontend every 5 seconds.
- Data: `{"memTotal": "31g", "memUsage": 4.6, "memUsageUnit": "g", "cpuUsage": 13}`.

### `config.set`

- Sets configuration values.
- Request: `{"ui.scale": 1.75, "ui.compactNavbar": false}`
- Response: `{}`