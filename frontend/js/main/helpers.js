// summit frontend/js/main/helpers.js - helper functions

_.helpers.newEl = (tag, classes, content) => 
    Object.assign(document.createElement(tag),
        {
            className: classes,
            innerHTML: content
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;")
                .replace(/\n/g, "<br>")
        }
    );

_.helpers.getCSSVar = (varName) => getComputedStyle(document.documentElement).getPropertyValue(`--${varName}`).trim();
_.helpers.shortenText = (str, head = 14, tail = 4) => (str.length <= head + tail + 1) ? str : `${str.slice(0, head)}...${str.slice(-tail)}`;

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