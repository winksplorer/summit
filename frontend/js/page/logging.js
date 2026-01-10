// summit frontend/js/page/logging.js - handles the logging page

_.page.page = 0;

// events are in format of {"2025-10-24": [{"time": 1761364590, "source": "test", "msg": "message"}]}
_.page.displayEvents = (events) => {
    if (!events) throw new Error('no events to display')
    const container = $('logging_events-container')
    container.replaceChildren();

    // iterate through days
    for (const [day, list] of Object.entries(events).sort(([a], [b]) => b.localeCompare(a))) {
        const dayContainer = _.helpers.newEl('section', null, 'container logging_event_list');
        dayContainer.appendChild(_.helpers.newEl('h3', day))

        // iterate through events
        const frag = document.createDocumentFragment();
        for (const {time, msg, source} of list) {
            const div = _.helpers.newEl('div');
            div.append(
                _.helpers.newEl('span', new Date(time*1000).toLocaleTimeString(_CONFIG.ui?.timeFormat || 'en-uk', _.timeOptions), 'time'),
                _.helpers.newEl('span', msg, 'msg'),
                _.helpers.newEl('span', source, 'source')
            )

            frag.appendChild(div);
        }

        dayContainer.appendChild(frag);
        container.appendChild(dayContainer)
    }
}

_.page.readLogs = () => {
    _.comm.request('log.read', {
        "source": _.helpers.getInputValue($('source')),
        "amount": +_.helpers.getInputValue($('amount')),
        "page": _.page.page
    }).then(evs => _.page.displayEvents(evs))
}

_.page.updatePageCounter = (increase) => {
    increase ? _.page.page++ : _.page.page--;
    $('pageCounter').textContent = `page ${_.page.page+1}/x`;
    _.page.readLogs();
}

_.onReady(() => ['source', 'amount'].forEach(input => $(input).addEventListener('change', _.page.readLogs)));