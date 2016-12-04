#!/usr/bin/env python

import json
from kafka import KafkaConsumer

consumer=KafkaConsumer(bootstrap_servers='kafka.kubeless:9092',value_deserializer=json.dumps)
consumer.subscribe(['foobar'])
while True:
    for msg in consumer:
        print type(msg)
        print type(msg.value)
        print json.loads(msg.value)
