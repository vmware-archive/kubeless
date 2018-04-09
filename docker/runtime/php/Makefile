

build:
	docker build -t kubeless/php:7.2$$RUNTIME_TAG_MODIFIER .

push:
	docker push kubeless/php:7.2$$RUNTIME_TAG_MODIFIER

php-cs-fixer:
	docker run -it --rm -v $${PWD}:/wd -w /wd shouldbee/php-cs-fixer fix Controller.php || true

# Mandatory jobs
build-all: build
push-all: push