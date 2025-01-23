# summit
summit is an "all-in-one" web UI intended for managing Alpine Linux servers (but supports other, untested distributions).

## Code structure

### backend

The backend server code. It serves the frontend files, and provides a simple API using http endpoints. Written in Go.

### frontend

The frontend web UI code. Written in HTML, vanilla CSS, and vanilla JS. I'm sorry.

### sea.c

The self-extracting archive code so that the entire server is distributed as one file. Written in C.

## Building

### Alpine

```sh
apk add go make clang binutils libarchive-dev linux-pam-dev git \
    && git clone https://github.com/winksplorer/summit
    && cd summit
    && make
```

### Debian

```sh
apt install golang-go make clang binutils libarchive-dev libpam0g-dev git \
    && git clone https://github.com/winksplorer/summit
    && cd summit
    && make
```

summit will probably build & run on other distributions, but I haven't tried.