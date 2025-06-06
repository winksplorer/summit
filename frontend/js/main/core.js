// summit frontend/js/core.js - sets up namespaces

window._ = window._ || {}; // global _ object
_.helpers = _.helpers || {}; // helpers.js, helper functions
_.page = _.page || {}; // page-specific, like terminal or config boxes
_.comm = _.comm || {}; // backend communication
_.config = _.config || {}; // configuration
_.ui = _.ui || {}; // user interface shit, like messages



(function(){
    let ready = false;
    let callbacks = [];

    _.onReady = function(cb) {
        if (ready) cb();
        else callbacks.push(cb);
    };

    _.init = function() {
        ready = true;
        callbacks.forEach(cb => cb());
        callbacks = null;
    };
})();

_.onReady(function() {
    for (let el of document.querySelectorAll('[data-t]')) _.comm.request(el.dataset.t).then(data => el.textContent = data[el.dataset.key]);
});