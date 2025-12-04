// summit frontend/js/page/logging.js - handles the storage page

_.onReady(() => {
    const table = $('storage_devs-table');
    _.comm.request('storage.getdevs').then(devs => {
        for (const dev of Object.keys(devs).sort()) {
            let row = table.insertRow();
            for (const c of Object.keys(devs[dev])) {
                if (['parent', 'children'].includes(c)) continue;
                Object.assign(row.insertCell(), {textContent: devs[dev][c]});
            }
        }
    });
});