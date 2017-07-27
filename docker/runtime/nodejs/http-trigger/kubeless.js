'use strict';

const client = require('prom-client');
const express = require('express');
const helper = require('./lib/helper');
const morgan = require('morgan');

const app = express();
app.use(morgan('combined'));

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;

const statistics = helper.prepareStatistics('method', client);
const mod = helper.loadFunc(modName, funcHandler);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);

app.all('/', (req, res) => {
  const end = statistics.timeHistogram.labels(req.method).startTimer();
  statistics.callsCounter.labels(req.method).inc();
  const handleError = (err) => {
    statistics.errorsCounter.labels(req.method).inc();
    res.status(500).send('Internal Server Error');
    console.error(`Function failed to execute: ${err.stack}`);
  };
  try {
    Promise.resolve(mod[funcHandler](req, res)).then(() => {
      end();
    }).catch((err) => {
      // Catch asynchronous errors
      handleError(err);
    });
  } catch (err) {
    // Catch synchronous errors
    handleError(err);
  }
});

app.listen(8080);
