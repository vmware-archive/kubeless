FROM bitnami/minideb:jessie 


RUN apt-get update

RUN install_packages python-pip

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client

ADD startup.sh /
ADD kubeless-shell.py /

EXPOSE 8080

ENTRYPOINT ["/startup.sh"]
