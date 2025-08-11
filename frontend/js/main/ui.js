// summit frontend/js/main/ui.js - handles ui systems

// returns true if the current theme is the dark theme
_.ui.isDarkTheme = () => document.documentElement.dataset.theme === 'dark';

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
    $('stats-memrest').textContent = `${data.memUsageUnit}/${data.memTotal} ram`;

    if (!_.isColdEntry && !_.ui.odometerInitialized) Odometer.init();
    if (!_.ui.odometerInitialized) _.ui.odometerInitialized = true;
}

// sets theme
_.ui.setTheme = (theme) => {
    if (theme === 'sync') effectiveTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    else effectiveTheme = theme;

    document.documentElement.dataset.theme = effectiveTheme;
    localStorage.setItem('theme', effectiveTheme);
}

// sets accent color
_.ui.setAccentColor = (color) => {
    const rootStyle = document.documentElement.style;
    let [h, s, l] = _.helpers.rgb2hsl(..._.helpers.hex2rgb(color));
    [s, l] = [s*100, l*100];

    rootStyle.setProperty('--a0', _.helpers.formatHslAsCSS(h, s, _.ui.isDarkTheme() ? l-20 : l-40 ));
    rootStyle.setProperty('--a1', _.helpers.formatHslAsCSS(h, s, _.ui.isDarkTheme() ? l : l-20));
}

// sets the sidebar width
_.ui.setSidebarWidth = (width) => $('sidebar').style.minWidth = $('sidebar').style.maxWidth = `${width}px`;

_.onReady(() => {
    // inital navbar fill
    _.ui.updateNavItems();

    // message dismiss button will close message
    $('messageDismiss').addEventListener('click', () => $('message').style.display = 'none');

    // config shit
    _.ui.setTheme(_CONFIG.ui?.theme || 'sync');
    _.ui.setSidebarWidth(_CONFIG.ui?.sidebarWidth || 170);
    _.ui.setAccentColor(_CONFIG.ui?.accentColor || '#99c8ff');
});