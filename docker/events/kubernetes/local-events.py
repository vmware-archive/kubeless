import asyncio
import logging
import json

from kubernetes import client, config, watch

logger = logging.getLogger('k8s_events')
logger.setLevel(logging.DEBUG)

ch = logging.StreamHandler()
ch.setLevel(logging.DEBUG)

formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
ch.setFormatter(formatter)
logger.addHandler(ch)

config.load_kube_config()
#config.load_incluster_config()

v1 = client.CoreV1Api()
v1ext = client.ExtensionsV1beta1Api()

async def pods():
    w = watch.Watch()
    for event in w.stream(v1.list_pod_for_all_namespaces):
        logger.info("Event: %s %s %s" % (event['type'], event['object'].kind, event['object'].metadata.name))
        await asyncio.sleep(0) 

async def namespaces():
    w = watch.Watch()
    for event in w.stream(v1.list_namespace):
        logger.info("Event: %s %s %s" % (event['type'], event['object'].kind, event['object'].metadata.name))
        await asyncio.sleep(0)
        
async def services():
    w = watch.Watch()
    for event in w.stream(v1.list_service_for_all_namespaces):
        logger.info("Event: %s %s %s" % (event['type'], event['object'].kind, event['object'].metadata.name))
        await asyncio.sleep(0)
        
async def deployments():
    w = watch.Watch()
    for event in w.stream(v1ext.list_deployment_for_all_namespaces):
        logger.info("Event: %s %s %s" % (event['type'], event['object'].kind, event['object'].metadata.name))
        await asyncio.sleep(0)
    
async def replicasets():
    w = watch.Watch()
    for event in w.stream(v1ext.list_replica_set_for_all_namespaces):
        logger.info("Event: %s %s %s" % (event['type'], event['object'].kind, event['object'].metadata.name))
        await asyncio.sleep(0)

ioloop = asyncio.get_event_loop()

ioloop.create_task(pods())
ioloop.create_task(namespaces())
ioloop.create_task(services())
ioloop.create_task(deployments())
ioloop.create_task(replicasets())

try:
    # Blocking call interrupted by loop.stop()
    print('step: loop.run_forever()')
    ioloop.run_forever()
except KeyboardInterrupt:
    pass
finally:
    print('step: loop.close()')
    ioloop.close()
