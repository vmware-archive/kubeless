FROM bitnami/ruby:2.4

LABEL maintainer "Bitnami <containers@bitnami.com>"

ENV RACK_ENV="production"

RUN gem install sinatra ruby-kafka --no-rdoc --no-ri 

ADD kubeless.rb /

CMD ["ruby", "/kubeless.rb"]
