# summit Developer Documentation

## Building the backend without UPX

UPX compresses the final binary by a lot, but it slows down the build process and can hurt debugging sometimes. To build the final binary without UPX compression just run:

```
make frontend backend
```

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