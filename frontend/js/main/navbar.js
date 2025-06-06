// summit frontend/js/navbar.js - handles navbar, and is sort of the main js code

// redirect to login page if unauthorized
fetch('/api/am-i-authed').then(res => { if (!res.ok) window.location.replace('/') })

_.onReady(function(){
    // add navbar items
    fetch("/api/server-pages")
        .then(res => res.json())
        .then(pairs => pairs.forEach(([key, val]) => {
        let a = Object.assign(document.createElement("a"), {
            href: key,
            textContent: val,
            className: location.pathname.split("/").pop() == key ? "current" : ""
        });
        document.getElementById("navbar-items").appendChild(a);
        }));
});