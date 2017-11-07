'use strict';

const _ = require('lodash');

module.exports = {
    handler: (req, res) => {
        _.assign(req.body, {date: new Date().toTimeString()})
        res.end(JSON.stringify(req.body));
    },
};
