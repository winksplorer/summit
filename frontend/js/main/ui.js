// summit frontend/js/main/ui.js - handles ui systems

// updates navbar items
_.ui.updateNavItems = () => {
    // empty the group first
    navItemsEl = $('navbar-items');
    navItemsEl.innerHTML = '';
    
    // then add pages
    for(const page of _CONFIG.ui?.pages || []) {
        navItemsEl.appendChild(Object.assign(document.createElement("a"), {
            href: `${page}.html`,
            textContent: page,
            className: location.pathname.split("/").pop() === `${page}.html` ? "current" : "" // if entry is the current page then highlight it
        }));
    };
}

// dispatches ui message
_.ui.dispatchMsg = (title, subtitle) => {
    $('messageTitle').textContent = title;
    $('messageSubtitle').textContent = subtitle;
    $('message').style.display = 'flex';
}

// handles stats changes
_.ui.updateStats = (data) => {
    if (_.isColdEntry && !_.ui.odometerInitialized) Odometer.init();

    $('stats-cpu').textContent = data.cpuUsage;
    $('stats-memused').textContent = data.memUsage;
    $('stats-memrest').textContent = `${data.memUsageUnit}/${data.memTotal}`;

    if (!_.isColdEntry && !_.ui.odometerInitialized) Odometer.init();
    if (!_.ui.odometerInitialized) _.ui.odometerInitialized = true;
}

// toggles compact navbar
_.ui.compactNav = (enabled) => {
    $('nav-right').style.flexDirection = enabled ? "row" : "column";
    document.querySelector('.hostname').style.fontSize = enabled ? "1.5em" : "2em";
}

// sets ui scale
_.ui.setScale = (value) => {
    document.body.style.fontSize = `${value}em`;
}

_.onReady(() => {
    // inital navbar fill
    _.ui.updateNavItems();

    // message dismiss button will close message
    $('messageDismiss').addEventListener('click', () => $('message').style.display = 'none');

    // config shit
    _.ui.compactNav(_CONFIG.ui?.compactNavbar || false);
    _.ui.setScale(_CONFIG.ui?.scale || 1.5);
});