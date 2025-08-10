// summit frontend/js/independent/login.js - handles automatically redirecting if you are logged in, getting the hostname, or knowing when to show the auth fail message

// fail codes and messages
const failCodes = {
    'auth': 'authentication failed',
    'root': 'login as root is forbidden',
    'inv': 'invalid session'
}

// msg element
let msgEl = document.getElementById('msg')

// go to terminal page if already logged in
fetch('/api/authenticated').then(res => res.ok && window.location.replace('/terminal.html'));

// set hostname element
fetch('/api/hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch hostname:', err));


// set theme
document.documentElement.dataset.theme = localStorage.getItem('theme') || 'dark';

// url parameters (?auth)
const authParam = new URLSearchParams(window.location.search).get('err');

// if the server says we did something wrong, then unhide the error message element and show message
msgEl.style.visibility = Object.keys(failCodes).includes(authParam) ? 'visible' : 'hidden';
msgEl.textContent = failCodes[authParam] || 'a';

// visual feedback that something is happening
document.querySelector('form').addEventListener('submit', (e) => {
    document.querySelector('input[type="submit"]').disabled = true;
    msgEl.style.visibility = 'visible';
    msgEl.style.color = 'var(--text)';
    msgEl.textContent = 'logging in';
})