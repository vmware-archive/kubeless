# OCR Pipeline
As a application developer, I want to create an OCR pipelline so that when I drop a file in Minio objet store, it gets automatically parsed by Apache Tika  and the text gets stored in an ElasticSearch index, using serverless functions provided by Kubeless.

## TL;DR;

```console
$ helm install stable/ocr-pipeline
```

## Introduction

This chart bootstraps an [OCR](https://en.wikipedia.org/wiki/Optical_character_recognition) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

It makes use of [Minio](https://github.com/kubernetes/charts/tree/master/stable/minio), [Apache Tika](https://github.com/bitnami/charts/tree/tika-server/incubator/tika-server) and [MongoDB](https://github.com/kubernetes/charts/tree/master/stable/mongodb) as Object Store, Translator and Document Store.

## Prerequisites

- Kubernetes 1.4+ with Beta APIs enabled
- PV provisioner support in the underlying infrastructure

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release stable/ocr-pipeline
```

The command deploys WordPress on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the WordPress chart and their default values.

| Parameter                            | Description                                | Default                                                    |
| -------------------------------      | -------------------------------            | ---------------------------------------------------------- |
| `image`                              | WordPress image                            | `bitnami/wordpress:{VERSION}`                              |
| `imagePullPolicy`                    | Image pull policy                          | `IfNotPresent`                                             |
| `allowEmptyPassword`                 | Allow DB blank passwords                   | `yes`                                          |
| `serviceType`                        | Kubernetes Service type                    | `LoadBalancer`                                             |
| `ingress.enabled`                    | Enable ingress controller resource         | `false`                                                    |
| `ingress.hostname`                   | URL to address your WordPress installation | `wordpress.local`                                          |
| `ingress.tls`                        | Ingress TLS configuration                  | `[]`                                          |
| `persistence.enabled`                | Enable persistence using PVC               | `true`                                                     |
| `persistence.storageClass`           | PVC Storage Class                          | `nil` (uses alpha storage class annotation)                |
| `persistence.accessMode`             | PVC Access Mode                            | `ReadWriteOnce`                                            |
| `persistence.size`                   | PVC Storage Request                        | `10Gi`                                                      |                                              |

The above parameters map to the env variables defined in [bitnami/wordpress](http://github.com/bitnami/bitnami-docker-wordpress). For more information please refer to the [bitnami/wordpress](http://github.com/bitnami/bitnami-docker-wordpress) image documentation.

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install --name my-release \
  --set wordpressUsername=admin,wordpressPassword=password,mariadb.mariadbRootPassword=secretpassword \
    stable/wordpress
```

The above command sets the WordPress administrator account username and password to `admin` and `password` respectively. Additionally it sets the MariaDB `root` user password to `secretpassword`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install --name my-release -f values.yaml stable/wordpress
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Persistence

The [Bitnami WordPress](https://github.com/bitnami/bitnami-docker-wordpress) image stores the WordPress data and configurations at the `/bitnami` path of the container.

Persistent Volume Claims are used to keep the data across deployments. This is known to work in GCE, AWS, and minikube.
See the [Configuration](#configuration) section to configure the PVC or to disable persistence.

## Ingress

This chart provides support for Ingress resource. If you have available an Ingress Controller such as Nginx or Traefik you maybe want to set up `ingress.enabled` to true and choose a `ingress.hostname` for the URL. Then, you should be able to access the installation using that address.
