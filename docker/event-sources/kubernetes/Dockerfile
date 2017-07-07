FROM bitnami/minideb:jessie

RUN install_packages python3 curl ca-certificates git
RUN curl https://bootstrap.pypa.io/get-pip.py --output get-pip.py
RUN python3 ./get-pip.py
RUN pip3 install --upgrade kubernetes
RUN pip3 install --upgrade requests

RUN git clone https://github.com/dpkp/kafka-python
WORKDIR kafka-python
RUN python3 ./setup.py install

WORKDIR /
ADD events.py .

CMD ["python3", "/events.py"]
