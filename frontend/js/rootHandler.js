// summit rootHandler.js - asks the server if it's running as root, and changes some behavior of the webui

import { dispatchMessage } from './message.js';

// get the EUID the server is running as
fetch('/api/server-euid')
    .then(res => res.ok ? res.json() : Promise.reject(`HTTP ${res.status}`))
    .then(data => {
        if (data.euid !== 0) {
            // handle if we're not root: disable all controls that have class .rootRequired, and dispatch a message
            document.querySelectorAll('.rootRequired').forEach(element => element.style.display = 'none')
            dispatchMessage("summit isn't running as root.", "some controls and features are disabled. run the summit server as root to enable them.")
        }
    })
    .catch(err => console.error('failed to get server euid:', err));