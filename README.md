# summit
summit is an "all-in-one" web UI intended for managing Alpine Linux servers (but supports other, untested distributions).

![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/winksplorer/summit)

## TODO

- [X] SEA
- [X] HTTP/2
- [X] TLS
- [X] Login
    - [X] PAM
    - [X] Login page
    - [X] Cookies
    - [X] Admin system
    - [ ] Slight refactoring (use yawst login code instead)
- [ ] Stats
    - [X] Basic numerical stats
    - [ ] Graphing
- [X] WebSocket terminal
- [ ] Settings
    - [ ] Settings page
    - [ ] Settings system
- [ ] Logging page
- [ ] Storage page
- [ ] Networking page
- [ ] Services page
- [ ] Updates page


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
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make tlskey \
    && make
```

### Debian

```sh
apt install golang-go make clang binutils libarchive-dev libpam0g-dev git \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make tlskey \
    && make
```

summit will probably build & run on other distributions, but I haven't tried.