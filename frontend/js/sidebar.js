// summit sidebar.js - handles sidebar & automatic redirect to login page 

// redirect to login page if unauthorized
fetch('/api/am-i-authed')
    .then(res => { if (res.status !== 200) throw new Error('Not authenticated'); })
    .catch(() => window.location.replace('/'));

// hostname
fetch('/api/get-hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch hostname:', err));

function updateStats() {
    fetch('/api/stats')
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById('stats').innerHTML = data)
        .catch(err => console.error('failed to fetch server\'s stats', err));
}

updateStats();
setInterval(updateStats, 5000);