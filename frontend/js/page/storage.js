// summit frontend/js/page/logging.js - handles the storage page

_.onReady(() => {
    // ask for devices
    _.comm.request('storage.getdevs').then(devs => {
        const frag = document.createDocumentFragment();

        // go through each one
        for (const dev of devs) {
            const container = _.helpers.newEl('section', 'container storage_dev-container', '');

            // parent section
            const parent = _.helpers.newEl('div', 'storage_parent-dev', '');
            parent.appendChild(_.helpers.newEl(
                'p',
                '',
                `${dev.model} (${dev.serial})\n${dev.size} ${dev.controller} ${dev.type}\n${dev.partitions?.length} partitions`
            ));

            // partition section
            const parts = _.helpers.newEl('table', '', '');
            const header = parts.insertRow(0);

            for (const h of ['name', 'label', 'size', 'type', 'mountpoint', 'readonly?', 'uuid'])
                header.insertCell().outerHTML = `<th>${h}</th>`;

            if (dev.partitions) {
                for (const part of dev.partitions) {
                    const row = parts.insertRow();
                    for (const x of Object.keys(part)) 
                        row.insertCell().textContent = part[x];
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