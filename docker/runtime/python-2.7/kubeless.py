#!/usr/bin/env python

import os
import imp

import bottle

mod = imp.load_source('function',
                      '/kubeless/%s.py' % os.getenv('MOD_NAME'))
func = getattr(mod, os.getenv('FUNC_HANDLER'))

app = application = bottle.app()

@app.get('/')
def handler():
    return func()

@app.post('/')
def post_handler():
    return func(bottle.request)

@app.get('/healthz')
def healthz():
    return 'OK'

if __name__ == '__main__':
    import logging
    import sys
    import requestlogger
    loggedapp = requestlogger.WSGILogger(
        app,
        [logging.StreamHandler(stream=sys.stdout)],
        requestlogger.ApacheFormatter())
    bottle.run(loggedapp, server='cherrypy', host='0.0.0.0', port=8080)
