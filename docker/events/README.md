Kafka event processor
======================

Launch an RC and services to run zookeeper + kafka.

```
kubectl create -f kafka-cube.yml
```

You can now run a consumer

```
kubectl run consumer --image=skippbox/kubeless-consumer:0.0.2
```

It will just print the content of the msg of a json based message.

Run a producer of events so that something actually happens:

```
kubectl run producer --image=skippbox/kubeless-producer:0.0.2
```

The producer just sends a json message forever:

```
import json
from kafka import KafkaProducer

producer=KafkaProducer(bootstrap_servers='kafka:9092',value_serializer=lambda v: json.dumps(v).encode('utf-8'))
while True:
    producer.send('foobar',{'hello':'world'})
```

Now check the logs of the consumer with `kubectl`


Python based
------------

Both producer and consumer are based on the [Kafka Python client](https://github.com/dpkp/kafka-python)


