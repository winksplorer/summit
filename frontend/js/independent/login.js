// summit frontend/js/login.js - handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

fetch('/api/am-i-authed').then(res => { if (res.ok) window.location.replace('/terminal.html') })

fetch('/api/get-hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch /api/get-hostname (was going to use as text for #hostname):', err));

// checks if url has ?auth, because the server returns that on failure
const urlParams = new URLSearchParams(window.location.search);
document.getElementById('msg').style.display = ['fail', 'rootfail'].includes(urlParams.get('auth')) ? 'block' : 'none';
if (urlParams.get('auth') == 'rootfail') document.getElementById('msg').textContent = 'login as root is forbidden';

// disable login button after pressing it and display message
document.querySelector("form").addEventListener('submit', function(e) {
    document.querySelector('input[type="submit"]').disabled = true;
    var msgEl = document.getElementById('msg');
    msgEl.style.display = 'block';
    msgEl.style.color = 'black';
    msgEl.textContent = 'logging in';
})