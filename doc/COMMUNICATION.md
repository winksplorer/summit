# Backend <-> Frontend Communication (Concept)

## Basic Info

### Protocols

- Network protocol: `WebSocket`
- Message protocol: `MessagePack`

### Cases

The backend can either:

- Wait for a valid request from the frontend
- Send data when it likes, bidirectional

The frontend can either:

- Send a request to the backend
- Wait for data from the backend

### Data

Again, MessagePack is the data protocol.

How the frontend and backend would understand what data is for what, is through the `t` (type) "variable" that must be in every single message.

## Example cases

> All examples will be in JSON, however the actual communication would be done in MessagePack.

### Stats

Every 5 seconds, the server can unexpectedly just send the stats to the frontend, and the frontend would handle it.

1. Frontend: (no data)
2. Backend: `{"t": "stat.basic", "data": {"memTotal": "31g", "memUsage": 4.6, "memUsageUnit": "g", "cpuUsage": 13}}`

### Requesting buildstring

The frontend can send a request, by just filling out the `t` (type) and nothing else. The server will respond with `t` and the requested information.

1. Frontend: `{"t": "info.buildString"}`
2. Backend: `{"t": "info.buildString", "data": {"buildString": "summit v0.3 (built on 2025-May-30)"}}`

### Errors

Let's say that the frontend somehow manages to request with a type that doesn't exist. Here's how that would go:

1. Frontend: `{"t": "foo.bar"}`
2. Backend: `{"t": "foo.bar", "error": {"code": 404, "msg": "unknown type"}}`

The frontend (and backend, if it ever requests data from the frontend) will always check every server response for any `error` data, and if it does it will do the appropriate handling.