'use strict';

const _ = require('lodash');
const client = require('prom-client');
const express = require('express');
const helper = require('./lib/helper');
const kafka = require('kafka-node');
const morgan = require('morgan');

const app = express();
app.use(morgan('combined'));

const modName = process.env.MOD_NAME;
const funcHandler = process.env.FUNC_HANDLER;
const timeout = Number(process.env.FUNC_TIMEOUT || '180');
const funcPort = process.env.FUNC_PORT || 8080;

const kafkaSvc = _.get(process.env, 'KUBELESS_KAFKA_SVC', 'kafka');
const kafkaNamespace = _.get(process.env, 'KUBELESS_KAFKA_NAMESPACE', 'kubeless');
const kafkaHost = `${kafkaSvc}.${kafkaNamespace}:9092`;
const groupId = `${modName}${funcHandler}`;
const kafkaConsumer = new kafka.ConsumerGroup({
  kafkaHost,
  groupId,
}, [process.env.TOPIC_NAME]);

const statistics = helper.prepareStatistics('method', client);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);

const functionCallingCode = `
try {
  Promise.resolve(module.exports.${funcHandler}(message)).then(() => {
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

kafkaConsumer.on('message', (message) => {
  const funcLabel = modName + "-" + message.topic;
  const end = statistics.timeHistogram.labels(funcLabel).startTimer();
  const handleError = (err) => {
    statistics.errorsCounter.labels(funcLabel).inc();
    console.error(`Function failed to execute: ${err.stack}`);
  };
  statistics.callsCounter.labels(funcLabel).inc();
  const reqSandbox = Object.assign({
    message: message.value,
    end,
    handleError,
    process: Object.assign({}, process),
  }, sandbox);
  try {
    vmscript.runInNewContext(reqSandbox, { timeout: timeout*1000 });
  } catch (err) {
    if (err.toString().match("Error: Script execution timed out")) {
      // We cannot stop the spawned process (https://github.com/nodejs/node/issues/3020)
      // we need to abruptly stop this process
      console.error('CRITICAL: Unable to stop spawned process. Exiting')
      process.exit(1)
    } else {
      handleError(err);
    }
  }
});

app.listen(funcPort);
