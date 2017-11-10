#!/usr/bin/env python

from subprocess import Popen, PIPE, STDOUT
import bottle
import subprocess
import os
import shlex
import prometheus_client as prom

app = application = bottle.app()

def func(osparam):
    cmdline = '/bin/bash -c "source /kubeless/%s.sh && %s %s "' % (os.getenv('MOD_NAME'), os.getenv('FUNC_HANDLER'), osparam if osparam else '')
    args = shlex.split(cmdline)
    proc = Popen(args, stdout=PIPE, stderr=STDOUT, shell=False)
    out = proc.communicate()[0]
    return out.decode('utf-8')

func_calls = prom.Counter('function_calls_total',
                           'Number of calls to user function',
                          ['method'])
func_errors = prom.Counter('function_failures_total',
                           'Number of exceptions in user function',
                           ['method'])
func_hist = prom.Histogram('function_duration_seconds',
                           'Duration of user function in seconds',
                           ['method'])


@app.route('/', method=['GET', 'POST'])
def handler():
    req = bottle.request
    method = req.method
    func_calls.labels(method).inc()
    with func_errors.labels(method).count_exceptions():
        with func_hist.labels(method).time():
            if method == 'GET':
                return func(None)
            else:
                return func(bottle.request.body.read())


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
    bottle.run(loggedapp, server='cherrypy', host='0.0.0.0', port=8080)
   
