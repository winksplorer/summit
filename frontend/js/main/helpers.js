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
_.helpers.hex2rgb = (hex) => hex.match(/\w\w/g).map(x => parseInt(x, 16));
_.helpers.formatHslAsCSS = (h, s, l) => `hsl(${h}, ${s}%, ${l}%)`;

// https://gist.github.com/vahidk/05184faf3d92a0aa1b46aeaa93b07786
_.helpers.rgb2hsl = (r, g, b) => {
    r /= 255; g /= 255; b /= 255;
    let max = Math.max(r, g, b);
    let min = Math.min(r, g, b);
    let d = max - min;
    let h;
    if (d === 0) h = 0;
    else if (max === r) h = (g - b) / d % 6;
    else if (max === g) h = (b - r) / d + 2;
    else if (max === b) h = (r - g) / d + 4;
    let l = (min + max) / 2;
    let s = d === 0 ? 0 : d / (1 - Math.abs(2 * l - 1));
    return [(h * 60).toFixed(2), s.toFixed(2), l.toFixed(2)];
}