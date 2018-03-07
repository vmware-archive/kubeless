FROM bitnami/minideb:jessie

LABEL maintainer="Nguyen Anh Tu <tuna@bitnami.com>"

RUN install_packages ca-certificates

ADD kubeless-controller /kubeless-controller

ENTRYPOINT ["/kubeless-controller"]
