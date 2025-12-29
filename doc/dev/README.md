# summit Developer Documentation

## Hot reloading the frontend

By default, the backend uses a frontend embedded in the binary, but you can pass a directory as an argument and it will try to use that. This can be used for hot reloads or for completely custom frontends.

If your frontend is at `./frontend`, you simply just do this:

```
summit ./frontend
```

This lets you update a file and see it change when you refresh, rather than having to rebuild the binary.

You can confirm that it's using the correct frontend by checking if something like this appeared in the startup logs:

```
[2025-10-26 15:45:18-0700] Using /home/wink/repos/summit/frontend for frontend directory.
```