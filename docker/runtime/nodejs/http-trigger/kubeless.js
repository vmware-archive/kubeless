'use strict';

const bodyParser = require('body-parser');
const client = require('prom-client');
const express = require('express');
const helper = require('./lib/helper');
const morgan = require('morgan');

const app = express();
app.use(morgan('combined'));
app.use(bodyParser.json()); // support json encoded bodies
app.use(bodyParser.urlencoded({ extended: true })); // support encoded bodies

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;

const statistics = helper.prepareStatistics('method', client);
const mod = helper.loadFunc(modName, funcHandler);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);

app.all('*', (req, res) => {
  res.header('Access-Control-Allow-Origin', '*');
  if (req.method === 'OPTIONS') {
    // CORS preflight support (Allow any method or header requested)
    res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
    res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers'])
    res.end();
  } else {
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
  }
});

app.listen(8080);
