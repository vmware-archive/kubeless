'use strict';

const fs = require('fs');

function readDependencies(pkgFile) {
  try {
    const data = JSON.parse(fs.readFileSync(pkgFile));
    const deps = data.dependencies;
    return (deps && typeof deps === 'object') ? Object.getOwnPropertyNames(deps) : [];
  } catch(e) {
    return [];
  }
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
  readDependencies,
  prepareStatistics,
  routeLivenessProbe,
  routeMetrics,
};
