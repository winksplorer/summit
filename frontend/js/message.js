// summit message.js - handles messages

export function dispatchMessage(title, subtitle) {
    document.getElementById('messageTitle').textContent = title;
    document.getElementById('messageSubtitle').textContent = subtitle;
    document.getElementById('message').style.display = 'flex';
}

document.getElementById('messageDismiss').addEventListener('click', function() {
    document.getElementById('message').style.display = 'none';
});