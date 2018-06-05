FROM node:8 AS build-env

RUN apt-get update &&  apt-get install git

ADD lib/helper.js kubeless_rt/lib/
ADD kubeless.js kubeless_rt/
ADD package.json kubeless_rt/

WORKDIR kubeless_rt/

RUN npm install


FROM gcr.io/distroless/nodejs@sha256:91b207c7278667472dcd08d8e137ed98c99a3b92120f6a7ec977fc3f63323848
COPY --from=build-env /kubeless_rt /kubeless_rt

WORKDIR kubeless_rt/

USER 1000

CMD ["kubeless.js"]
