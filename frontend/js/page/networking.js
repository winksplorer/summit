// summit frontend/js/page/networking.js - handles the networking page

_.page.statsLast = []; // [{name, rx, tx}]

_.onReady(() => {
    // ask for devices
    _.comm.request('net.getnics').then(nics => {
        const frag = document.createDocumentFragment();

        // go through each one
        for (const nic of nics) {
            const container = _.helpers.newEl('section', 'container', '');
            const list = _.helpers.newEl('ul', 'storage_smart', '');

            for (const ip of nic.ips || []) list.appendChild(_.helpers.newEl('li', '', ip));

            // put it all together
            container.append(
                _.helpers.newEl('h3', '', nic.name),
                _.helpers.newEl('p', '', _.page.constructInfoString(nic)),
                list,
                _.helpers.newElWithID('p', '', `${nic.name}-stats`)
            );

            frag.appendChild(container);
        }

        document.querySelector('.grid').appendChild(frag);
    });
    
    // subscribe to net.stats
    _.comm.subscribe('net.stats', null, nics => {
        if (_.page.statsLast.length) {
            for (const nic of nics) {
                const old = _.page.statsLast.find(x => x.name === nic.name);
                $(`${nic.name}-stats`).textContent = `RX: ${nic.rx_bytes - old.rx_bytes} B/s - TX: ${nic.tx_bytes - old.tx_bytes} B/s`
            }
        }

        _.page.statsLast = nics;
    })
});

_.page.constructInfoString = (nic) => {
    const line1 = `${nic.speed} ${nic.duplex} duplex ${nic.virtual ? 'virtual' : ''} NIC`;
    const line2 = nic.mac ? `\nMAC address: ${nic.mac}` : '';
    const line3 = nic.ips ? '\nIP addresses:' : '';

    return line1+line2+line3
}