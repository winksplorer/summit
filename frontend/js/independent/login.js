// summit frontend/js/independent/login.js - handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

// go to terminal page if already logged in
fetch('/api/am-i-authed').then(res => { if (res.ok) window.location.replace('/terminal.html') })

// set hostname element
fetch('/api/get-hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch /api/get-hostname (was going to use as text for #hostname):', err));

// url parameters (?auth)
const urlParams = new URLSearchParams(window.location.search);

// if the server says we did something wrong, then unhide the error message element
document.getElementById('msg').style.display = ['fail', 'rootfail'].includes(urlParams.get('auth')) ? 'block' : 'none';

// by default it's the fail message, if it's rootfail then set rootfail message
if (urlParams.get('auth') == 'rootfail') document.getElementById('msg').textContent = 'login as root is forbidden';

// visual feedback that something is happening
document.querySelector("form").addEventListener('submit', function(e) {
    document.querySelector('input[type="submit"]').disabled = true;
    var msgEl = document.getElementById('msg');
    msgEl.style.display = 'block';
    msgEl.style.color = 'black';
    msgEl.textContent = 'logging in';
})