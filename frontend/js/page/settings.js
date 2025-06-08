// summit frontend/js/page/config.js - handles the config page

// this is placeholder code, do not add to it

(function(){
    _.page.compactui = function (enabled) {
        document.getElementById('nav-right').style.flexDirection = enabled ? "row" : "column";
        document.querySelector('.hostname').style.fontSize = enabled ? "1.5em" : "2em";
    }
    _.page.segmentstats = function (enabled) {
        document.getElementById('stats').style.background = enabled ? "black" : "transparent";
        document.getElementById('stats').style.color = enabled ? "green" : "black";
        document.getElementById('stats').style.padding = enabled ? "0.15em" : "0";
    }

    _.onReady(function () {
        document.getElementById('compactui').addEventListener('change', function () {
            _.page.compactui(this.checked)
        });

        document.getElementById('segmentstats').addEventListener('change', function () {
            _.page.segmentstats(this.checked)
        });
    });
})();