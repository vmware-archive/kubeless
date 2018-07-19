FROM bitnami/minideb-runtimes:jessie

# Install required system packages and dependencies
RUN install_packages build-essential ca-certificates curl git libbz2-1.0 libc6 libncurses5 libreadline6 libsqlite3-0 libssl1.0.0 libtinfo5 pkg-config unzip wget zlib1g
RUN wget -nc -P /tmp/bitnami/pkg/cache/ https://downloads.bitnami.com/files/stacksmith/python-3.6.3-0-linux-x64-debian-8.tar.gz && \
    echo "efbf832408cf62b6a2fb4c44010252cfe374528f22bc2a7d2b6240512a77322b  /tmp/bitnami/pkg/cache/python-3.6.3-0-linux-x64-debian-8.tar.gz" | sha256sum -c - && \
    tar -zxf /tmp/bitnami/pkg/cache/python-3.6.3-0-linux-x64-debian-8.tar.gz -P --transform 's|^[^/]*/files|/opt/bitnami|' --wildcards '*/files' && \
    rm -rf /tmp/bitnami/pkg/cache/python-3.6.3-0-linux-x64-debian-8.tar.gz

ENV BITNAMI_APP_NAME="python" \
    BITNAMI_IMAGE_VERSION="3.6.3-r0" \
    PATH="/opt/bitnami/python/bin:$PATH"

RUN curl https://bootstrap.pypa.io/get-pip.py --output get-pip.py
RUN python ./get-pip.py

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client

WORKDIR /
ADD kubeless.py .

USER 1000

CMD ["python", "/kubeless.py"]
