# Use an existing Kafka cluster with Kubeless

In Kubeless [release page](https://github.com/kubeless/kubeless/releases), we provide along with Kubeless manifests a collection of Kafka and Zookeeper statefulsets which helps user to quickly deploying PubSub function. These statefulsets are deployed in `kubeless` namespace. However, if you have a Kafka cluster already running in the same Kubernetes cluster, this doc will walk you through how to deploy Kubeless PubSub function with it.

Let's assume that you have Kafka cluster running at `pubsub` namespace like below:

```console
$ kubectl -n pubsub get po
NAME      READY     STATUS    RESTARTS   AGE
kafka-0   1/1       Running   0          7h
zoo-0     1/1       Running   0          7h

$ kubectl -n pubsub get svc
NAME        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
kafka       ClusterIP   10.55.253.151   <none>        9092/TCP            7h
zookeeper   ClusterIP   10.55.248.146   <none>        2181/TCP            7h
```

**Note**: If you want to use the command `kubeless topic` you need add a label to your Kafka deployment (`kubeless=kafka`) in order for the CLI to find it. 

And Kubeless already running at `kubeless` namespace:

```console
$ kubectl -n kubeless get po
NAME                                           READY     STATUS    RESTARTS   AGE
kubeless-controller-manager-58676964bb-l79gh   1/1       Running   0          5d
```

Now we need to deploy the Kafka consumer and the Kafka Trigger CRD. We can do that extracting the Deployment, CRD and ClusterRoles from the generic Kafka manifest. The key part is adding the environment variable `KAFKA_BROKERS` pointing to the right URL:

```yaml
$ echo '
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    kubeless: kafka-trigger-controller
  name: kafka-trigger-controller
  namespace: kubeless
spec:
  selector:
    matchLabels:
      kubeless: kafka-trigger-controller
  template:
    metadata:
      labels:
        kubeless: kafka-trigger-controller
    spec:
      containers:
      - image: bitnami/kafka-trigger-controller:latest
        imagePullPolicy: IfNotPresent
        name: kafka-trigger-controller
        env:
        - name: KAFKA_BROKERS
          value: kafka.pubsub:9092 # CHANGE THIS!
      serviceAccountName: controller-acct
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: kafkatriggers.kubeless.io
spec:
  group: kubeless.io
  names:
    kind: KafkaTrigger
    plural: kafkatriggers
    singular: kafkatrigger
  scope: Namespaced
  version: v1beta1
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: kafka-controller-deployer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kafka-controller-deployer
subjects:
- kind: ServiceAccount
  name: controller-acct
  namespace: kubeless
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: kafka-controller-deployer
rules:
- apiGroups:
  - ""
  resources:
  - services
  - configmaps
  verbs:
  - get
  - list
- apiGroups:
  - kubeless.io
  resources:
  - functions
  - kafkatriggers
  verbs:
  - get
  - list
  - watch
  - update
  - delete
' | kubectl create -f -
deployment "kafka-trigger-controller" created
clusterrolebinding "kafka-controller-deployer" created
clusterrole "kafka-controller-deployer" created
customresourcedefinition "kafkatriggers.kubeless.io" created
```

Now we need to create `s3-python` topic and try to publish some messages. You can do it on your own kafka client. In this example, I will try to use the bundled binaries in the kafka container:

```console
# create s3-python topic
$ kubectl -n pubsub exec -it kafka-0 -- /opt/bitnami/kafka/bin/kafka-topics.sh --create --zookeeper zookeeper.pubsub:2181 --replication-factor 1 --partitions 1 --topic s3-python

# send test message to s3-python topic
$ kubectl -n pubsub exec -it kafka-0 -- /opt/bitnami/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic s3-python
> hello world
```

Open another terminal and check for the pubsub function log to see if it receives the message:

```console
$ kubectl logs -f pubsub-python-5445bdcb64-48bv2
hello world
```

When using SASL you must add `KAFKA_ENABLE_SASL`, `KAFKA_USERNAME` and `KAFKA_PASSWORD` env var to set authentification (might use a secret).:

```yaml
$ echo '
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    kubeless: kafka-trigger-controller
  name: kafka-trigger-controller
  namespace: kubeless
spec:
  selector:
    matchLabels:
      kubeless: kafka-trigger-controller
  template:
    metadata:
      labels:
        kubeless: kafka-trigger-controller
    spec:
      containers:
      - image: bitnami/kafka-trigger-controller:latest
        imagePullPolicy: IfNotPresent
        name: kafka-trigger-controller
        env:
        ...
        - name: KAFKA_ENABLE_SASL
          value: true # CHANGE THIS!
        - name: KAFKA_USERNAME
          value: kafka-sasl-username # CHANGE THIS!
        - name: KAFKA_PASSWORD
          value: kafka-sasl-password # CHANGE THIS!
...
```

When using SSL to secure kafka communication, you must set `KAFKA_ENABLE_TLS`, and specify some of these: 
* `KAFKA_CACERTS` to check server certificate
* `KAFKA_CERT` and `KAFKA_KEY` to check client certificate
* `KAFKA_INSECURE` to skip TLS verfication
