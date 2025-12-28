// summit frontend/js/page/settings.js - handles the settings page

_.page.changedSettings = {};

_.page.saveSettings = () => {
    _.comm.request('config.set', _.page.changedSettings).then(() => {
        $('save-button').disabled = true; // firefox, why do you hate me?
        location.reload()
    });
}

_.onReady(() => {
    for (const setting of document.querySelectorAll('[data-key]')) {
        try {
            const key = setting.dataset.key;
            const og = _.helpers.getObjectValue(_CONFIG, key);

            _.helpers.setInputValue(setting, og)

            setting.addEventListener(setting.type === 'text' ? 'input' : 'change', e => {
                const val = _.helpers.getInputValue(e.target);

                if (key in _.page.changedSettings && og === val)
                    delete _.page.changedSettings[key];
                else
                    _.page.changedSettings[key] = val;

                $('save-button').disabled = !Object.keys(_.page.changedSettings).length;
            });
        } catch(err) {
            console.error('error processing setting:', setting, err);
        }
    }
});