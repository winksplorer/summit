// summit frontend/js/page/networking.js - handles the networking page

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
                list
            );

            frag.appendChild(container);
        }

        document.querySelector('.grid').appendChild(frag);
    });
});

_.page.constructInfoString = (nic) => {
    const line1 = `${nic.speed} ${nic.duplex} duplex ${nic.virtual ? 'virtual' : ''} NIC`;
    const line2 = nic.mac ? `\nMAC address: ${nic.mac}` : '';
    const line3 = nic.ips ? '\nIP addresses:' : '';

    return line1+line2+line3
}