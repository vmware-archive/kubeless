FROM bitnami/minideb:jessie 

RUN apt-get update

ADD install_dependencies.sh /bin/
ADD requirements /tmp
RUN /bin/install_dependencies.sh


RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client


ADD kubeless-shell.py /
EXPOSE 8080
ENTRYPOINT ["python", "/kubeless-shell.py"]
