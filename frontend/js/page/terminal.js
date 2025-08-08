// summit frontend/js/term.js - handles terminal

// init terminal object
_.page.t = new Terminal({
    fontFamily: '"Courier New", monospace',
    fontSize: 14,
    cursorBlink: true,
    convertEol: true,
    theme: {
        background: '#070707',
        foreground: '#eeeef6'
    }
});

// init terminal fit addon (dynamically resize based on container size)
_.page.fit = new FitAddon.FitAddon();

// init reszie observer
_.page.terminalResizeObserver = new ResizeObserver(() => _.page.fit?.fit());

_.onReady(() => {
    // init terminal and websocket connection
    _.page.t.loadAddon(_.page.fit);
    _.page.t.open($('terminal'));
    _.page.fit.fit();

    const socket = new WebSocket("wss://" + location.host + "/api/pty");

    // send initial resize
    socket.onopen = () =>
        socket.send(msgpack.serialize({
            type: "resize",
            cols: _.page.t.cols,
            rows: _.page.t.rows
        }));

    // pty -> terminal
    socket.onmessage = (event) => _.page.t.write(event.data);

    // terminal -> pty
    _.page.t.onData(data => socket.send(data));

    // handle resize
    _.page.t.onResize(({ cols, rows }) => {
        socket.send(msgpack.serialize({
            type: "resize",
            cols: cols,
            rows: rows
        }));
    });
    
    _.page.terminalResizeObserver.observe($('terminal'));
});