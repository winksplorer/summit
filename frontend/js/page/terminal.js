// summit frontend/js/term.js - handles terminal

// todo: move this somewhere else
_.page.terminalThemes = {
    base: {
        background: _.helpers.getCSSVar('s0'),
        foreground: _.helpers.getCSSVar('text'),
        cursor: _.helpers.getCSSVar('text'),
        cursorAccent: _.helpers.getCSSVar('s5'),
        selectionBackground: document.documentElement.dataset.theme === 'dark' ? 'rgba(171,178,191,0.3)' : 'rgba(42,43,51,0.3)',
    },
    oneDark: {
        black: '#1e2127',
        red: '#e06c75',
        green: '#98c379',
        yellow: '#d19a66',
        blue: '#61afef',
        magenta: '#c678dd',
        cyan: '#56b6c2',
        white: '#abb2bf',

        brightBlack: '#5c6370',
        brightRed: '#e06c75',
        brightGreen: '#98c379',
        brightYellow: '#d19a66',
        brightBlue: '#61afef',
        brightMagenta: '#c678dd',
        brightCyan: '#56b6c2',
        brightWhite: '#ffffff',
    },
    oneLight: {
        black: '#000000',
        red: '#de3d35',
        green: '#3e953a',
        yellow: '#d2b67b',
        blue: '#2f5af3',
        magenta: '#a00095',
        cyan: '#3e953a',
        white: '#bbbbbb',

        brightBlack: '#000000',
        brightRed: '#de3d35',
        brightGreen: '#3e953a',
        brightYellow: '#d2b67b',
        brightBlue: '#2f5af3',
        brightMagenta: '#a00095',
        brightCyan: '#3e953a',
        brightWhite: '#ffffff',
    }
}

// init terminal object
_.page.t = new Terminal({
    fontFamily: '"Courier New", monospace',
    fontSize: 14,
    cursorBlink: true,
    convertEol: true,
    theme: { ..._.page.terminalThemes.base, ...document.documentElement.dataset.theme === 'dark' ? _.page.terminalThemes.oneDark : _.page.terminalThemes.oneLight }
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