FROM python:2.7.11-alpine

RUN apk add --no-cache python py-pip
RUN pip install --upgrade pip
RUN pip install kafka-python prometheus_client

ADD kubeless.py .

CMD ["python", "/kubeless.py"]
EXPOSE 8080
