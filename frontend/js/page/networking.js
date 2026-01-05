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
                _.helpers.newEl(
                    'p',
                    '',
                    (`${nic.speed === -1 ? '' : `${nic.speed}Mb/s`} ${nic.duplex} duplex ${nic.virtual ? 'virtual' : ''} NIC`) + (nic.mac ? `\nMAC address: ${nic.mac}` : '') + (nic.ips ? '\nIP addresses:' : '')
                ),
                list,
                _.helpers.newElWithID('p', '', `${nic.name}-stats`)
            );

            frag.appendChild(container);
        }

        $('networking_nics').appendChild(frag);
    });

    // graph
    let rxSeries = new TimeSeries();
    let txSeries = new TimeSeries();

    let chart = new SmoothieChart({
        grid: { strokeStyle: _.helpers.getCSSVar('line'), fillStyle: _.helpers.getCSSVar('s0') },
        tooltip: true,
        fps: 30,
        minValue: 0,
        responsive: true,
        labels: { fillStyle: _.ui.isDarkTheme() ? '#fff' : '#000', fontSize: 15, fontFamily: 'Archivo Narrow' }
    });

    chart.addTimeSeries(rxSeries, { strokeStyle: _.ui.isDarkTheme() ? '#0f0' : '#070' });
    chart.addTimeSeries(txSeries, { strokeStyle: _.ui.isDarkTheme() ? '#f00' : '#700' });
    chart.streamTo($('networking_traffic'), 1000);
    
    // subscribe to net.stats
    _.comm.subscribe('net.stats', null, nics => {
        if (_.page.statsLast.length) {
            let totalRx = 0, totalTx = 0;

            for (const nic of nics) {
                const old = _.page.statsLast.find(x => x.name === nic.name);
                $(`${nic.name}-stats`).textContent = `RX: ${_.helpers.humanReadable(nic.rx_bytes - old.rx_bytes,1,true)}/s - TX: ${_.helpers.humanReadable(nic.tx_bytes - old.tx_bytes,1,true)}/s`;
                totalRx += (nic.rx_bytes - old.rx_bytes) / 1000;
                totalTx += (nic.tx_bytes - old.tx_bytes) / 1000;
            }

            rxSeries.append(Date.now(), totalRx);
            txSeries.append(Date.now(), totalTx);
        }

        _.page.statsLast = nics;
    })
});