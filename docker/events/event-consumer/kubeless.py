#!/usr/bin/env python

import sys
import os
import imp
import json

from kafka import KafkaConsumer

mod_name = os.getenv('MOD_NAME')
func_handler = os.getenv('FUNC_HANDLER')

mod_path = '/kubeless/' + mod_name + '.py'

try:
    mod = imp.load_source('lambda', mod_path)
except ImportError:
    print("No valid module found for the name: lambda, Failed to import module")

consumer=KafkaConsumer(bootstrap_servers='kafka.kubeless:9092',value_deserializer=json.dumps)
consumer.subscribe(['kubeless'])
while True:
    for msg in consumer:
	    getattr(mod, func_handler)(json.loads(msg.value))
