FROM bitnami/minideb:jessie

ADD http-controller /http-controller

RUN install_packages ca-certificates

ENTRYPOINT ["/http-controller"]
