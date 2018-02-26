'use strict';

const bodyParser = require('body-parser');
const client = require('prom-client');
const express = require('express');
const helper = require('./lib/helper');
const morgan = require('morgan');

const app = express();
app.use(morgan('combined'));
const bodParserOptions = {
  type: '*/*'
};
app.use(bodyParser.raw(bodParserOptions));

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;
const timeout = Number(process.env.FUNC_TIMEOUT || '180');
const funcPort = Number(process.env.FUNC_PORT || '8080');

const statistics = helper.prepareStatistics('method', client);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);
const functionCallingCode = `
try {
  Promise.resolve(module.exports.${funcHandler}(event, context)).then((result) => {
    switch(typeof result) {
      case 'string':
        _res.end(result);
        break;
      case 'object':
        _res.json(result);
        break;
      default:
        _res.end(JSON.stringify(result))
    }
    _end();
  }).catch((err) => {
    // Catch asynchronous errors
    _handleError(err);
  });
} catch (err) {
  // Catch synchronous errors
  _handleError(err);
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
    let data = req.body.toString('utf-8');
    if (req.get('content-type') === 'application/json') {
      data = JSON.parse(data)
    }
    const reqSandbox = Object.assign({
      event: {
        data,
        'event-type': req.get('event-type'),
        'event-id': req.get('event-id'),
        'event-time': req.get('event-time'),
        'event-namespace': req.get('event-namespace'),
        extensions: {
          request: req,
        },
      },
      context: {
        'function-name': funcHandler,
        timeout,
        runtime: process.env.FUNC_RUNTIME,
        'memory-limit': process.env.FUNC_MEMORY_LIMIT,
      },
      _handleError: handleError,
      _res: res,
      _end: end,
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
