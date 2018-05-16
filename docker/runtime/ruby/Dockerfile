FROM bitnami/ruby:2.4

LABEL maintainer "Bitnami <containers@bitnami.com>"

ENV RACK_ENV="production"
ENV FUNC_PROCESS="ruby /kubeless.rb"
RUN gem install sinatra --no-rdoc --no-ri

ADD kubeless.rb /

USER 1000

ADD proxy /
CMD [ "/proxy" ]
