FROM python:2.7-slim

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client

ADD kubeless.py /

EXPOSE 8080
CMD ["python", "/kubeless.py"]
