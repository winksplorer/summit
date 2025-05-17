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
- As of May 12th, 2025, the entire final executable is 3.1 MB
- PAM-based login system
- hterm-based terminal
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
    - [ ] Make Odometer actually fit in the design
    - [ ] Graphing
- [X] WebSocket terminal
    - [X] Switch to xterm.js
    - [X] Firefox compatibility
    - [ ] Fix the fucking thing
- [X] UI Notifications
- [ ] Get rid of the dumb awful edition system
- [ ] Settings
    - [ ] Settings page
    - [ ] Settings system
- [ ] Work on frontend design
    - [ ] Add some sort of theme system
- [ ] Logging page
- [ ] Storage page
- [ ] Networking page
- [ ] Virtual machines
    - [ ] Virtual machine page
    - [ ] Backend system
- [ ] Services page
- [ ] Updates page
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

```sh
apk add go make clang binutils libarchive-dev linux-pam-dev git openssl \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make all install
```

### Debian

```sh
apt install golang-go make clang binutils libarchive-dev libpam0g-dev git openssl \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && make all install
```

### FreeBSD

```sh
pkg install go gmake binutils libarchive git openssl \
    && git clone https://github.com/winksplorer/summit \
    && cd summit \
    && gmake all bsdinstall
    && mkdir -p /etc/ssl/private
```