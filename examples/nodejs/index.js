'use strict';

module.exports = {
    helloGet: require('./helloget').foo,
    helloWithData: require('./hellowithdata').handler,
    helloEvent: require('./helloevent').handler,
}
