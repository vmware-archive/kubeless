#!/usr/bin/env python

import sys#!/usr/bin/env python

import sys
import traceback
import os
import imp
import json

from multiprocessing import Process, Queue
import prometheus_client as prom
import boto3

mod_name = os.getenv('MOD_NAME')
func_handler = os.getenv('FUNC_HANDLER')
queue_name = os.getenv('QUEUE_NAME')
timeout = float(os.getenv('FUNC_TIMEOUT', 180))

aws_access_key = os.getenv('AWS_ACCESS_KEY_ID')
aws_secret_key = os.getenv('AWS_SECRET_ACCESS_KEY')
aws_region_name = os.getenv('AWS_REGION','us-east-1')

sqs = boto3.resource('sqs', region_name=aws_region_name)

group = mod_name + func_handler

queue = sqs.get_queue_by_name(QueueName=queue_name)

mod = imp.load_source('function', '/kubeless/%s.py' % mod_name)
func = getattr(mod, func_handler)

func_hist = prom.Histogram('function_duration_seconds',
                           'Duration of user function in seconds',
                           ['queue'])
func_calls = prom.Counter('function_calls_total',
                           'Number of calls to user function',
                          ['queue'])
func_errors = prom.Counter('function_failures_total',
                           'Number of exceptions in user function',
                           ['queue'])

def funcWrap(q, payload):
    q.put(func(payload))

def handle(msg):
    func_calls.labels(queue_name).inc()
    with func_errors.labels(queue_name).count_exceptions():
        with func_hist.labels(queue_name).time():
            q = Queue()
            p = Process(target=funcWrap, args=(q,msg,))
            p.start()
            p.join(timeout)
            # If thread is still active
            if p.is_alive():
                p.terminate()
                p.join()
                raise Exception('Timeout while processing the function')
            else:
                return q.get()

if __name__ == '__main__':
    prom.start_http_server(8080)

    while True:
        for msg in queue.receive_messages():
            try:
                res = handle(msg)
                sys.stdout.write(str(res) + '\n')
                sys.stdout.flush()

            except Exception:
                traceback.print_exc()
                
            msg.delete()


