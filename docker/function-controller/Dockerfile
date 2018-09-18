FROM bitnami/minideb:jessie

RUN install_packages ca-certificates

ADD kubeless-function-controller /kubeless-function-controller

ENTRYPOINT ["/kubeless-function-controller"]
