#!/usr/bin/env python

import os
import imp
import datetime
import julia

from multiprocessing import Process, Queue
import bottle
import prometheus_client as prom

func_port = os.getenv('FUNC_PORT', 8080)

timeout = float(os.getenv('FUNC_TIMEOUT', 180))

app = application = bottle.app()

func_hist = prom.Histogram('function_duration_seconds',
                           'Duration of user function in seconds',
                           ['method'])
func_calls = prom.Counter('function_calls_total',
                           'Number of calls to user function',
                          ['method'])
func_errors = prom.Counter('function_failures_total',
                           'Number of exceptions in user function',
                           ['method'])

function_context = {
    'timeout': timeout,
    'runtime': os.getenv('FUNC_RUNTIME'),
    'memory-limit': os.getenv('FUNC_MEMORY_LIMIT'),
}

    
def funcWrap(q, event, c):
    ju = julia.Julia()
    julia_function = ju.include('/kubeless/%s.jl' % os.getenv('MOD_NAME'))
    c['function-name'] = str(julia_function)
    try:
        q.put(julia_function(event, c))
    except Exception as inst:
        q.put(inst)

@app.route('/', method=['GET', 'POST', 'PATCH', 'DELETE'])
def handler():
    req = bottle.request
    content_type = req.get_header('content-type')
    data = req.body.read()
    if content_type == 'application/json':
        data = req.json
    event = {
        'data': data,
        'event-id': req.get_header('event-id'),
        'event-type': req.get_header('event-type'),
        'event-time': req.get_header('event-time'),
        'event-namespace': req.get_header('event-namespace'),
        'extensions': {
            'request': req
        }
    }
    method = req.method
    func_calls.labels(method).inc()
    with func_errors.labels(method).count_exceptions():
        with func_hist.labels(method).time():
            q = Queue()
            p = Process(target=funcWrap, args=(q, event, function_context))
            p.start()
            p.join(timeout)
            # If thread is still active
            if p.is_alive():
                p.terminate()
                p.join()
                return bottle.HTTPError(408, "Timeout while processing the function")
            else:
                res = q.get()
                if isinstance(res, Exception):
                    raise res
                return res

@app.get('/healthz')
def healthz():
    return 'OK'

@app.get('/metrics')
def metrics():
    bottle.response.content_type = prom.CONTENT_TYPE_LATEST
    return prom.generate_latest(prom.REGISTRY)

if __name__ == '__main__':
    import logging
    import sys
    import requestlogger
    loggedapp = requestlogger.WSGILogger(
        app,
        [logging.StreamHandler(stream=sys.stdout)],
        requestlogger.ApacheFormatter())
    bottle.run(loggedapp, server='cherrypy', host='0.0.0.0', port=func_port)
