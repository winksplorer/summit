// summit frontend/js/login.js - handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

fetch('/api/am-i-authed')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => window.location.replace(data))
    .catch(err => console.error('auth call failed, not redirecting:', err));

fetch('/api/get-hostname')
        .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
        .then(data => document.getElementById('hostname').textContent = data)
        .catch(err => console.error('failed to fetch /api/get-hostname (was going to use as text for #hostname):', err));

// checks if url has ?auth, because the server returns that on failure
const urlParams = new URLSearchParams(window.location.search);
document.getElementById('msg').style.display = ['fail', 'rootfail'].includes(urlParams.get('auth')) ? 'block' : 'none';
if (urlParams.get('auth') == 'rootfail') document.getElementById('msg').textContent = 'login as root is forbidden';