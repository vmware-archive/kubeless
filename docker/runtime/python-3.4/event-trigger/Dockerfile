FROM bitnami/minideb:jessie

RUN install_packages python3 curl ca-certificates git
RUN curl https://bootstrap.pypa.io/get-pip.py --output get-pip.py
RUN python3 ./get-pip.py

RUN pip install kafka-python prometheus_client

WORKDIR /
ADD kubeless.py .

EXPOSE 8080
CMD ["python3", "/kubeless.py"]
