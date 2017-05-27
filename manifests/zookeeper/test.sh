kubectl exec -n kubeless  zoo-0 -- /opt/bitnami/zookeeper/bin/zkCli.sh create /bitnami foobar;
kubectl exec  -n kubeless zoo-2 -- /opt/bitnami/zookeeper/bin/zkCli.sh get /bitnami;
