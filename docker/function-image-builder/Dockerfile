FROM debian:bullseye

RUN apt-get update
RUN apt-get -y install skopeo nodejs

ADD imbuilder /
ADD entrypoint.sh /

ENTRYPOINT [ "/entrypoint.sh" ]
