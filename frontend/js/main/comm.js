// summit frontend/js/comm.js - handles communication

_.comm.socket = new WebSocket("wss://" + location.host + "/api/comm");
_.comm.pending = {};

_.comm.socketReady = null;

_.comm.socketWait = function() {
    if (_.comm.socket.readyState === WebSocket.OPEN) return Promise.resolve()
    if (_.comm.socketReady) return _.comm.socketReady
  
    _.comm.socketReady = new Promise(resolve => {
      _.comm.socket.addEventListener('open', () => resolve(), { once: true })
    })
    return _.comm.socketReady
  }

_.comm.socket.onmessage = async event => {
    let arrayBuffer;

    if (event.data instanceof Blob) arrayBuffer = await event.data.arrayBuffer();
    else if (event.data instanceof ArrayBuffer) arrayBuffer = event.data;
    else {
        console.error('Unexpected data type:', typeof event.data);
        return;
    }

    const msg = msgpack.deserialize(new Uint8Array(arrayBuffer));

    if (msg.id && _.comm.pending[msg.id]) {
        if (msg.error) {
            _.comm.pending[msg.id].reject(`error ${msg.error.code}: ${msg.error.msg}`);
            _.ui.dispatchMsg("error while contacting server", `asked for "${msg.t}," but server responded with "error ${msg.error.code}: ${msg.error.msg}"`);
        }

        _.comm.pending[msg.id].resolve(msg.data);
        delete _.comm.pending[msg.id];
    }

    // handle unsolicited messages
    switch (msg.t) {
        case "auth.status":
            if (!msg.data.authed) window.location.replace('/');
            return;
        case "stat.basic":
            _.onReady(function () {
                document.getElementById("stats-cpu").textContent = msg.data.cpuUsage;
                document.getElementById("stats-memused").textContent = msg.data.memUsage;
                document.getElementById("stats-memrest").textContent = `${msg.data.memUsageUnit}/${msg.data.memTotal}`;
            });
            return;
    }
}

_.comm.socket.onerror = async err => {
    _.ui.dispatchMsg("error while contacting server", `a serious error has occurred. all communication has been halted. check devtools for more info.`);
    for (const id in _.comm.pending) _.comm.pending[id].reject(err);
    Object.keys(_.comm.pending).forEach(id => delete _.comm.pending[id]);
}

_.comm.request = function(t) {
    return _.comm.socketWait().then(() => {
        return new Promise((resolve, reject) => {
            const id = Math.floor(Math.random() * 2**32);
            _.comm.pending[id] = { resolve, reject };
            _.comm.socket.send(msgpack.serialize({ t, id }));
        });
    })
}