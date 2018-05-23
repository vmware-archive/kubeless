FROM bitnami/minideb:jessie

ADD kinesis-controller /kinesis-controller

RUN install_packages ca-certificates

ENTRYPOINT ["/kinesis-controller"]
