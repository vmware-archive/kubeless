FROM bitnami/minideb:jessie

RUN install_packages ca-certificates

ADD kubeless-controller-manager /kubeless-controller-manager

ENTRYPOINT ["/kubeless-controller-manager"]
