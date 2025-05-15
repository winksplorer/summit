// summit frontend/js/navbar.js - handles navbar, and is sort of the main js code

import { fetchDataForElementText, fetchDataForElementHtml } from './modules/helpers.js';

// redirect to login page if unauthorized
fetch('/api/am-i-authed').then(res => { if (!res.ok) window.location.replace('/') })

// hostname
fetchDataForElementText('/api/get-hostname', 'hostname');

// build string
fetchDataForElementText('/api/buildstring', 'footer');

// stats
function updateStats() {
    fetch('/api/stats')
        .then(res => res.ok ? res.json() : Promise.reject(`HTTP ${res.status}`))
        .then(data => {
            document.getElementById("stats-cpu").textContent = Math.round(data.processorUsage);
            document.getElementById("stats-memused").textContent = Math.round(data.memoryUsage);
            document.getElementById("stats-memrest").textContent = `${data.memoryUsageUnit}/${data.memoryTotal}`;
        })
        .catch(err => {
            console.error(`failed to fetch ${endpoint} (was going to use as text for #${elementId}):`, err);
            dispatchMessage("failed to contact server", "this may be due to the server shutting down or restarting. check the browser console for more info.");
        });
}

updateStats();
setInterval(() => updateStats(), 5000);

// add navbar items
fetch("/api/server-pages")
    .then(res => res.json())
    .then(pairs => pairs.forEach(([key, val]) => {
    let a = Object.assign(document.createElement("a"), {
        href: key,
        textContent: val,
        className: location.pathname.split("/").pop() == key ? "current" : ""
    });
    document.getElementById("navbar-items").appendChild(a);
    }));

