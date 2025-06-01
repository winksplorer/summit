// summit frontend/js/message.js - handles ui messages

_.ui.dispatchMsg = function(title, subtitle) {
    document.getElementById('messageTitle').textContent = title;
    document.getElementById('messageSubtitle').textContent = subtitle;
    document.getElementById('message').style.display = 'flex';
}

_.onReady(function(){
    document.getElementById('messageDismiss').addEventListener('click', function() {
        document.getElementById('message').style.display = 'none';
    });
})