// summit frontend/js/page/settings.js - handles the settings page

_.page.changedSettings = {};

_.page.saveSettings = () => {
    _.comm.request('config.set', _.page.changedSettings);
    location.reload();
}

_.onReady(() => {
    for (const setting of document.querySelectorAll('[data-key]')) {
        _.helpers.setInputValue(setting, _.helpers.getObjectValue(_CONFIG, setting.dataset.key))

        setting.addEventListener('change', e => {
            const key = e.target.dataset.key;
            const val = _.helpers.getInputValue(e.target)

            if (key in _.page.changedSettings && _.helpers.getObjectValue(_CONFIG, key) === val) {
                delete _.page.changedSettings[key];
                return;
            }

            _.page.changedSettings[key] = val;
        });
    }
});