// summit frontend/js/main/comm.js - handles communication (browser-side)

// variables
_.comm.socket = new WebSocket(`wss://${location.host}/api/comm`); // actual websocket connection
_.comm.pending = {}; // pending requests
_.comm.socketReady = null; // if the socket is ready for data or not

// wait for the socket to be open
_.comm.socketWait = () => {
    if (_.comm.socket.readyState === WebSocket.OPEN) return Promise.resolve()
    if (_.comm.socketReady) return _.comm.socketReady

    _.comm.socketReady = new Promise(resolve => _.comm.socket.addEventListener('open', () => resolve(), { once: true }))
    return _.comm.socketReady
}

_.comm.socket.onmessage = async (event) => {
    // get data in the form of whatever js can understand
    const msg = msgpack.deserialize(await event.data.arrayBuffer());

    // attempt to match id with any pending requests
    if (msg.id && _.comm.pending[msg.id]) {
        if (msg.error) { // if server returned error
            _.comm.pending[msg.id].reject(`error ${msg.error.code}: ${msg.error.msg}`);
            _.ui.dispatchMsg('error while contacting server', `asked for "${msg.t}," but server responded with "error ${msg.error.code}: ${msg.error.msg}"`);
        }

        _.comm.pending[msg.id].resolve(msg.data);
        delete _.comm.pending[msg.id];
    }

    // handle unsolicited messages
    switch (msg.t) {
        case 'auth.status':
            if (!msg.data.authed) window.location.replace('/');
            return;
        case 'stat.basic':
            _.onReady(() => _.ui.updateStats(msg.data));
            return;
    }
}

_.comm.socket.onerror = async (err) => {
    _.ui.dispatchMsg('error while contacting server', 'a serious error has occurred. all communication has been halted. check devtools for more info.');
    for (const id in _.comm.pending) {
        _.comm.pending[id].reject(err);
        delete _.comm.pending[id];
    }
}

_.comm.request = (t, data) => {
    // wait for socket to be open then send data
    return _.comm.socketWait().then(() => {
        return new Promise((resolve, reject) => {
            const id = Math.floor(Math.random() * 2**32);
            _.comm.pending[id] = { resolve, reject };
            _.comm.socket.send(msgpack.serialize({ id, t, data }));
        });
    })
}