FROM node:6.10-alpine

RUN apk add --no-cache git

ADD kubeless.js /

ADD package.json /

RUN npm install

CMD ["node", "/kubeless.js"]
