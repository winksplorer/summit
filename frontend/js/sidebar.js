// summit frontend/js/sidebar.js - handles navbar, and is sort of the main js code

import { fetchDataForElementText, fetchDataForElementHtml } from './modules/helpers.js';

// redirect to login page if unauthorized
fetch('/api/am-i-authed')
    .then(res => { if (res.status !== 200) throw new Error('Not authenticated'); })
    .catch(() => window.location.replace('/'));

// hostname
fetchDataForElementText('/api/get-hostname', 'hostname')

// stats every 5 seconds
setInterval(fetchDataForElementHtml('/api/stats', 'stats'), 5000);