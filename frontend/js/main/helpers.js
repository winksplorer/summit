// summit frontend/js/main/helpers.js - helper functions

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

_.helpers.getCSSVar = (varName) => getComputedStyle(document.documentElement).getPropertyValue(`--${varName}`).trim();