// summit frontend/js/term.js - handles terminal


(function(){
    // init terminal object
    _.page.t = t = new Terminal({
        fontFamily: '"Courier New", monospace',
        fontSize: 14,
        cursorBlink: true,
        convertEol: true,
        theme: {
            background: '#101010',
            foreground: '#eeeef6'
        }
    });
    
    // init terminal fit addon (dynamically resize based on container size)
    _.page.fit = fit = new FitAddon.FitAddon();
    
    // init reszie observer
    _.page.terminalResizeObserver = terminalResizeObserver = new ResizeObserver(function(entries) {
        try {
            fit && fit.fit();
        } catch (err) {
            console.log(err);
        }
    });
    
    _.onReady(function(){
        // init terminal and websocket connection
        t.loadAddon(fit);
        t.open(document.getElementById('terminal'));
        fit.fit();
        const socket = new WebSocket("wss://" + location.host + "/api/pty");
    
        // send initial resize
        socket.onopen = () =>
            socket.send(msgpack.serialize({
                type: "resize",
                cols: t.cols,
                rows: t.rows
            }));
    
        // pty -> terminal
        socket.onmessage = event => t.write(event.data);
    
        // terminal -> pty
        t.onData(data => socket.send(data));
    
        // handle resize
        t.onResize(({ cols, rows }) => {
            socket.send(msgpack.serialize({
                type: "resize",
                cols: cols,
                rows: rows
            }));
        });
        
        terminalResizeObserver.observe(document.getElementById('terminal'));
    });
})();
