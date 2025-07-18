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
_.onReady = (cb) => document.readyState !== 'complete'
    ? document.addEventListener('DOMContentLoaded', cb)
    : cb();

// true if we just logged in or if we came here directly from another site (no reloads, no same-domain pages except login & admin)
_.isColdEntry = performance.getEntriesByType("navigation")[0]?.type !== 'reload'
                && ['', location.origin, location.origin + '/'].includes(document.referrer);


_.onReady(() => {
    // data-t is what type to request from server, and data-key is what part of the response should be used for element text
    for (let el of document.querySelectorAll('[data-t]'))
        _.comm.request(el.dataset.t).then(data => el.textContent = data[el.dataset.key]);
});