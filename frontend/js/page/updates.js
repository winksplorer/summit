// summit frontend/js/page/updates.js - handles the updates page

_.onReady(() => {
    // container title = package manager version
     _.comm.request('updates.pkgmgr').then(ver => {
        $('updates_header').textContent = ver.version_str;
        $('updates_update-index-cmd').textContent = ver.update_index_cmd;
    });

    // update indexes
    _.comm.request('updates.updateindex').then(() => {
        $('updates_updating-index').style.display = 'none';

        _.comm.request('updates.list').then(pkgs => {
            const frag = document.createDocumentFragment();
            pkgs && pkgs.sort((a, b) => a.name.localeCompare(b.name));

            frag.append(
                _.helpers.newButton('upgrade all', pkgs ? !pkgs.length : true, null),
                _.helpers.newEl('hr')
            );

            for (const pkg of pkgs || []) {
                const inputPair = _.helpers.newEl('div', null, 'input-pair');
                inputPair.append(
                    _.helpers.newEl('span', `${pkg.name} - ${pkg.current_ver} \u2192 ${pkg.new_ver}`),
                    _.helpers.newButton('upgrade', false, null)
                );

                frag.appendChild(inputPair);
                pkgs[pkgs.length-1].name !== pkg.name && frag.appendChild(_.helpers.newEl('hr'));
            }

            $('updates_pkgs').appendChild(frag);
        });
    }).catch(e => _.ui.dispatchMsg("couldn't update package index", e));
})