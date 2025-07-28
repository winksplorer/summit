// summit frontend/js/main/ui.js - handles ui systems

// updates navbar items
_.ui.updateNavItems = () => {
    // empty the group first
    navItemsEl = $('navbar-items')
    navItemsEl.innerHTML = '';
    
    // request via comm, and go through the response list
    _.comm.request("info.pages").then(data => {
        for(const page of data) {
            navItemsEl.appendChild(Object.assign(document.createElement("a"), {
                href: `${page}.html`,
                textContent: page,
                className: location.pathname.split("/").pop() === `${page}.html` ? "current" : "" // if entry is the current page then highlight it
            }));
        };
    });
}

// dispatches ui message
_.ui.dispatchMsg = (title, subtitle) => {
    $('messageTitle').textContent = title;
    $('messageSubtitle').textContent = subtitle;
    $('message').style.display = 'flex';
}

// handles stats changes
_.ui.updateStats = (data) => {
    if (_.isColdEntry) Odometer.init();

    $('stats-cpu').textContent = data.cpuUsage;
    $('stats-memused').textContent = data.memUsage;
    $('stats-memrest').textContent = `${data.memUsageUnit}/${data.memTotal}`;

    if (!_.isColdEntry) Odometer.init();
}

// toggles compact navbar
_.ui.compactnav = (enabled) => {
    $('nav-right').style.flexDirection = enabled ? "row" : "column";
    document.querySelector('.hostname').style.fontSize = enabled ? "1.5em" : "2em";
}

_.onReady(() => {
    // inital navbar fill, and register message dismiss button
    _.ui.updateNavItems()

    // message dismiss button will close message
    $('messageDismiss').addEventListener('click', () => $('message').style.display = 'none');

    // config shit
    _.comm.request('config.get', {'key': 'ui.compactNavbar'}).then(v => _.ui.compactnav(v.value || false));
});