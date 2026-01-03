// summit frontend/js/main/helpers.js - helper functions

_.helpers.newButton = (text, disabled, cb) => Object.assign(document.createElement('button'), { textContent: text, onclick: cb, disabled: disabled || false });
_.helpers.newCheckbox = (checked, cb) => Object.assign(document.createElement('input'), { type: 'checkbox', onchange: cb, checked: checked });
_.helpers.newElWithID = (tag, classes, id) => Object.assign(document.createElement(tag), { className: classes, id: id});
_.helpers.newEl = (tag, classes, content, id) => 
    Object.assign(document.createElement(tag),
        {
            className: classes,
            id: id || '',
            innerHTML: content
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;")
                .replace(/\n/g, "<br>")
        }
    );

_.helpers.getCSSVar = (varName) => getComputedStyle(document.documentElement).getPropertyValue(`--${varName}`).trim();
_.helpers.shortenText = (str, head = 14, tail = 4) => (str.length <= head + tail + 1) ? str : `${str.slice(0, head)}...${str.slice(-tail)}`;
_.helpers.humanReadableSplit = (size, places = 2, baseUnit = false, base = _CONFIG.ui.binaryUnits ? 1024 : 1000) => {
    const units = baseUnit ? ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB'] : ['', 'k', 'm', 'g', 't', 'p', 'e']
    const i = size && Math.floor(Math.log(size) / Math.log(base));
    return {'n': +(size / Math.pow(base, i)).toFixed(places), 'unit': units[i]};
}

_.helpers.humanReadable = (size, places = 2, baseUnit = false, base = _CONFIG.ui.binaryUnits ? 1024 : 1000) => {
    const split = _.helpers.humanReadableSplit(size,places,baseUnit,base);
    return split.n + split.unit;
}

_.helpers.resizeCanvas = (canvas) => {
    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();

    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;

    const ctx = canvas.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
}

_.helpers.getInputValue = (el) => el.type === 'checkbox' ? el.checked : el.value;
_.helpers.setInputValue = (el, val) => el.type === 'checkbox' ? (el.checked = val) : (el.value = val);
_.helpers.getObjectValue = (obj, key) => {
    const parts = key.split('.');
    let value = obj;

    for (const part of parts) {
        if (typeof value !== 'object' || value === null) throw new Error(`not an object at ${part}`);
        if (!(part in value)) throw new Error(`couldn't find ${part}`);
        value = value[part];
    }

    return value;
}