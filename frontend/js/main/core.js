// summit frontend/js/main/core.js - sets up namespaces

window._ = window._ || {}; // global _ object
_.helpers = _.helpers || {}; // helpers.js, helper functions
_.page = _.page || {}; // page-specific, like terminal or config boxes
_.comm = _.comm || {}; // backend communication
_.config = _.config || {}; // configuration
_.ui = _.ui || {}; // user interface shit, like messages

// configure odometer
window.odometerOptions = {
    auto: false
};

// mfw jquery wannabe
window.$ = document.getElementById.bind(document);

// runs callback only when page is fully loaded
_.onReady = (cb) => document.readyState === 'loading'
    ? document.addEventListener('DOMContentLoaded', cb)
    : cb();

// true if we just logged in or if we came here directly from another site (no reloads, no same-domain pages except login & admin)
_.isColdEntry = performance.getEntriesByType("navigation")[0]?.type !== 'reload'
                && ['', location.origin, location.origin + '/'].includes(document.referrer.split('?')[0]);

// options for toLocaleTimeString
_.timeOptions = {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
};