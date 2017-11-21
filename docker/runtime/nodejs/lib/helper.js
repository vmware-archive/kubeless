'use strict';

const fs = require('fs');
const path = require('path');
const Module = require('module');
const Vm = require('vm');

function loadFunc(name, handler, params) {
  const modRootPath = process.env.MOD_ROOT_PATH ? process.env.MOD_ROOT_PATH : '/kubeless/';
  const modPath = path.join(modRootPath, `${name}.js`);
  console.log('Loading', modPath);
  const mod = new Module(modPath);
  mod.paths = module.paths;
  const scriptContext = fs.readFileSync(modPath, { encoding: 'utf-8' });
  console.log(scriptContext);
  const script = `${scriptContext}
try {
  Promise.resolve(module.exports.${handler}(${params})).then(() => {
    end();
  }).catch((err) => {
    // Catch asynchronous errors
    handleError(err);
  });
} catch (err) {
  // Catch synchronous errors
  handleError(err);
}`
  const vmscript = new Vm.Script(script, {
    filename: modPath,
    displayErrors: true,
  });
  const sandbox = {
    module: mod,
    __filename: modPath,
    __dirname: path.dirname(modPath),
    setInterval, setTimeout, setImmediate,
    clearInterval, clearTimeout, clearImmediate,
    console: console,
    require: function (p) {
      return mod.require(p);
    },
  };
  return { vmscript, sandbox };
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
