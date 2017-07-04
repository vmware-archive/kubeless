#!/usr/bin/env python

import sys
import traceback
import os
import imp
import json

from kafka import KafkaConsumer
import prometheus_client as prom

mod_name = os.getenv('MOD_NAME')
func_handler = os.getenv('FUNC_HANDLER')
topic_name = os.getenv('TOPIC_NAME')

group = mod_name + func_handler

mod = imp.load_source('function', '/kubeless/%s.py' % mod_name)
func = getattr(mod, func_handler)

func_hist = prom.Histogram('function_duration_seconds',
                           'Duration of user function in seconds',
                           ['topic'])
func_calls = prom.Counter('function_calls_total',
                           'Number of calls to user function',
                          ['topic'])
func_errors = prom.Counter('function_failures_total',
                           'Number of exceptions in user function',
                           ['topic'])

def json_safe_loads(msg):
    try:
        data = json.loads(msg)
        return {'type': 'json', 'payload': data}
    except:
        return {'type': 'text', 'payload': msg}

consumer = KafkaConsumer(
    bootstrap_servers='kafka.kubeless:9092',
    group_id=group, value_deserializer=json_safe_loads)
consumer.subscribe([topic_name])

def handle(msg):
    func_calls.labels(topic_name).inc()
    with func_errors.labels(topic_name).count_exceptions():
        with func_hist.labels(topic_name).time():
            return func(msg.value['payload'])

if __name__ == '__main__':
    prom.start_http_server(8080)

    while True:
        for msg in consumer:
            try:
                res = handle(msg)
                sys.stdout.write(str(res) + '\n')
                sys.stdout.flush()

            except Exception:
                traceback.print_exc()
