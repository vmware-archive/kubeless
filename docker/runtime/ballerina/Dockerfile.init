FROM ballerina/ballerina-platform:0.981.0

USER 1000

WORKDIR /kubeless

# Install controller
ADD kubeless_run.tpl.bal /ballerina/files/src/
ADD compile-function.sh /compile-function.sh

ENTRYPOINT [""]
