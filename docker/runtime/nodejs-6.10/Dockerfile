FROM node:6.10-alpine

ARG TARGET=http-trigger
RUN apk add --no-cache git

ADD lib/helper.js /lib/
ADD ${TARGET}/kubeless.js /
ADD ${TARGET}/package.json /

RUN npm install

CMD ["node", "/kubeless.js"]
