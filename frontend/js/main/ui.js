// summit frontend/js/main/ui.js - handles ui systems

// updates navbar items
_.ui.updateNavItems = function() {
    // empty the group first
    navItemsEl = document.getElementById('navbar-items')
    navItemsEl.innerHTML = '';
    
    // request via comm, and go through the response list
    _.comm.request("info.pages")
        .then(data => {
            data.forEach(page => {
                navItemsEl.appendChild(Object.assign(document.createElement("a"), {
                    href: `${page}.html`,
                    textContent: page,
                    className: location.pathname.split("/").pop() === `${page}.html` ? "current" : "" // if entry is the current page then highlight it
                }));
            });
        });
}

// dispatches ui message
_.ui.dispatchMsg = function(title, subtitle) {
    document.getElementById('messageTitle').textContent = title;
    document.getElementById('messageSubtitle').textContent = subtitle;
    document.getElementById('message').style.display = 'flex';
}

_.onReady(function(){
    // inital navbar fill, and register message dismiss button
    _.ui.updateNavItems()

    document.getElementById('messageDismiss').addEventListener('click', function() {
        document.getElementById('message').style.display = 'none';
    });
});