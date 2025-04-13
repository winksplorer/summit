// summit frontend/js/term.js - handles terminal

import { hterm } from './hterm.min.js';

const t = new hterm.Terminal();

t.onTerminalReady = function() {
    const io = t.io.push()
    const socket = new WebSocket("wss://" + location.host + "/api/pty");
    socket.onopen = (event) => socket.send(JSON.stringify({ type: "resize", cols: t.screenSize.width, rows: t.screenSize.height }));

    socket.onmessage = (event) => t.io.print(event.data.replace(/\n/g, "\r\n"));
    io.onVTKeystroke = (str) => socket.send(str);
    io.sendString = (str) => socket.send(str);
    io.onTerminalResize = (columns, rows) => socket.send(JSON.stringify({ type: "resize", cols: columns, rows: rows }));
};

t.decorate(document.getElementById('terminal'));
t.installKeyboard();