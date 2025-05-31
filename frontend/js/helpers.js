// summit frontend/js/modules/helpers.js - helper functions

import { dispatchMessage } from './message.js';

export function fetchDataForElementText(endpoint, elementId) {
    fetch(endpoint)
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById(elementId).textContent = data)
        .catch(err => {
            console.error(`failed to fetch ${endpoint} (was going to use as text for #${elementId}):`, err);
            dispatchMessage("failed to contact server", "this may be due to the server shutting down or restarting. check the browser console for more info.");
        });
}

export function fetchDataForElementHtml(endpoint, elementId) {
    fetch(endpoint)
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById(elementId).innerHTML = data)
        .catch(err => {
            console.error(`failed to fetch ${endpoint} (was going to use as text for #${elementId}):`, err);
            dispatchMessage("failed to contact server", "this may be due to the server shutting down or restarting. check the browser console for more info.");
        });
}