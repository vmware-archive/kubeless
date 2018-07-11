FROM ballerina/ballerina:0.980.0

USER 1000

WORKDIR /kubeless

ENV FUNC_PROCESS="ballerina run kubeless_run.balx -c kubeless.toml"

ADD proxy /

CMD [ "/proxy" ]