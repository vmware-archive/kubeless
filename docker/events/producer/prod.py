#!/usr/bin/env python

import json
from kafka import KafkaProducer

producer=KafkaProducer(bootstrap_servers='kafka:9092',value_serializer=lambda v: json.dumps(v).encode('utf-8'))
while True:
    producer.send('foobar',{'hello':'world'})

