// summit login.js: handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

fetch('/api/am-i-authed')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => window.location.replace(data))
    .catch(err => console.error('auth call failed, not redirecting:', err));


fetch('/api/get-hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch hostname:', err));


// checks if url has ?auth=fail, because the server gives that link to you if you weren't able to login.
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.get('auth') === 'fail') document.getElementById('msg').style.display = 'block';