import { hterm } from './hterm.js'; // Ensure the correct path

const t = new hterm.Terminal();

t.onTerminalReady = function() {
    const io = t.io.push()
    const socket = new WebSocket("wss://" + location.host + "/api/pty");

    socket.onmessage = function (event) {
        t.io.print(event.data.replace(/\n/g, "\r\n"));
    };

    io.onVTKeystroke = (str) => {
        socket.send(str);
        t.io.print(str.replace("", "\b \b"));
    };

    io.sendString = (str) => {
        socket.send(str);
        t.io.print(str);
    };

    io.onTerminalResize = (columns, rows) => {
        socket.send(JSON.stringify({ type: "resize", cols: columns, rows: rows }));
    };
};

t.decorate(document.querySelector('#terminal'));
t.installKeyboard();