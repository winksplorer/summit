// summit frontend/js/helpers.js - helper functions

_.helpers.fetchForElementText = function(endpoint, elementId) {
    fetch(endpoint)
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById(elementId).textContent = data)
        .catch(err => {
            console.error(`failed to fetch ${endpoint} (was going to use as text for #${elementId}):`, err);
            _.ui.dispatchMsg("failed to contact server", "this may be due to the server shutting down or restarting. check the browser console for more info.");
        });
}

_.helpers.fetchforElementHTML = function(endpoint, elementId) {
    fetch(endpoint)
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById(elementId).innerHTML = data)
        .catch(err => {
            console.error(`failed to fetch ${endpoint} (was going to use as text for #${elementId}):`, err);
            _.ui.dispatchMsg("failed to contact server", "this may be due to the server shutting down or restarting. check the browser console for more info.");
        });
}