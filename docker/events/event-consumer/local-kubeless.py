#!/usr/bin/env python

import sys
import os
import imp
import json

from kafka import KafkaConsumer

func_handler = 'handler'
mod_path = 'hello.py'

try:
    mod = imp.load_source('function', mod_path)
except ImportError:
    print("No valid module found for the name: function, Failed to import module")

consumer=KafkaConsumer(bootstrap_servers='10.0.0.110:9092',value_deserializer=json.dumps)
consumer.subscribe(['kubeless'])
while True:
    for msg in consumer:
	    getattr(mod, func_handler)(json.loads(msg.value))
