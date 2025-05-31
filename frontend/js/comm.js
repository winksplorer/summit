// summit frontend/js/comm.js - handles communication

window.comm = comm = new WebSocket("wss://" + location.host + "/api/comm");

comm.onopen = () => {
    var sourceData = {
        number: 123,
        number2: -0.129,
        text: "A",
        flag: true,
        list: [ 1, 2, 3 ],
        obj: { a: 1, b: "2", c: false, d: { a: 0, b: -1 } },
        time: Date.now()
    };

    var bytes = msgpack.serialize(sourceData);
    comm.send(bytes)
};

comm.onmessage = async event => {
    let arrayBuffer;

    if (event.data instanceof Blob) {
        arrayBuffer = await event.data.arrayBuffer();
    } else if (event.data instanceof ArrayBuffer) {
        arrayBuffer = event.data;
    } else {
        console.error('Unexpected data type:', typeof event.data);
        return;
    }

    var msg = msgpack.deserialize(new Uint8Array(arrayBuffer));

    switch (msg.t) {
        case "stat.basic":
            document.getElementById("stats-cpu").textContent = Math.round(msg.data.cpuUsage);
            document.getElementById("stats-memused").textContent = Math.round(msg.data.memUsage);
            document.getElementById("stats-memrest").textContent = `${msg.data.memUsageUnit}/${msg.data.memTotal}`;
            break;
    }
};
