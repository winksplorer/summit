// summit frontend/js/main/ui.js - handles ui systems

_.ui.updateNavItems = function() {
    navItemsEl = document.getElementById('navbar-items')

    navItemsEl.style.visibility = 'hidden';
    navItemsEl.innerHTML = '';
    
    _.comm.request("info.pages")
        .then(data => {
            data.forEach(page => {
                navItemsEl.appendChild(Object.assign(document.createElement("a"), {
                    href: `${page}.html`,
                    textContent: page,
                    className: location.pathname.split("/").pop() === `${page}.html` ? "current" : ""
                }));
            });
            navItemsEl.style.visibility = 'visible';
        });
}

_.ui.dispatchMsg = function(title, subtitle) {
    document.getElementById('messageTitle').textContent = title;
    document.getElementById('messageSubtitle').textContent = subtitle;
    document.getElementById('message').style.display = 'flex';
}

_.onReady(function(){
    _.ui.updateNavItems()

    document.getElementById('messageDismiss').addEventListener('click', function() {
        document.getElementById('message').style.display = 'none';
    });
});