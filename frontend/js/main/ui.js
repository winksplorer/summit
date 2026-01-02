// summit frontend/js/main/ui.js - handles ui systems

// returns true if the current theme is the dark theme
_.ui.isDarkTheme = () => document.documentElement.dataset.theme === 'dark';

// returns true if odometer should init for stats
_.ui.shouldInitOdometerStats = (afterValueSet) =>
    (afterValueSet ? !_.isColdEntry : _.isColdEntry)
    && (!_.ui.odometerInitialized) && (_CONFIG.ui?.odometerStats);

// updates navbar items
_.ui.updateNavItems = () => {
    // empty the group first
    navItemsEl = $('navbar-items');
    navItemsEl.replaceChildren();
    
    // then add pages
    const frag = document.createDocumentFragment();
    for(const page of _CONFIG.ui?.pages || [])
        frag.appendChild(Object.assign(document.createElement("a"), {
            href: page,
            textContent: page,
            className: location.pathname.split("/").pop() === page ? "current" : "" // if entry is the current page then highlight it
        }));

    navItemsEl.appendChild(frag);
}

// dispatches ui message
_.ui.dispatchMsg = (title, subtitle) => {
    $('messageTitle').textContent = title;
    $('messageSubtitle').textContent = subtitle;
    $('message').style.display = 'flex';
}

// handles stats changes.
_.ui.updateStats = (data) => {
    if (!_.ui.odometerInitialized) $('stats').style.visibility = 'visible';
    if (_.ui.shouldInitOdometerStats(false)) Odometer.init();

    const memUsed = _.helpers.humanReadableSplit(data.memUsed,1);

    $('stats-cpu').textContent = data.cpuUsage;
    $('stats-memused').textContent = memUsed.n;
    $('stats-memrest').textContent = `${memUsed.unit}/${_.helpers.humanReadable(data.memTotal,0)} ram`;

    if (_.ui.shouldInitOdometerStats(true)) Odometer.init();
    if (!_.ui.odometerInitialized) _.ui.odometerInitialized = true;
}

// sets theme
_.ui.setTheme = (theme) => {
    if (theme === 'sync' && window.matchMedia) effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    else effectiveTheme = theme;

    document.documentElement.dataset.theme = effectiveTheme;
    localStorage.setItem('theme', effectiveTheme);
}

// sets the sidebar width
_.ui.setSidebarWidth = (width) => $('sidebar').style.minWidth = $('sidebar').style.maxWidth = `${width}px`;

_.onReady(() => {
    // inital navbar fill
    _.ui.updateNavItems();

    // message dismiss button will close message
    $('messageDismiss').addEventListener('click', () => $('message').style.display = 'none');

    // subscribe to stats
    _.comm.subscribe('stat.basic', null, _.ui.updateStats);

    // config shit
    _.ui.setTheme(_CONFIG.ui?.theme || 'sync');
    _.ui.setSidebarWidth(_CONFIG.ui?.sidebarWidth || 170);
});