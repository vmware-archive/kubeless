FROM julia-python3

RUN mkdir -p /kubeless
COPY hello.jl /kubeless

CMD ["python", "/kubeless-julia.py"]


#RUN apt-get update
#RUN apt-get -y install python3
#RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client julia=0.1.5
#
#WORKDIR /
#ADD kubeless-julia.py .
#RUN mkdir -p /kubeless
#COPY test.jl /kubeless
#USER 1000
#
#CMD ["python", "/kubeless-julia.py"]
