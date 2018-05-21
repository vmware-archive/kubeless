FROM ubuntu
RUN apt install julia
RUN julia -e 'Pkg.add("PyCall")'
RUN apt install python3
RUN apt install python3-pip
RUN pip3 install julia
RUN pip3 install bottle
RUN pip install prometheus_client




