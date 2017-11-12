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
const timeout = Number(process.env.FUNC_TIMEOUT || '180');
const funcPort = process.env.FUNC_PORT || 8080;

const statistics = helper.prepareStatistics('method', client);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);
const functionCallingCode = `
try {
  Promise.resolve(module.exports.${funcHandler}(req, res)).then(() => {
    end();
  }).catch((err) => {
    // Catch asynchronous errors
    handleError(err);
  });
} catch (err) {
  // Catch synchronous errors
  handleError(err);
}`;
const { vmscript, sandbox } = helper.loadFunc(modName, functionCallingCode);

app.all('*', (req, res) => {
  res.header('Access-Control-Allow-Origin', '*');
  if (req.method === 'OPTIONS') {
    // CORS preflight support (Allow any method or header requested)
    res.header('Access-Control-Allow-Methods', req.headers['access-control-request-method']);
    res.header('Access-Control-Allow-Headers', req.headers['access-control-request-headers'])
    res.end();
  } else {
    const funcLabel = modName + "-" + req.method;
    const end = statistics.timeHistogram.labels(funcLabel).startTimer();
    statistics.callsCounter.labels(funcLabel).inc();
    const handleError = (err) => {
      statistics.errorsCounter.labels(funcLabel).inc();
      res.status(500).send('Internal Server Error');
      console.error(`Function failed to execute: ${err.stack}`);
    };
    const reqSandbox = Object.assign({
      req,
      res,
      end,
      handleError,
      process: Object.assign({}, process),
    }, sandbox);
    try {
      vmscript.runInNewContext(reqSandbox, { timeout: timeout*1000 });
    } catch (err) {
      if (err.toString().match("Error: Script execution timed out")) {
          res.status(408).send(err);
          // We cannot stop the spawned process (https://github.com/nodejs/node/issues/3020)
          // we need to abruptly stop this process
          console.error('CRITICAL: Unable to stop spawned process. Exiting')
          process.exit(1)
      } else {
        handleError(err);
      }
    }
  }
});

app.listen(funcPort);
