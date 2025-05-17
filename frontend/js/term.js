// summit frontend/js/term.js - handles terminal

const t = new Terminal({
    fontFamily: '"Iosevka NF", "Courier New", monospace',
    cursorBlink: true,
    theme: {
        background: '#101010',
        foreground: '#eeeef6'
    }
});

const fit = new FitAddon.FitAddon();
t.loadAddon(fit);

t.open(document.getElementById('terminal'));
fit.fit();
const socket = new WebSocket("wss://" + location.host + "/api/pty");

socket.onopen = () => {
    socket.send(JSON.stringify({
        type: "resize",
        cols: t.cols,
        rows: t.rows
    }));
};

socket.onmessage = event => {
    t.write(event.data.replace(/\n/g, '\r\n'));
};

t.onData(data => {
    socket.send(data);
});

t.onResize(({ cols, rows }) => {
    socket.send(JSON.stringify({ type: "resize", cols, rows }));
});