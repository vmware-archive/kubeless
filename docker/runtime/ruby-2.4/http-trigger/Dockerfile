FROM bitnami/ruby:2.4

LABEL maintainer "Bitnami <containers@bitnami.com>"

ENV RACK_ENV="production"

RUN install_packages git

RUN gem install sinatra --no-rdoc --no-ri 

ADD kubeless.rb /

CMD ["ruby", "/kubeless.rb"]
