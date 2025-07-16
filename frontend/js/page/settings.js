// summit frontend/js/page/config.js - handles the config page

// this is placeholder code, do not add to it

_.page.compactui = (enabled) => {
    $('nav-right').style.flexDirection = enabled ? "row" : "column";
    document.querySelector('.hostname').style.fontSize = enabled ? "1.5em" : "2em";
}
_.page.scale = (value) => {
    document.body.style.fontSize = `${value}em`;
}

_.onReady(() => {
    $('compactui').addEventListener('change', e => _.page.compactui(e.target.checked));
    $('scale').addEventListener('change',  e => _.page.scale(e.target.value));
});