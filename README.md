# summit
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/winksplorer/summit)

summit is an "all-in-one" web UI designed to be portable and (somewhat) minimal while still being useful.

Currently, summit is *not* ready for production. Many, *many* things are missing, and there are likely lots of bugs and security issues I've yet to fix.

And yes, I'm aware that the code is awful.

## Features
- All of summit is distributed as one file
- Portable (Tested with Debian, Alpine, and FreeBSD so far)
- Simple, understandable code (I hope)
- Simple, understandable user experience (I also hope)
- As of May 25th, 2025, the entire final executable is 2.9 MB
- PAM-based login system
- xterm.js-based terminal
- HTTP/2 & HTTPS

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
- [ ] Stats
    - [X] Basic numerical stats
    - [X] Implement Odometer
    - [ ] Make Odometer actually fit in the design (requires themes system)
    - [ ] Make stat values persist across pages
- [ ] WebSocket terminal
    - [X] Switch to xterm.js
    - [X] Firefox compatibility
    - [X] Fix the fucking thing
    - [ ] Fix doas
- [X] UI Notifications
- [X] Get rid of the dumb awful edition system
- [X] Implement templates
- [ ] Use a WebSocket connection instead of HTTP endpoints
    - [X] Proof of concept (Stats test)
    - [ ] Fully switch to WebSockets
- [ ] Simplify entire frontend
    - [X] Simplify HTML using templates
    - [ ] Simplify JS
        - [X] Use `window._`
        - [ ] Bundling
        - [ ] Minification
    - [ ] Simplify CSS
- [ ] We need code comments :sob:
    - [X] SEA
    - [ ] Backend
    - [ ] Frontend
- [ ] Settings
    - [ ] Settings page
    - [ ] Settings system
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
- [ ] Work on frontend design
    - [ ] Add some sort of theme system
- [ ] Installer shell script

## Code structure

### backend

The backend server code. It serves the frontend files, and provides a simple API using http endpoints. Written in Go.

### frontend

The frontend web UI code. Written in HTML, vanilla CSS, and vanilla JS. I'm sorry.

### sea.c

The self-extracting archive code so that the entire server is distributed as one file. Written in C.

## Building

All of these commands assume you're running as root.

summit will likely build & run on other systems, but I haven't tested.

### Alpine

> Make sure you have the community repo enabled.

```sh
apk add go make clang binutils libarchive-dev linux-pam-dev git openssl minify \
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
pkg install go gmake binutils libarchive git openssl minify \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && gmake all bsdinstall
    && mkdir -p /etc/ssl/private
```