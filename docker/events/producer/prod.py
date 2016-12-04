#!/usr/bin/env python

import json
from kafka import KafkaProducer

producer=KafkaProducer(bootstrap_servers='10.0.0.110:9092',value_serializer=lambda v: json.dumps(v).encode('utf-8'))
producer.send('foobar',{'hello':'world'})

