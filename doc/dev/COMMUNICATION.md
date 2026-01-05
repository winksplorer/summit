# Backend <-> Frontend Communication

> [!NOTE]
> All examples will be in JSON, however the actual communication is done using MessagePack.

## Basic Info

### Protocols

- Network protocol: WebSocket
- Message protocol: MessagePack

### Cases

The backend can either:

- Wait for a valid request from the frontend
- Push unsolicited messages to the frontend

The frontend can either:

- Send a request to the backend
- Wait for data from the backend

### Data

As previously stated, MessagePack is the data protocol.

A message is made up of 4 parts:

- A `t` (type) value must be in every message. It's used to say what the purpose of the message is.
- `id` is a random uint32 that must be repeated in all response messages.
- `data` represents any attached data in a message.
- `error` represents any error. On successes, this is just a blank struct. On failures, this is filled out.

## Example cases

### Setting Configuration Values

The frontend can send a request by filling out `t`, `id`, and nothing else. The backend will respond with the same `t` and `id`, plus the requested information. `data` can also be filled out if needed.

1. Frontend: `{"t": "config.set", "id": 123, "data": {"ui.scale": 1.75, "ui.compactNavbar": false}}`
2. Backend: `{"t": "config.set", "id": 123, "data": {}}`

### Subscribing to events

The frontend can ask to receive unsolicited messages from the backend. This is done by calling `_.comm.subscribe(t, data, cb)`.

1. Frontend: `{"t": "subscribe", "id": 456, "data": null}`
2. Backend: (no data)

### Receiving events

Once the frontend subscribes to an event, the backend can send unsolicited messages.

1. Frontend: (no data)
2. Backend: `{"t": "stat.basic", "id": 456, "data": {"memTotal": "31g", "memUsage": 4.6, "memUsageUnit": "g", "cpuUsage": 13}}`

### Errors

Here is what would happen if the frontend requests a message type that doesn't exist:

1. Frontend: `{"t": "foo.bar", "id": 789, "data": null}`
2. Backend: `{"t": "foo.bar", "id": 789, "error": {"code": 404, "msg": "unknown type"}}`

The frontend (and backend, if it ever requests data from the frontend) will always check every response for an `error` value.

## Message Types (subscriptions/events)

### `stat.basic`

- Basic numerical stats.
- Backend sends this out every 5 seconds.
- Data: `{"memTotal": "31g", "memUsage": 4.6, "memUsageUnit": "g", "cpuUsage": 13}`

### `net.stat`

- Network statistics, for each NIC.
- Backend sends this out every second.
- Data: `[{"name": "enp5s0", "rx_bytes": 1914090068, "tx_bytes": 62487696}]`

## Message Types (request/response)

### `config.set`

- Sets configuration values.
- Request: `{"ui.scale": 1.75, "ui.compactNavbar": false}`
- Response: `{}`

### `log.read`

- Reads system logs.
- Request: `{"source": "test", "amount": 100, "page": 2}`
- Response: `[{"time": 1755330209, "source": "test", "msg": "Message"}]`

### `storage.getdevs`

- Lists block devices.
- Request: `{}`
- Response:
```json
[
    {
        "name": "nvme0n1",
        "controller": "NVMe",
        "model": "Samsung SSD 980 1TB",
        "removable": false,
        "serial": "<serial number>",
        "size": "932g",
        "type": "SSD",
        "vendor": "unknown",
        "partitions": [
            {
                "fs_label": "unknown",
                "mountpoint": "/boot/efi",
                "name": "nvme0n1p1",
                "ro": false,
                "size": "1g",
                "type": "vfat",
                "uuid": "<fs uuid>"
            }
        ],
        "smart": {
            "ata_realloc_sectors": 0,
            "ata_uncorrectable_errs": 0,
            "available": true,
            "nvme_avail_spare": 100,
            "nvme_crit_warning": 0,
            "nvme_media_errs": 0,
            "nvme_percent_used": 7,
            "nvme_unsafe_shutdowns": 91,
            "power_cycles": 1042,
            "power_on_hours": 2356,
            "read": "26m",
            "temperature": 43,
            "written": "127.5m"
        }
    }
]
```

### `net.getnics`

- Lists network interfaces.
- Request: `{}`
- Response: `[{"name": "enp5s0", "mac": "<mac addr>", "virtual": false, "speed": "1000Mb/s", "duplex": "full", "ips": ["192.168.1.42/24", "<ipv6 addr>/64"]}]`

### `srv.initver`

- Returns the init system version.
- Request: `{}`
- Response: `"systemd 259 (259-1)"`

### `srv.list`

- Lists all services.
- Request: `{}`
- Response: `[{"name": "alsa-restore.service", "description": "Save/Restore Sound Card State", "running": false, "enabled": false}]`

### `srv.start`

- Starts a service.
- Request: `"ssh.service"`
- Response: `{}`

### `srv.stop`

- Stops a service.
- Request: `"ssh.service"`
- Response: `{}`

### `srv.restart`

- Restarts a service.
- Request: `"ssh.service"`
- Response: `{}`

### `srv.enable`

- Enables a service.
- Request: `"ssh.service"`
- Response: `{}`

### `srv.disable`

- Disables a service.
- Request: `"ssh.service"`
- Response: `{}`