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
group = mod_name + func_handler

try:
    mod = imp.load_source('function', mod_path)
except ImportError:
    print("No valid module found for the name: function, Failed to import module")

def json_safe_loads(msg):
    try:
        data = json.loads(msg)
        return {'type':'json','payload':data}
    except:
        return {'type':'text','payload':msg}

consumer=KafkaConsumer(bootstrap_servers='kafka.kubeless:9092', group_id=group, value_deserializer=json_safe_loads)
consumer.subscribe([topic_name])
while True:
    for msg in consumer:
        res = getattr(mod, func_handler)(msg.value['payload'])
        print res
