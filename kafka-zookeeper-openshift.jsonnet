local k = import "ksonnet.beta.1/k.libsonnet";

local manifest = import "kafka-zookeeper.jsonnet";

manifest + {
  kafkaSts+:
  {spec+: {template+: {spec+: {initContainers: []}}}},
  zookeeperSts+:
  {spec+: {template+: {spec+: {initContainers: []}}}},  
}