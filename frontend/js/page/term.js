// summit frontend/js/term.js - handles terminal


(function(){
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
    
    _.page.fit = fit = new FitAddon.FitAddon();
    
    _.page.terminalResizeOberserver = terminalResizeOberserver = new ResizeObserver(function(entries) {
        try {
            fit && fit.fit();
        } catch (err) {
            console.log(err);
        }
    });
    
    _.onReady(function(){
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
            t.write(event.data);
        };
    
        t.onData(data => {
            socket.send(data);
        });
    
        t.onResize(({ cols, rows }) => {
            socket.send(JSON.stringify({ type: "resize", cols, rows }));
        });
    
        terminalResizeOberserver.observe(document.getElementById('terminal'))
    });
})();