// summit frontend/js/term.js - handles terminal

// the basic terminal "theme" that guarantees contrast and matching backgrounds/text with ui
_.page.baseTerminalTheme = {
    background: _.helpers.getCSSVar('s0'),
    foreground: _.helpers.getCSSVar('text'),
    cursor: _.helpers.getCSSVar('text'),
    cursorAccent: _.helpers.getCSSVar('s5'),
    selectionBackground: _.ui.isDarkTheme() ? 'rgba(171,178,191,0.3)' : 'rgba(42,43,51,0.3)',
}

// init terminal object
_.page.t = new Terminal({
    fontFamily: _CONFIG.terminal?.font || 'monospace',
    fontSize: 14,
    cursorBlink: true,
    convertEol: true,
    theme: { ..._.page.baseTerminalTheme, ..._.ui.isDarkTheme() ? _CONFIG.terminal?.darkTheme  : _CONFIG.terminal?.lightTheme }
});

// init terminal fit addon (dynamically resize based on container size)
_.page.fit = new FitAddon.FitAddon();

_.onReady(() => {
    // init terminal and websocket connection
    _.page.t.loadAddon(_.page.fit);
    _.page.t.open($('terminal'));
    _.page.fit.fit();

    const socket = new WebSocket("wss://" + location.host + "/api/pty");

    // send initial resize
    socket.onopen = () =>
        socket.send(msgpack.serialize({
            Type: "resize",
            Cols: _.page.t.cols,
            Rows: _.page.t.rows
        }));

    // pty -> terminal
    socket.onmessage = (event) => _.page.t.write(event.data);
    socket.onclose = (event) => {
        if (event.code !== 1001) _.page.t.write(`\n\x1b[0m--- summit terminal: \x1b[91mconnection closed\x1b[0m (code: ${event.code}, reason: ${event.reason || 'none'}, wasClean: ${event.wasClean}) ---`);
    }

    // terminal -> pty
    _.page.t.onData(data => socket.send(data));

    // handle resize
    _.page.t.onResize(({ cols, rows }) => {
        socket.send(msgpack.serialize({
            Type: "resize",
            Cols: cols,
            Rows: rows
        }));
    });
    
    window.addEventListener('resize', () => _.page.fit?.fit());
});