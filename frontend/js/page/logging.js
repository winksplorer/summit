// summit frontend/js/page/logging.js - handles the logging page

// events are in format of {"2025-10-24": [{"time": 1761364590, "source": "test", "msg": "message"}]}
_.page.displayEvents = (events) => {
    if (!events) throw new Error('no events to display')
    const container = $('events-container')
    container.replaceChildren();

    for (const [day, list] of Object.entries(events).sort(([a], [b]) => b.localeCompare(a))) {
        const dayContainer = _.helpers.newEl('section', 'container logging_event_list', '');
        dayContainer.appendChild(_.helpers.newEl('h3', '', day))

        const frag = document.createDocumentFragment();
        for (const {time, msg, source} of list) {
            const div = _.helpers.newEl('div', '', '');
            div.append(
                _.helpers.newEl('span', 'time', new Date(time*1000).toLocaleTimeString("en-uk")),
                _.helpers.newEl('span', 'msg', msg),
                _.helpers.newEl('span', 'source', source)
            )

            frag.appendChild(div);
        }

        dayContainer.appendChild(frag);
        container.appendChild(dayContainer)
    }
}

_.page.testLogging = () => {
    let source = _.helpers.getInputValue($('source'));

    console.log(source);

    _.comm.request('log.read', {
        "source": source,
        "amount": 25,
        "page": 0
    }).then(evs => _.page.displayEvents(evs))
}