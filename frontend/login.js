// summit login.js: handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

fetch('/api/am-i-authed')
    .then(response => {
        if (response.status !== 401) return response.text();
        throw new Error('didn\'t auth, with error ' + response.status);
    })
    .then(data => { window.location.href = data; })
    .catch(error => { console.error('couldn\'t ask server for auth:', error); });

fetch('/api/get-hostname')
    .then(response => {
        if (response.ok) return response.text();
        throw new Error('network response was not ok, status ' + response.status);
    })
    .then(data => { document.getElementById('hostname').innerText = data; })
    .catch(error => { console.error('couldn\'t ask server for hostname:', error); });


// checks if url has ?auth=fail, because the servers gives that link to you if you weren't able to login.
const urlParams = new URLSearchParams(window.location.search);
if (urlParams.get('auth') === 'fail') document.getElementById('authfail').style.display = 'block';