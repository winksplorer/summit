# Backend <-> Frontend Communication

## Basic Info

### Protocols

- Network protocol: `WebSocket`
- Message protocol: `MessagePack`

### Cases

The backend can either:

- Wait for a valid request from the frontend
- Send data whenever it wants

The frontend can either:

- Send a request to the backend
- Wait for data from the backend

### Data

Again, MessagePack is the data protocol.

The `t` (type) "variable" must be in every message, and is used to say what the purpose of the message is.

`id` is a random uint32 that must be repeated if it's a response. Not required for unsolicited messages from backend.

## Example cases

> All examples will be in JSON, however the actual communication would be done in MessagePack.

### Stats

Every 5 seconds, the server can unexpectedly send stats to the frontend, and the frontend receive and display them.

1. Frontend: (no data)
2. Backend: `{'t': 'stat.basic', 'data': {'memTotal': '31g', 'memUsage': 4.6, 'memUsageUnit': 'g', 'cpuUsage': 13}}`

### Requesting buildstring

The frontend can send a request by filling out `t` (type), `id`, and nothing else. The server will respond with the same `t` and `id`, plus the requested information.

1. Frontend: `{'t': 'info.buildString', 'id': 123}`
2. Backend: `{'t': 'info.buildString', 'id': 123, 'data': {'buildString': 'summit v0.3 (built on 2025-May-30)'}}`

### Errors

Let's say that the frontend somehow manages to request a type that doesn't exist. Here's how that would go:

1. Frontend: `{'t': 'foo.bar', 'id': 456}`
2. Backend: `{'t': 'foo.bar', 'id': 456, 'error': {'code': 404, 'msg': 'unknown type'}}`

The frontend (and backend, if it ever requests data from the frontend) will always check every response for `error` data and will do the appropriate handling.

## Currently implemented types

> All representations will be in JSON, however the actual communication would be done in MessagePack.

### `auth.status`

- Authentication status.
- Backend pushes whenever it changes or on connection open.
- Data is in form of `{'authed': true}`.

### `stat.basic`

- Basic numerical stats.
- Backend pushes to the frontend every 5 seconds.
- Data is in form of `{'memTotal': '31g', 'memUsage': 4.6, 'memUsageUnit': 'g', 'cpuUsage': 13}`.

### `info.hostname`

- Server hostname.
- Frontend request to the backend (no data).
- Response data is in form of `{'hostname': 'puter'}`.

### `info.buildString`

- Server build string.
- Frontend request to the backend (no data).
- Response data is in form of `{'buildString': 'summit v0.3 (built on 2025-May-30)'}`.

### `info.pages`

- Page list with navbar names.
- Frontend request to the backend (no data).
- Response data is in form of `{'terminal', 'logging', 'storage', 'networking', 'containers', 'services', 'updates', 'settings'}`.

### `config.set`

- Sets a value in the configuration key/value store.
- Frontend request to the backend, with request data in form of `{'key': 'scale', 'value': 1.75}`
- Reponse data is in form of `{}`

### `config.get`

- Returns a value in the configuration key/value store.
- Frontend request to the backend, with request data in form of `{'key': 'scale'}`
- Response data is in form of `{'value': 1.75}`