// summit frontend/js/page/storage.js - handles the storage page

_.onReady(() => {
    // ask for devices
    _.comm.request('storage.getdevs').then(devs => {
        const frag = document.createDocumentFragment();

        // go through each one
        for (const dev of devs) {
            const container = _.helpers.newEl('section', 'container storage_dev-container', '');

            // parent section
            const parent = _.helpers.newEl('div', 'storage_parent-dev', '');
            parent.appendChild(_.helpers.newEl('p', dev.smart.available ? 'storage_before-smart' : '', _.page.constructParentString(dev)));
            dev.smart.available && parent.appendChild(_.page.getSmartList(dev));

            // partition section
            const parts = _.helpers.newEl('table', '', '');
            const header = parts.insertRow(0);

            for (const h of ['name', 'label', 'size', 'type', 'mountpoint', 'readonly?', 'uuid'])
                header.insertCell().outerHTML = `<th>${h}</th>`;

            if (dev.partitions) {
                for (const part of dev.partitions) {
                    const row = parts.insertRow();
                    for (const i of Object.keys(part)) 
                        row.insertCell().textContent = part[i] !== 'unknown' ? (i === 'size' ? _.helpers.humanReadable(part[i],0) : part[i]) : '';
                }
            }

            // put it all together
            container.append(
                _.helpers.newEl('h3', '', dev.name),
                parent,
                _.helpers.newEl('div', 'storage_container-divider', ''),
                parts
            );

            frag.appendChild(container);
        }

        document.querySelector('main').appendChild(frag);
    });
});

_.page.constructParentString = (dev) => {
    let line1 = `${dev.model} (${_.helpers.shortenText(dev.serial)})\n`;
    if (line1 === 'unknown (unknown)\n') line1 = '';

    const line2 = `${_.helpers.humanReadable(dev.size,0)} ${dev.removable ? 'removable' : ''} ${dev.controller} ${dev.type}\n`;
    const line3 = `${dev.partitions ? dev.partitions?.length : '0'} partitions`;
    const line4 = dev.smart.available ? '\n\nSMART data:' : '';

    return line1+line2+line3+line4
}

_.page.getSmartList = (dev) => {
    const list = _.helpers.newEl('ul', 'storage_smart', '');

    for (const v of [
            `temperature: ${dev.smart.temperature}°C`,
            `blocks read: ${dev.smart.read}`,
            `blocks written: ${dev.smart.written}`,
            `power on hours: ${dev.smart.power_on_hours}`,
            `power cycles: ${dev.smart.power_cycles}`
        ]) list.appendChild(_.helpers.newEl('li', '', v));
    
    if (dev.controller === 'NVMe') {
        for (const v of [
            `critical warning: 0x${dev.smart.nvme_crit_warning.toString(16)}`,
            `available spare: ${dev.smart.nvme_avail_spare}%`,
            `percent used: ${dev.smart.nvme_percent_used}%`,
            `unsafe shutdowns: ${dev.smart.nvme_unsafe_shutdowns}`,
            `media and data integrity errors: ${dev.smart.nvme_media_errs}`
        ]) list.appendChild(_.helpers.newEl('li', '', v));
    } else if (['IDE', 'SCSI'].includes(dev.controller)) {
        for (const v of [
            `reallocated sectors: ${dev.smart.ata_realloc_sectors}`,
            `uncorrectable errors: ${dev.smart.ata_uncorrectable_errs}`,
        ]) list.appendChild(_.helpers.newEl('li', '', v));
    }

    return list
}