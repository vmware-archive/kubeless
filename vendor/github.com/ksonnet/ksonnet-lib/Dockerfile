# Builds a Docker image allows you to run Jsonnet and ksonnet on a
# file in your local directory. Specifically, this image contains:
#
# 1. Jsonnet, added to /usr/local/bin
# 2. ksonnet-lib, added to the Jsonnet library paths, so you can
#    compile against the ksonnet libraries without specifying the -J
#    flag.
#
# USAGE: Define a function like `ksonnet` below, and then run:
#
#   `ksonnet <jsonnet-file-and-options-here>`
#
# ksonnet() {
#   docker run -it --rm   \
#     --volume "$PWD":/wd \
#     --workdir /wd       \
#     ksonnet             \
#     jsonnet "$@"
# }

FROM alpine:3.6

# Get Jsonnet v0.9.4.
RUN apk update && apk add git make g++
RUN git clone https://github.com/google/jsonnet.git
RUN cd jsonnet && git checkout tags/v0.9.4 -b v0.9.4 && make -j4 && mv jsonnet /usr/local/bin

# Get ksonnet-lib, add to the Jsonnet -J path.
RUN git clone https://github.com/ksonnet/ksonnet-lib.git
RUN mkdir -p /usr/share/v0.9.4
RUN cp -r ksonnet-lib/ksonnet.beta.2 /usr/share/v0.9.4
