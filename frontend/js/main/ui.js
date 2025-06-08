// summit frontend/js/main/ui.js - handles ui systems

_.onReady(function(){
    _.ui.updateNavItems()

    document.getElementById('messageDismiss').addEventListener('click', function() {
        document.getElementById('message').style.display = 'none';
    });
});

_.ui.updateNavItems = function() {
    navItemsEl = document.getElementById('navbar-items')

    navItemsEl.style.visibility = 'hidden';
    navItemsEl.innerHTML = '';
    
    _.comm.request("info.pages")
        .then(data => {
            const keys = Object.keys(data.pages).sort();
            const orderedKeys = data.order.map(i => keys[i]);

            orderedKeys.forEach(key => {
                navItemsEl.appendChild(Object.assign(document.createElement("a"), {
                    href: data.pages[key],
                    textContent: key,
                    className: location.pathname.split("/").pop() === data.pages[key] ? "current" : ""
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