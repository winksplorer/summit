// summit frontend/js/page/logging.js - handles the logging page

// in format of {"2025-10-24": [{"time": 1761364590, "source": "test", "msg": "message"}]}
_.page.events = {};

_.page.testLogging = () => {
    let source = _.helpers.getInputValue($('source'));

    console.log(source);

    _.comm.request('log.read', {
        "source": source,
        "amount": 25,
        "page": 0
    }).then(evs => {
        evs.forEach(ev => {
            let el = document.createElement("p");
            el.textContent = `${ev['time']}: ${ev['source']}: ${ev['msg']}`;
            document.getElementsByClassName("main")[0].appendChild(el);
        });
    })
}