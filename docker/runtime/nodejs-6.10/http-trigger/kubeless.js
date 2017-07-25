'use strict';

const client = require('prom-client');
const express = require('express');
const path = require('path');

const app = express();

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;

const timeHistogram = new client.Histogram({
  name: 'function_duration_seconds',
  help: 'Duration of user function in seconds',
});
const callsCounter = new client.Counter({
  name: 'function_calls_total',
  help: 'Number of calls to user function',
  labelNames: ['method'],
});
const errorsCounter = new client.Counter({
  name: 'function_failures_total',
  help: 'Number of exceptions in user function',
  labelNames: ['method'],
});

const modRootPath = process.env.MOD_ROOT_PATH ? process.env.MOD_ROOT_PATH : '/kubeless/';
const modPath = path.join(modRootPath, `${modName}.js`);
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

console.log('mod[funcHandler]', funcHandler, mod[funcHandler]);

app.get('/healthz', (req, res) => {
  res.status(200).send('OK');
});

app.get('/metrics', (req, res) => {
  res.status(200);
  res.type(client.register.contentType);
  res.send(client.register.metrics());
});

app.all('/', (req, res) => {
  const end = timeHistogram.startTimer();
  callsCounter.labels(req.method).inc();
  Promise.resolve(mod[funcHandler](req, res)).then(() => {
    end();
  }).catch((err) => {
    errorsCounter.labels(req.method).inc();
    res.status(500).send('Internal Server Error');
    console.error(`Function failed to execute: ${err.stack}`);
  });
});

app.listen(8080);
