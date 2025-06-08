// summit frontend/js/main/core.js - sets up namespaces

window._ = window._ || {}; // global _ object
_.helpers = _.helpers || {}; // helpers.js, helper functions
_.page = _.page || {}; // page-specific, like terminal or config boxes
_.comm = _.comm || {}; // backend communication
_.config = _.config || {}; // configuration
_.ui = _.ui || {}; // user interface shit, like messages

_.onReady = cb => {
    if (document.readyState !== 'complete') document.addEventListener("DOMContentLoaded", cb);
    else cb();
};

_.onReady(function() {
    // data-t is what type to request from server, and data-key is what part of the response should be used for element text
    for (let el of document.querySelectorAll('[data-t]')) _.comm.request(el.dataset.t).then(data => el.textContent = data[el.dataset.key]);
});