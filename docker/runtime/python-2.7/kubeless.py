#!/usr/bin/env python

import sys
import os
import imp

from bottle import route, run, request

mod_name = os.getenv('MOD_NAME')
func_handler = os.getenv('FUNC_HANDLER')

mod_path = '/kubeless/' + mod_name + '.py'

try:
    mod = imp.load_source('lambda', mod_path)
except ImportError:
    print("No valid module found for the name: lambda, Failed to import module")

@route('/', method="GET")
def handler():
    return getattr(mod, func_handler)()

@route('/', method="POST")
def post_handler():
    return getattr(mod, func_handler)(request)

@route('/healthz', method="GET")
def healthz():
    return "OK"

run(host='0.0.0.0', port=8080)
