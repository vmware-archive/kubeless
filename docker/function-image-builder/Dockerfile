FROM fedora:27

RUN dnf install -y skopeo nodejs

ADD imbuilder /
ADD entrypoint.sh /

ENTRYPOINT [ "/entrypoint.sh" ]
