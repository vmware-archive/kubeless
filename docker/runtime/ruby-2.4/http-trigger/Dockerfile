FROM bitnami/ruby:2.4

RUN install_packages git

RUN gem install sinatra

ADD kubeless.rb /

CMD ["ruby", "/kubeless.rb"]
