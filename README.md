# summit
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/winksplorer/summit)

summit is an "all-in-one" server management web UI for *nix systems designed to be portable and (somewhat) minimal while still being useful.

Currently, summit is *not* ready for production. Many, *many* things are missing, and there are likely lots of bugs and security issues I haven't fixed yet.

And yes, I'm aware that the code is awful.

## Features
- All of summit is distributed as one file
- Portable (Tested with Debian, Alpine, and FreeBSD so far)
- Simple, understandable user experience (I also hope)
- As of July 18th, 2025, the entire final executable is ~3.0 MB
- PAM-based login system
- xterm.js-based terminal
- HTTP/2 & HTTPS

## Building

All of these commands assume you're running as root. summit will likely build & run on other systems, but I have not tested.

The final compiled output is simply `./summit`.

### Alpine

> [!NOTE]
> Make sure you have the community repo enabled

```sh
apk add go make clang binutils libarchive-dev linux-pam-dev git openssl minify xz \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make all install
```

### Debian

```sh
apt install golang-go make clang binutils libarchive-dev libpam0g-dev git openssl minify \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make all install
```

### FreeBSD

```sh
pkg install go gmake binutils libarchive git openssl minify gsed \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && gmake all install PREFIX=/usr/local LD=ld.gold SED=gsed \
    && mkdir -p /etc/ssl/private
```

## Code structure

### backend

The backend server code. It serves the frontend, and provides a simple API using [WebSockets + MessagePack](doc/COMMUNICATION.md). Written in Go.

### frontend

The frontend web UI code. Written in HTML, vanilla CSS, and vanilla JS. I'm sorry.

### sea.c

The self-extracting archive code so that the entire server is distributed as one file. Written in C.

## TODO

- [X] SEA
- [X] HTTP/2
- [X] TLS
- [X] Login
    - [X] PAM
    - [X] Login page
    - [X] Cookies
    - [X] Admin system
    - [X] Slight refactoring (use my other login impl instead)
- [X] Stats
    - [X] Basic numerical stats
    - [X] Implement Odometer
    - [X] Make stat values persist across pages
- [X] WebSocket terminal
    - [X] Switch to xterm.js
    - [X] Firefox compatibility
    - [X] Fix the fucking thing
    - [X] Fix doas
- [X] UI Notifications
- [X] Get rid of the dumb awful edition system
- [X] Implement templates
- [X] Use a WebSocket connection instead of HTTP endpoints
    - [X] Proof of concept (Stats test)
    - [X] Fully switch to WebSockets
- [X] Simplify entire frontend
    - [X] Simplify HTML using templates
    - [X] Simplify JS
        - [X] Use `window._`
        - [X] Bundling
        - [X] Minification
    - [X] Simplify CSS
- [X] We need code comments :sob:
    - [X] SEA
    - [X] Backend
    - [X] Frontend
- [ ] Settings
    - [X] Settings page
    - [X] Backend settings system
    - [ ] Frontend settings system
    - [X] Default config
- [ ] Work on frontend design
    - [ ] Hamburger menu navbar
    - [ ] Complete redesign
- [ ] Logging
    - [ ] Backend
        - [ ] Application-specific logs?
        - [ ] Init system logs
        - [ ] Kernel logs
    - [ ] Logging page
- [ ] Storage page
- [ ] Networking page
- [ ] Containers
    - [ ] Backend
        - [ ] Podman support
        - [ ] Docker support
    - [ ] Container page
- [ ] Services
    - [ ] Backend
        - [ ] OpenRC support
        - [ ] systemd support
        - [ ] other init systems
    - [ ] Services page
- [ ] Updates
    - [ ] Backend
        - [ ] apt support
        - [ ] apk support
        - [ ] pkg support
        - [ ] other package managers
    - [ ] Updates page
- [ ] Installer shell script
- [ ] Translations