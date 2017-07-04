FROM python:2.7.11-alpine

RUN apk add --no-cache python py-pip
RUN pip install -U pip==9.0.1
RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client

ADD kubeless.py /

CMD ["python", "/kubeless.py"]
EXPOSE 8080
