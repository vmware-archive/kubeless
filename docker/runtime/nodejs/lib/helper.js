'use strict';

const path = require('path');

function loadFunc(name, handler) {
  const modRootPath = process.env.MOD_ROOT_PATH ? process.env.MOD_ROOT_PATH : '/kubeless/';
  const modPath = path.join(modRootPath, `${name}.js`);
  console.log('Loading', modPath);
  let mod = null;
  try {
    mod = require(modPath); // eslint-disable-line global-require
  } catch (e) {
    console.error(
      'No valid module found for the name: function, Failed to import module:\n' +
      `${e.message}`
    );
    process.exit(1);
  }
  console.log('mod[funcHandler]', handler, mod[handler]);
  return mod;
}

function prepareStatistics(label, promClient) {
  const timeHistogram = new promClient.Histogram({
    name: 'function_duration_seconds',
    help: 'Duration of user function in seconds',
    labelNames: [label],
  });
  const callsCounter = new promClient.Counter({
    name: 'function_calls_total',
    help: 'Number of calls to user function',
    labelNames: [label],
  });
  const errorsCounter = new promClient.Counter({
    name: 'function_failures_total',
    help: 'Number of exceptions in user function',
    labelNames: [label],
  });
  return {
    timeHistogram,
    callsCounter,
    errorsCounter,
  };
}

function routeLivenessProbe(expressApp) {
  expressApp.get('/healthz', (req, res) => {
    res.status(200).send('OK');
  });
}

function routeMetrics(expressApp, promClient) {
  expressApp.get('/metrics', (req, res) => {
    res.status(200);
    res.type(promClient.register.contentType);
    res.send(promClient.register.metrics());
  });
}

module.exports = {
  loadFunc,
  prepareStatistics,
  routeLivenessProbe,
  routeMetrics,
};
