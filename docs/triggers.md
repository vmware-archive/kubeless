# Triggers

To invoke deployed functions, you need to create **triggers**. A function can have multiple triggers, but each of those will only reference a single deployed function.

Each trigger has its own schema and usage, so we've created a separate page for each one of those.

## Available triggers

In this section, we're going to list our triggers. Since Kubeless is an open-source tool there are multiple triggers that we haven't listed here. Feel free to add your trigger to this list.

* [HTTP Trigger](/docs/http-triggers)
* [CronJob Trigger](/docs/cronjob-triggers)
* [PubSub Triggers](/docs/pubsub-functions)
  * [Kafka Trigger](/docs/pubsub-functions#kafka)
  * [NATS Trigger](/docs/pubsub-functions#nats)
* [Data Stream Triggers](/docs/streaming-functions)
  * [AWS Kinesis Trigger](/docs/streaming-functions/#aws-kinesis)
 

## Creating a new trigger

It is really simple to create a new trigger on Kubeless. Take a look at the [Implementing a New Trigger](/docs/implementing-new-trigger) page to learn more about it.
