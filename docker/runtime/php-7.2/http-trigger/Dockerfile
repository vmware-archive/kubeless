FROM php:7.2-stretch

# Install composer
ENV COMPOSER_HOME /composer
ENV PATH="/root/.composer/vendor/bin:${PATH}"
ENV COMPOSER_ALLOW_SUPERUSER=1
ENV FUNC_PORT 8080

RUN apt-get update \
    && apt-get install -y git unzip libxml2-dev libpng-dev libjpeg-dev libssl-dev libicu-dev libcurl4-openssl-dev libmcrypt-dev postgresql-server-dev-all \
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
ENTRYPOINT [""]
CMD ["sh", "-c", "exec php -S 0.0.0.0:${FUNC_PORT} Controller.php"]
