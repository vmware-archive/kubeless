'use strict';

const _ = require('lodash');

module.exports = {
    handler: (event, context) => {
        _.assign(event, {date: new Date().toTimeString()})
        return JSON.stringify(event);
    },
};
