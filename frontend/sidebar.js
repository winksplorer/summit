// summit sidebar.js - handles sidebar & automatic redirect to login page 

// redirect to login page if unauthorized
fetch('/api/am-i-authed')
    .then(response => {
        if (response.status !== 200) window.location.href = "/";
    })

// hostname
fetch('/api/get-hostname')
    .then(response => {
        if (response.ok) return response.text();
        throw new Error('network response was not ok, status ' + response.status);
    })
    .then(data => { document.getElementById('hostname').innerText = data; })
    .catch(error => { console.error('couldn\'t ask server for hostname:', error); });


function updateMemory() {
    fetch('/api/stat-memory')
        .then(response => {
            if (response.ok) return response.text();
            throw new Error('network response was not ok, status ' + response.status);
        })
        .then(data => { document.getElementById('memory').innerText = data; })
        .catch(error => { console.error('couldn\'t ask server for memory stat:', error); });
}

function updateCpu() {
    fetch('/api/stat-cpu')
        .then(response => {
            if (response.ok) return response.text();
            throw new Error('network response was not ok, status ' + response.status);
        })
        .then(data => { document.getElementById('cpu').innerText = data; })
        .catch(error => { console.error('couldn\'t ask server for cpu stat:', error); });
}

updateMemory();
updateCpu();
setInterval(updateMemory, 5000);
setInterval(updateCpu, 5000);