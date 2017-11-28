FROM docker:17.11.0-ce-dind

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:/usr/local/bats/bin:$PATH \
    CGO_ENABLED=0

# Install packages that requires persistence
RUN set -eux; \
    apk add --no-cache \
    bash \
    git \
    make \
    sudo \
    gcc \
    musl-dev \
    openssl \
    ca-certificates \
    zip \
    curl \
    go && \
    # Install kubectl
    KUBECTL_VERSION=$(wget -qO- https://storage.googleapis.com/kubernetes-release/release/stable.txt) && \
    wget "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl" -O "/usr/local/bin/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    # Install gox and golint
    go get github.com/mitchellh/gox github.com/golang/lint/golint && \
    # Install bats
    git clone https://github.com/sstephenson/bats /usr/local/bats && \
    # Install kubecfg
    wget "https://github.com/ksonnet/kubecfg/releases/download/v0.5.0/kubecfg-linux-amd64" -O "/usr/local/bin/kubecfg" && chmod +x "/usr/local/bin/kubecfg"

WORKDIR $GOPATH

ADD ./entry-point.sh /
ENTRYPOINT [ "/entry-point.sh" ]
