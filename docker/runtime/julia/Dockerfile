FROM debian

RUN apt-get update \
 && apt-get install -y python3 python3-dev python3-pip python3-setuptools wget git bzip2 tar curl

RUN mkdir -p /usr/local/julia
RUN wget https://julialang-s3.julialang.org/bin/linux/x64/0.6/julia-0.6.2-linux-x86_64.tar.gz
RUN tar -xzvf julia-0.6.2-linux-x86_64.tar.gz -C /usr/local/
RUN mv /usr/local/julia-d386e40c17 /usr/local/julia/julia
RUN rm julia-0.6.2-linux-x86_64.tar.gz
ENV PATH /usr/local/julia/julia/bin:$PATH
RUN pip3 install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger prometheus_client
RUN pip3 install julia==0.1.5

RUN useradd julia -u 1000
RUN usermod -g root julia
RUN chown julia /home/
USER julia
RUN julia -e 'Pkg.add("PyCall")'

WORKDIR /
RUN chown julia /home/
ADD kubeless-julia.py /home/julia/

CMD ["python3", "/home/julia/kubeless-julia.py"]
