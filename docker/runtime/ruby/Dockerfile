FROM bitnami/ruby:2.4

LABEL maintainer "Bitnami <containers@bitnami.com>"

ENV RACK_ENV="production"

RUN gem install sinatra --no-rdoc --no-ri

ADD kubeless.rb /

USER 1000

CMD ["ruby", "/kubeless.rb"]
