FROM ballerina/ballerina-platform:0.970.1

USER 1000

WORKDIR /kubeless

ENV FUNC_PROCESS="ballerina run kubeless_run.balx -c kubeless.toml"

ADD proxy /

CMD [ "/proxy" ]
