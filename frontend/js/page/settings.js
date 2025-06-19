// summit frontend/js/page/config.js - handles the config page

// this is placeholder code, do not add to it

(function(){
    _.page.compactui = function (enabled) {
        document.getElementById('nav-right').style.flexDirection = enabled ? "row" : "column";
        document.querySelector('.hostname').style.fontSize = enabled ? "1.5em" : "2em";
    }
    _.page.scale = function (value) {
        document.body.style.fontSize = `${value}em`;
    }

    _.onReady(function () {
        document.getElementById('compactui').addEventListener('change', function () {
            _.page.compactui(this.checked)
        });

        document.getElementById('scale').addEventListener('change', function () {
            _.page.scale(this.value)
        });
    });
})();