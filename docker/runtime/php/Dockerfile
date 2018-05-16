FROM php:7.2-fpm-stretch

# Install composer
ENV COMPOSER_HOME /composer
ENV PATH="/root/.composer/vendor/bin:${PATH}"
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV FUNC_PROCESS="/usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf"

RUN apt-get update \
    && apt-get install -y \
       git unzip libxml2-dev libpng-dev libjpeg-dev libssl-dev libicu-dev \
       libcurl4-openssl-dev libmcrypt-dev postgresql-server-dev-all nginx supervisor \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* \
    && curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

RUN docker-php-ext-install \
    opcache \
    bcmath \
    ctype \
    curl \
    dom \
    iconv \
    gd \
    gettext \
    intl \
    json \
    mysqli \
    pgsql \
    pcntl \
    pdo \
    ftp \
    pdo_mysql \
    pdo_pgsql \
    phar \
    simplexml \
    xmlrpc \
    zip

WORKDIR /app
COPY . /app

RUN composer install --prefer-dist --optimize-autoloader
COPY config/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY config/default.conf /etc/nginx/conf.d/default.conf

# Prepare service to run as non root
RUN rm /etc/nginx/sites-available/default && \
    mkdir -p /app && \
    touch /run/nginx.pid && \
    chmod -R a+w /run/nginx.pid /app /var/lib/nginx /var/log/nginx
USER 1000

ADD proxy /
CMD [ "/proxy" ]
