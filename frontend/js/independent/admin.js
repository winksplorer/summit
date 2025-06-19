// summit frontend/js/independent/admin.js - handles admin.html root login page

// set hostname element
fetch('/api/get-hostname')
    .then(res => res.ok ? res.text() : Promise.reject(`HTTP ${res.status}`))
    .then(data => document.getElementById('hostname').textContent = data)
    .catch(err => console.error('failed to fetch hostname:', err));

// elements
const msgEl = document.getElementById('msg')
const opEl = document.getElementById('operation')
const buttonEl = document.querySelector('input[type="submit"]')

// url parameters (?operation)
const urlParams = new URLSearchParams(window.location.search);

// handles operation text
switch (urlParams.get('operation')) {
    case 'reboot':
        opEl.textContent = 'reboot the server'; break;
    case 'poweroff':
        opEl.textContent = 'power off the server'; break;
}

// handles the form submit
document.getElementById('authform').addEventListener('submit', function(e) {
    e.preventDefault();

    // visual feedback that something is happening
    buttonEl.disabled = true;
    var msgEl = document.getElementById('msg');
    msgEl.style.display = 'block';
    msgEl.style.color = 'black';
    msgEl.textContent = 'authenticating';

    // send our data
    fetch('/api/sudo', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            password: document.getElementById('password').value,
            operation: urlParams.get('operation')
        }),
    })
    .then(res => res.ok ? res.ok : Promise.reject(`HTTP ${res.status}`))
    .then(() => { // if we don't fail
        msgEl.innerHTML = 'operation completed. <a href="/">return</a>';
        msgEl.style.color = '#006b0c';
        msgEl.style.display = 'block';
        buttonEl.disabled = false;
    })
    .catch(err => { // if we fail
        console.error('failed to send form data:', err);
        msgEl.innerHTML = err === 'HTTP 401' ? 'authentication failed <a href="/">return</a>' : `operation failed (${err}). <a href="/">return</a>`; // print special message for 401, otherwise just a generic message
        msgEl.style.color = '#9f0000';
        msgEl.style.display = 'block';
        buttonEl.disabled = false;
    });
});
