// summit frontend/js/config.js - handles configuration and the config page

function compactui(enabled) {
    document.getElementById('nav-right').style.flexDirection = enabled ? "row" : "column";
    document.getElementById('hostname').style.fontSize = enabled ? "1.5em" : "2em";
}
function segmentstats(enabled) {
    document.getElementById('stats').style.fontFamily = enabled ? '"Seven Segment", sans-serif' : '"Helvetica Neue Condensed", sans-serif';
    document.getElementById('stats').style.background = enabled ? "black" : "transparent";
    document.getElementById('stats').style.color = enabled ? "green" : "black";
    document.getElementById('stats').style.padding = enabled ? "0.15em" : "0";
}

document.getElementById('compactui').addEventListener('change', function() {
    const data = {
        checked: this.checked
    };
    compactui(data.checked)
});

document.getElementById('segmentstats').addEventListener('change', function() {
    const data = {
        checked: this.checked
    };
    segmentstats(data.checked)
});