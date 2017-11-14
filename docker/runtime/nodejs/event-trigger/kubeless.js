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

const kafkaSvc = _.get(process.env, 'KUBELESS_KAFKA_SVC', 'kafka');
const kafkaNamespace = _.get(process.env, 'KUBELESS_KAFKA_NAMESPACE', 'kubeless');
const kafkaHost = `${kafkaSvc}.${kafkaNamespace}:9092`;
const groupId = `${modName}${funcHandler}`;
const kafkaConsumer = new kafka.ConsumerGroup({
  kafkaHost,
  groupId,
}, [process.env.TOPIC_NAME]);

const statistics = helper.prepareStatistics('method', client);
const mod = helper.loadFunc(modName, funcHandler);
helper.routeLivenessProbe(app);
helper.routeMetrics(app, client);

kafkaConsumer.on('message', (message) => {
  const end = statistics.timeHistogram.labels(message.topic).startTimer();
  const handleError = (err) => {
    statistics.errorsCounter.labels(message.topic).inc();
    console.error(`Function failed to execute: ${err.stack}`);
  };
  statistics.callsCounter.labels(message.topic).inc();
  try {
    Promise.resolve(mod[funcHandler](message.value)).then(() => {
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
