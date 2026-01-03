// summit frontend/js/page/services.js - handles the services page

_.onReady(() => {
    // container title = init version
     _.comm.request('srv.initver').then(ver => $('services_header').textContent = ver);

     // list
    _.comm.request('srv.list').then(srvs => {
        const frag = document.createDocumentFragment();

        srvs.sort((a, b) => a.name.localeCompare(b.name));

        for (const srv of srvs) {
            const inputPair = _.helpers.newEl('div', 'input-pair', '');
            const actions = _.helpers.newElWithID('div', 'services_actions', `${srv.name}_actions`);

            actions.append(
                _.helpers.newButton('start', srv.running, () =>
                    _.comm.request('srv.start', srv.name).then(_.page.updateService(srv.name, true)).catch(e => _.ui.dispatchMsg("couldn't start service", e))
                ),

                _.helpers.newButton('stop', !srv.running, () =>
                    _.comm.request('srv.stop', srv.name).then(_.page.updateService(srv.name, false)).catch(e => _.ui.dispatchMsg("couldn't stop service", e))
                ),

                _.helpers.newButton('restart', !srv.running, () =>
                    _.comm.request('srv.restart', srv.name).then(_.page.updateService(srv.name, true, true)).catch(e => _.ui.dispatchMsg("couldn't restart service", e))
                ),

                _.helpers.newCheckbox(srv.enabled, v =>
                    _.comm.request(v.target.checked ? 'srv.enable' : 'srv.disable', srv.name).catch(e => _.ui.dispatchMsg(`couldn't ${v.target.checked ? 'enable' : 'disable'} service`, e))
                )
            )

            inputPair.append(
                _.helpers.newEl('span', srv.running ? '' : 'services_stopped', `${srv.name.replace('.service', '')}: ${srv.description}`, `${srv.name}_name`),
                actions,
            );

            frag.appendChild(inputPair);
            srvs[srvs.length-1].name !== srv.name && frag.appendChild(_.helpers.newEl('hr', '', ''))
        }

        $('services_srvs').appendChild(frag);
    })
})

_.page.updateService = (srvName, running, restart) => {
    $(`${srvName}_actions`).children[0].disabled = running; // start
    $(`${srvName}_actions`).children[1].disabled = !running; // stop
    $(`${srvName}_actions`).children[2].disabled = !running; // restart
    const nameEl = $(`${srvName}_name`); // name

    if (restart) {
        nameEl.className = 'services_stopped';
        setTimeout(() => nameEl.className = '', 300);
        return
    }
    
    nameEl.className = running ? '' : 'services_stopped'; // name
}