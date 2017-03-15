#!/usr/bin/env python

import sys
import os
import imp
import json

from kafka import KafkaConsumer

mod_name = os.getenv('MOD_NAME')
func_handler = os.getenv('FUNC_HANDLER')
topic_name = os.getenv('TOPIC_NAME')

mod_path = '/kubeless/' + mod_name + '.py'

try:
    mod = imp.load_source('lambda', mod_path)
except ImportError:
    print("No valid module found for the name: lambda, Failed to import module")

consumer=KafkaConsumer(bootstrap_servers='kafka.kubeless:9092', value_deserializer=json.loads)
consumer.subscribe([topic_name])
while True:
    for msg in consumer:
        res = getattr(mod, func_handler)(msg.value)
        print res
        print type(res)
