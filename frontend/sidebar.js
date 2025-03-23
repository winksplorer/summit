// summit sidebar.js - handles sidebar & automatic redirect to login page 

// redirect to login page if unauthorized
fetch('/api/am-i-authed')
    .then(response => {
        if (response.status !== 200 && response.status !== 404) window.location.href = "/";
    })

// hostname
fetch('/api/get-hostname')
    .then(response => {
        if (response.ok) return response.text();
        throw new Error('network response was not ok, status ' + response.status);
    })
    .then(data => { document.getElementById('hostname').innerText = data; })
    .catch(error => { console.error('couldn\'t ask server for hostname:', error); });


function updateStats() {
    fetch('/api/stats')
        .then(response => {
            if (response.ok) return response.text();
            throw new Error('network response was not ok, status ' + response.status);
        })
        .then(data => { document.getElementById('stats').innerHTML = data; })
        .catch(error => { console.error('couldn\'t ask server for memory stat:', error); });
}

updateStats();
setInterval(updateStats, 5000);